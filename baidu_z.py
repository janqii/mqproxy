# -*- coding: utf-8 -*-
import sys
import os
import time
import cPickle as pickle
from kazoo.client import KazooClient
from kazoo.exceptions import *
from kazoo.security import ACL
from kazoo.security import Id
from kazoo.handlers.threading import TimeoutError

DEF_HOST = '10.46.135.28:2181'
#DEF_HOST = '10.23.253.43:8181'
HOST_MAP = {
    'online' : 'zkservers.sys.baidu.com:8181',
    'ksarch' : '10.46.135.28:2181',
    'hz' : 'hz.nszk.baidu.com:8181'
}
TIMEOUT = 0.3
VAR_PATH = os.getenv('Z_VAR_PATH', os.path.dirname(os.path.realpath(__file__)) + '/var')

import logging
logging.getLogger().setLevel(100)
#logging.basicConfig()

def error_print(*msg):
    sys.stderr.write('ERROR: ')
    sys.stderr.write(' '.join(msg))
    sys.stderr.write("\n")

class ZError(BaseException):
    def __init__(self, msg):
        self.msg = msg
    def get(self):
        return self.msg

class Z:
    #utils
    def pretty_print_dict(self, d, indent = 0):
        maxlen = len(max(d.keys(), key=len))
        for k in sorted(d.keys()):
            print ' ' * indent + k, ' '*(maxlen - len(k)) + ':', d[k]

    def call(self, action, param):
        if not self.action_exist(action):
            raise ZError("cmd '%s' not found" % action)
        f = getattr(self, 'cmd_' + action)
        try:
            return f(*param)
        except NoAuthError:
            raise ZError("Not authorized")

    def action_exist(self, cmd):
        return hasattr(self, 'cmd_' + cmd)

    def memoize(fn):
        tmp = {}
        def wrap(self, *args, **kwargs):
            key = pickle.dumps((args, kwargs))
            if key in tmp:
                return tmp[key]
            ret = fn(self, *args, **kwargs)
            tmp[key] = ret
            return ret
        return wrap

    @memoize
    def get_docs(self, cmd):
        if not hasattr(self, 'cmd_' + cmd):
            return None
        doc = [ x.strip() for x in getattr(self, 'cmd_' + cmd).__doc__.strip().split("\n") if x.strip() ]
        d = {}
        for i in doc:
            if i[0] == '@':
                i = i.split(":", 1)
                key = i[0][1:].strip()
                value = i[1].strip()
                d[key] = value
            elif i[0] == '+':
                i = i.split(":", 1)
                key = i[0][1:].strip()
                value = i[1].strip()
                if key not in d or not isinstance(d[key], list):
                    d[key] = []
                d[key].append(value)
            else:
                d['desc'] = i
        return d

    def get_desc(self, cmd):
        return self.get_docs(cmd).get('desc', None)

    def get_usage(self, cmd):
        return self.get_docs(cmd).get('usage', None)

    def get_options(self, cmd):
        return self.get_docs(cmd).get('option', None)

    def get_long_msg(self, cmd):
        return '\n'.join(self.get_docs(cmd).get('long', ''))

    #decorator
    def ensure_exist(fn):
        def ret(self, *args, **kwargs):
            try:
                return fn(self, *args, **kwargs)
            except NoNodeError as e:
                raise ZError("Node does not exist")
        ret.__doc__ = fn.__doc__
        return ret

    def need_connect(fn):
        def ret(self, *args, **kwargs):
            try:
                self.connect()
                return fn(self, *args, **kwargs)
            finally:
                self.disconnect()
        ret.__doc__ = fn.__doc__
        return ret

    def alias(tgt, defaults = {}):
        def wrap(fn):
            def ret(self, *args, **kwargs):
                defaults.update(kwargs)
                method = getattr(self, 'cmd_' + tgt)
                return method(*args, **defaults)
            ret.__doc__ = "Alias of %s" % tgt
            return ret
        return wrap

    def fix_path(*idx):
        def wrap(fn):
            def ret(self, *oriargs, **kwargs):
                cwd = self.get_user_conf_key('cwd')
                args = list(oriargs)
                for i in idx:
                    if i > len(args):
                        continue
                    if cwd and (len(args[i-1]) == 0 or args[i-1][0] != '/'):
                        args[i-1] = cwd + '/' + args[i-1]
                    path = args[i-1].split('/')
                    newpath = []
                    for p in path:
                        if p == '':
                            continue
                        if p == '.':
                            continue
                        if p == '..':
                            if len(newpath) > 0:
                                newpath.pop()
                            continue
                        newpath.append(p)
                    args[i-1] = '/' + '/'.join(newpath)
                return fn(self, *args, **kwargs)
            ret.__doc__ = fn.__doc__
            return ret
        return wrap

    def option(opt, key, desc = '', default = None, has_arg = False, default_arg = None):
        def wrap(fn):
            def ret(self, *oriargs, **orikwargs):
                args = list(oriargs)
                kwargs = dict(orikwargs)
                try:
                    index = args.index(opt)
                except ValueError:
                    index = False
                if index is not False:
                    args.pop(index)
                    if has_arg:
                        try:
                            value = args[index]
                            args.pop(index)
                        except IndexError:
                            if default_arg is None:
                                raise ZError("Parameter needed for option %s"%opt)
                            value = default_arg
                    else:
                        value = True
                    kwargs[key] = value
                else:
                    kwargs[key] = default
                return fn(self, *args, **kwargs)
                        
            if has_arg:
                if default_arg is None:
                    doc = "%s %s: %s" % (opt, key, desc)
                else:
                    doc = "%s [%s]: %s (default %s)" % (opt, key, desc, default_arg)
            else:
                doc = "%s: %s" % (opt, desc)
            ret.__doc__ = fn.__doc__ + "\n+option:" + doc
            return ret
        return wrap

    #common helper functions
    def connect(self):
        if not hasattr(self, 'zk'):
            try:
                self.zk = KazooClient(self.host)
                self.zk.start(TIMEOUT)
            except TimeoutError:
                raise ZError("Connect to zookeeper timeout")

    def disconnect(self):
        if hasattr(self, 'zk'):
            self.zk.stop()

    def get_version(self, path):
        ret = self.zk.get(path, None)
        return ret[1].version

    # tty is used as user conf file key
    def get_tty(self):
        tty = None
        if os.getenv('Z_USE_KEY', '') != '':
            return os.getenv('Z_USE_KEY')
        if os.getenv('Z_USE_PPID') == '1':
            return 'ppid-' + str(os.getppid())
        for i in [0,1,2]:
            try:
                tty = os.ttyname(i)
                break
            except:
                pass
        return tty

    def get_conf_file(self):
        tty = self.get_tty()
        if tty is None:
            return None
        return VAR_PATH + '/' + str(os.getuid()) + '-' + tty.replace('/', '-')

    @memoize
    def load_user_conf(self):
        tty = self.get_tty()
        if tty is None:
            return {}
        try:
            data = pickle.load(open(self.get_conf_file(), 'r'))
        except IOError:
            return {}
        if time.time() - data['time'] < 3600:
            return data['data']
        return {}

    def save_user_conf(self, data):
        tty = self.get_tty()
        if tty is None:
            return
        if not os.path.isdir(os.path.dirname(self.get_conf_file())):
            os.makedirs(os.path.dirname(self.get_conf_file()))
        with open(self.get_conf_file(), 'w') as f:
            pickle.dump({'time': time.time(), 'data': data}, f)

    def get_user_conf_key(self, key):
        if key in self.load_user_conf():
            return self.load_user_conf()[key]
        else:
            return None

    def __init__(self):
        if os.getenv('Z_HOST', None):
            self.host = os.getenv('Z_HOST')
        elif 'host' in self.load_user_conf():
            self.host = self.load_user_conf()['host']
        else:
            self.host = DEF_HOST
        if self.host in HOST_MAP:
            self.host = HOST_MAP[self.host]

    #commands
    def cmd_credit(self, *unused, **options):
        """Show credits"""
        print "Zookeeper shell"
        print "    by lauchingjun@baidu.com"

    def cmd_help(self, action = None, *unused, **options):
        """
        Show help message
        @usage: [command]
        """
        out = {}
        if action is None:
            print "Usage: %s command [params..]" % os.getenv('ARGV0', sys.argv[0])
            print "Commands available:"
            for name in dir(self):
                if not name.startswith('cmd_'):
                    continue
                doc = getattr(self, name).__doc__
                if doc is not None:
                    out[name[4:]] = self.get_desc(name[4:])
            self.pretty_print_dict(out, 2)
        else:
            if not self.action_exist(action):
                raise ZError("cmd '%s' not found" % action)
            print '%s - %s' % (action, self.get_desc(action))
            if self.get_usage(action):
                print 'Usage: %s %s' % (action, self.get_usage(action))
            if self.get_long_msg(action):
                print
                print self.get_long_msg(action)
                print
            options = self.get_options(action)
            if options is not None:
                print "Options available:"
                for o in options:
                    print "  " + o

    def cmd_sh(self, host = DEF_HOST, *unused, **options):
        """
        Connect to a zookeeper instance
        @usage: [host]
        +long: There are several ways to choose a host to connect to
        +long: 1. Use an environment variable Z_HOST, eg. Z_HOST=online z ls
        +long: 2. Initiate a connection using z sh, eg. z sh online
        +long: 3. Just use the default
        +long:
        +long: We supports specifying host by alias, we have two aliases built-in
        +long: z sh ksarch, which is the default, connects to ksarch test server
        +long: z sh online, connects to baidu production zookeeper server
        """
        data = {'host' : host}
        self.save_user_conf(data)

    def cmd_curhost(self, *unused, **options):
        """Get current host"""
        print self.host

    @need_connect
    @ensure_exist
    @fix_path(1)
    def cmd_ls(self, path = None, *unused, **options):
        """
        List children
        @usage: path
        """
        if path is None:
            path = self.get_user_conf_key('cwd')
        if path is None:
            path = '/'
        for i in sorted(self.zk.get_children(path, None)):
            print i

    @need_connect
    @ensure_exist
    @fix_path(1)
    def cmd_get(self, path, *unused, **options):
        """
        Get value
        @usage: path
        """
        ret = self.zk.get(path, None)
        print ret[0]

    @alias('get')
    def cmd_cat(self, path):
        pass

    @need_connect
    @fix_path(1)
    def cmd_set(self, path, value = None, *unused, **options):
        """
        Set value, read value from stdin if value not supplied
        @usage: path [value]
        """
        if value is None:
            value = sys.stdin.read()
        if self.zk.exists(path):
            version = self.get_version(path)
            self.zk.set(path, value, version)
        else:
            self.zk.create(path, value)

    @need_connect
    @option('-p', 'recursive', 'Create parent directories if needed', default = False)
    @ensure_exist
    @fix_path(1)
    def cmd_create(self, path, value = None, *unused, **options):
        """
        Create node if does not already exist
        @usage: [options] path [value]
        """
        if value is None:
            value = ''
        if not self.zk.exists(path):
            self.zk.create(path, value, makepath = options['recursive'] == True)
        else:
            raise ZError('Node already exist')

    @alias('create')
    def cmd_touch(self, path, value = None):
        pass

    @alias('create')
    def cmd_mkdir(self, path, value = None):
        pass

    def _cp(self, src, tgt):
        acl = self.zk.get_acls(src)[0]
        value = self.zk.get(src)[0]
        self.zk.create(tgt, value, acl)
        print src, ' -> ', tgt
        for c in self.zk.get_children(src):
            self._cp(src + '/' + c, tgt + '/' + c)

    @need_connect
    @fix_path(1,2)
    def cmd_cp(self, src, tgt, *unused, **options):
        """
        Copy a node tree to new location
        @usage: source target
        """
        if not self.zk.exists(src):
            raise ZError('Source does not exist')
        if self.zk.exists(tgt):
            raise ZError('Target already exists')
        self._cp(src, tgt)

    def _rm(self, path, recursive):
        if recursive:
            for c in self.zk.get_children(path):
                self._rm(path + '/' + c, recursive)
        version = self.get_version(path)
        self.zk.delete(path, version)
        print "deleted " + path

    @need_connect
    @option('-r', 'recursive', 'Recursively delete all subnodes', default = False)
    @fix_path(1)
    def cmd_rm(self, path, *unused, **options):
        """
        Delete single node, optionally delete a tree
        @usage: [options] path
        """
        if not self.zk.exists(path):
            return
        if not options['recursive'] and len(self.zk.get_children(path)) > 0:
            raise ZError('Node is not empty')
        self._rm(path, options['recursive'])

    @need_connect
    @fix_path(1)
    def cmd_stat(self, path, *unused, **options):
        """
        Get stats of a node
        @usage: path
        """
        stat = self.zk.exists(path)
        if stat:
            self.pretty_print_dict(stat._asdict())

    @need_connect
    @ensure_exist
    @fix_path(1)
    def cmd_getacl(self, path, *unused, **options):
        """
        Get acl of a node
        @usage: path
        """
        acl = self.zk.get_acls(path)[0]
        for i in acl:
            print i[1][0] + ':' + i[1][1] + ':' + str(i[0])

    def _parse_acl_string(self, s):
        s = s.split(':')
        if (len(s) < 3):
            return None
        return ACL(int(s[2]), Id(scheme=s[0], id=s[1]))

    @need_connect
    @ensure_exist
    @option('-a', 'append', 'Append mode', default = False)
    @fix_path(1)
    def cmd_setacl(self, path, *aclstr, **options):
        """
        Set acl of a node
        +long: Acl string example:
        +long: - world:anyone:1
        +long: - ip:10.34.56.78:31
        +long: 
        +long: Permission bit:
        +long: - 1 READ
        +long: - 2 WRITE
        +long: - 4 CREATE
        +long: - 8 DELETE
        +long: - 16 ADMIN
        @usage: path acl [acl...]
        """
        if len(aclstr) == 0:
            raise ZError("Acl not given")
        
        if options['append']:
            acl = self.zk.get_acls(path)[0]
        else:
            acl = []

        for s in aclstr:
            cur = self._parse_acl_string(s)
            if cur is None:
                raise ZError("Error in this acl string: " + s)
            acl.append(cur)

        self.zk.set_acls(path, acl)

    def _export(self, path):
        node = {}
        node['data'] = self.zk.get(path)[0]
        node['acl'] = self.zk.get_acls(path)[0]
        node['children'] = {}
        for c in self.zk.get_children(path):
            node['children'][c] = self._export(path + '/' + c)
        return node

    def _import(self, path, node, keepacl = True):
        print "Import " + path
        if keepacl:
            self.zk.create(path, node['data'], node['acl'])
        else:
            self.zk.create(path, node['data'])
        for i in node['children'].keys():
            self._import(path + '/' + i, node['children'][i], keepacl) 

    @need_connect
    @ensure_exist
    @option('-f', 'filename', 'Specify output filename, write to stdout by default', has_arg = True)
    @fix_path(1)
    def cmd_export(self, path, *unused, **options):
        """
        Export zk tree to a single file
        @usage: [options] path
        """
        filename = options['filename']
        if filename is None:
            print pickle.dumps(self._export(path))
        else:
            open(filename, 'w').write(pickle.dumps(self._export(path)))

    @need_connect
    @option('-f', 'filename', 'Specify input filename, read from stdin by default', has_arg = True)
    @option('-n', 'noacl', 'Use default acl (if not specified will import acl from file as well)')
    @fix_path(1)
    def cmd_import(self, path, *unused, **options):
        """
        Import zookeeper tree from previously exported file
        @usage: [options] path
        """
        filename = options['filename']
        if (self.zk.exists(path)):
            raise ZError("Node already exist")
        try:
            if filename is None:
                data = pickle.load(sys.stdin)
            else:
                data = pickle.load(open(filename))
        except IOError as e:
            raise ZError(e.strerror)
        self._import(path, data)

    @fix_path(1)
    def cmd_cd(self, path, *unused, **options):
        """
        Change working directory
        @usage: path
        """
        data = self.load_user_conf()
        data['cwd'] = path
        self.save_user_conf(data)

    def cmd_pwd(self, *unused, **options):
        """
        Show working directory
        """
        path = self.get_user_conf_key('cwd')
        if path is None:
            print '/'
        else:
            print path

if __name__ == '__main__':
    z = Z()
    try:
        action = sys.argv[1]
    except IndexError:
        action = 'help'

    try:
        param = sys.argv[2:]
    except IndexError:
        param = []

    try:
        z.call(action, param)
    except TypeError as e:
        error_print(e.message)
        sys.exit(1)
    except ZError as e:
        error_print(e.get())
        sys.exit(1)

# vim:set ft=python ts=4 sw=4 et:
