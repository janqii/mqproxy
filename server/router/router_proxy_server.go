package router

import (
	"gitlab.baidu.com/hanjianqi/mqproxy/server/action"
	"net/http"
)

func ProxyServerRouter(mux map[string]func(http.ResponseWriter, *http.Request)) {
	mux["/topics"] = action.TopicsAction
	mux["/parititons"] = action.PartitionsAction

	mux["/produce"] = action.HttpProducerAction

	mux["/consumer/register"] = action.RegisterConsumerAction
	mux["/consumer/delete"] = action.DeleteConsumerAction
	mux["/consumer/fetch"] = action.FetchMessageAction

	mux["/offset/commit"] = action.CommitOffsetAction
	mux["/offset/fetch"] = action.FetchOffsetAction

	mux["/brokers"] = action.BrokersAction
	mux["/proxy"] = action.ProxyClusterAction
	mux["/controller"] = action.ControllerAction
}
