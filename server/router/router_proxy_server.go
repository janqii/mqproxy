package router

import (
	"github.com/janqii/mqproxy/server/action"
	"net/http"
)

func ProxyServerRouter(mux map[string]func(http.ResponseWriter, *http.Request)) {
	mux["/topics"] = action.TopicsAction

	mux["/produce"] = action.HttpProducerAction
	mux["/consumer/fetch"] = action.FetchMessageAction
}
