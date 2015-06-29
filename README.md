# mqproxy

###依赖
go get gopkg.in/Shopify/sarama.v1

go get gopkg.in/vmihailenco/msgpack.v2

###启动
./mqproxy -c ./proxy.cfg 

###生产消息

```php
<?php
    $ch =curl_init("http://10.46.188.58:8081/trigger?log_id=132");
    curl_setopt($ch, CURLOPT_HEADER, 0);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch, CURLOPT_BINARYTRANSFER, true);

    $data = array(
        'cmd'   => '/home/arch/local/agent/test/test.sh',
        "type"  => "hook",
        "env"   => array(
            "HOST_NAME" => "tc-arch-redis.baidu.com",
            "UNIT_ID"   => "confilter.ksarch.all"
        ),
        "param" => array(
            "version"   => "1.0.0.0",
        ),
    );

    $json_str = json_encode($data);
    var_dump($json_str);
    curl_setopt($ch, CURLOPT_POSTFIELDS, $json_str);
    $res = curl_exec($ch);
    var_dump($res);
?>
```
