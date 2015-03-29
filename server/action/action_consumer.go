package action

import (
	"io"
	"net/http"
	//"io/ioutil"
)

func RegisterConsumerAction(w http.ResponseWriter, r *http.Request) {
	//TODO:
	io.WriteString(w, "RegisterConsumerAction!")
}

func FetchMessageAction(w http.ResponseWriter, r *http.Request) {
	//TODO:
	io.WriteString(w, "FetchMessageAction!")
}

func DeleteConsumerAction(w http.ResponseWriter, r *http.Request) {
	//TODO:
	io.WriteString(w, "DeleteConsumerAction!")
}
