package server

import (
	"gitlab.baidu.com/hanjianqi/mqproxy/utils"
)

type ProxyController struct {
	ZkClient     *utils.ZK
	IsRunning    bool
	ID           int
	IsController bool
}

func (cc *ProxyController) Startup() {
	//TODO:
	// just for debugging
	cc.IsRunning = true
	cc.IsController = true
}

func (cc *ProxyController) ReAssgin() {
	cc.IsRunning = true
	cc.IsController = false
}

func (cc *ProxyController) Shutdown() {
	//TODO: cleanup env
}
