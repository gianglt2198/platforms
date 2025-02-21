package encoder

import (
	"encoding/json"
)

type jsonEncoder struct{}

func NewJsonEncoder() *jsonEncoder {
	return &jsonEncoder{}
}

func (js *jsonEncoder) Encode(input interface{}) (string, error) {
	r, err := json.Marshal(input)
	return string(r), err
}

func (js *jsonEncoder) Decode(input string, output interface{}) error {
	return json.Unmarshal([]byte(input), output)
}
