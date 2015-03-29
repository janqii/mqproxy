package server

import (
	"gitlab.baidu.com/hanjianqi/mqproxy/utils"
)

type ZookeeperLeaderElector struct {
	ZkClient   *utils.ZK
	LeaderPath string
}

func (el *ZookeeperLeaderElector) Startup() bool {
	go subscribeDataChanges(el)
	return el.Elect()
}

func (el *ZookeeperLeaderElector) Shutdown() {
}

func (el *ZookeeperLeaderElector) Elect() bool {
	return false
}

func subscribeDataChanges(el *ZookeeperLeaderElector) {
	go handleDataChanged(el)
	go handleDataDeleted(el)
}

func handleDataChanged(el *ZookeeperLeaderElector) {
}

func handleDataDeleted(el *ZookeeperLeaderElector) {
}
