package action

import (
	"io"
	"net/http"
	//"io/ioutil"
)

func BrokersAction(w http.ResponseWriter, r *http.Request) {
	//TODO:
	io.WriteString(w, "BrokerAction!")
}
