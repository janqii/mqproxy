package server

import (
	"gitlab.baidu.com/hanjianqi/mqproxy/utils"
)

type ZookeeperHealthChecker struct {
	ZkClient  *utils.ZK
	IsRunning bool
}

func (checker *ZookeeperHealthChecker) Startup() {
	checker.IsRunning = true
}

func (checker *ZookeeperHealthChecker) ShutDown() {
}
