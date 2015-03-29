package serializer

import (
	"bytes"
	"encoding/json"
	"errors"
)

type Serializer struct {
	Converter string
}

func (s *Serializer) Unmarshal(data []byte, v interface{}) error {
	if s.Converter == "json" {
		return unmarshalJson(data, v)
	}

	return errors.New("unkonwn converter")
}

func (s *Serializer) Marshal(v interface{}) ([]byte, error) {
	if s.Converter == "json" {
		return marshalJson(v)
	}

	buf := &bytes.Buffer{}
	return buf.Bytes(), errors.New("unkonwn converter")
}

func marshalJson(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func unmarshalJson(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
