package action

import (
	"gitlab.baidu.com/hanjianqi/mqproxy/serializer"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type HttpProducerRequest struct {
	topic        string
	partitionKey string
	data         interface{}
}

type HttpProducerResponse struct {
	errno  int
	errmsg string
	data   []messageLocation
}

type messageLocation struct {
	partition int
	offset    int
}

func HttpProducerAction(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("read request body error")
		return
	}

	query := r.URL.Query()
	msgConverter := strings.Join(query["format"], "")

	reqData := &HttpProducerRequest{}
	s := &serializer.Serializer{Converter: msgConverter}

	if err = s.Unmarshal(body, &reqData); err != nil {
		log.Printf("Unmarshal HttpRequest error, %v", err)
		res := &HttpProducerResponse{
			errno:  -1,
			errmsg: "unmarshal http request data error",
			data:   make([]messageLocation, 0),
		}
		echo2client(w, s, res, err)
	}

	log.Println(reqData)

	io.WriteString(w, "HttpProducerAction!")
}

func NsheadProduerAction() {
}

func echo2client(w http.ResponseWriter, s *serializer.Serializer, res *HttpProducerResponse, e error) {
	b, e := s.Marshal(res)
	if e != nil {
		log.Printf("marshal http response error, %v", e)
	} else {
		io.WriteString(w, string(b))
	}
}
