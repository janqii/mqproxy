package action

import (
	"io"
	"net/http"
	//"io/ioutil"
)

func ControllerAction(w http.ResponseWriter, r *http.Request) {
	//TODO:
	io.WriteString(w, "ControllerAction!")
}
