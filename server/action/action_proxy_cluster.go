package action

import (
	"io"
	"net/http"
	//"io/ioutil"
)

func ProxyClusterAction(w http.ResponseWriter, r *http.Request) {
	//TODO:
	io.WriteString(w, "ProxyClusterAction!")
}
