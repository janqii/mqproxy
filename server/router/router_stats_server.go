package router

import (
	"github.com/janqii/mqproxy/server/action"
	"net/http"
)

func StatServerRouter(mux map[string]func(http.ResponseWriter, *http.Request)) {
	mux["/stats"] = action.StatsAction
}
