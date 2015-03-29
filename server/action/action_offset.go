package action

import (
	"io"
	"net/http"
	//"io/ioutil"
)

func CommitOffsetAction(w http.ResponseWriter, r *http.Request) {
	//TODO:
	io.WriteString(w, "CommitOffsetAction!")
}

func FetchOffsetAction(w http.ResponseWriter, r *http.Request) {
	//TODO:
	io.WriteString(w, "FetchOffsetAction!")
}
