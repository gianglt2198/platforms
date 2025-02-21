package utils

import (
	"bytes"
	"encoding/json"
)

func TransformMaptoStruct(in, out interface{}) error {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(in); err != nil {
		return err
	}
	if err := json.NewDecoder(buf).Decode(out); err != nil {
		return err
	}
	return nil
}
