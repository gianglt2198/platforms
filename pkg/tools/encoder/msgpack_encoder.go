package encoder

import (
	"bytes"

	"github.com/vmihailenco/msgpack/v5"
)

type msgpackEncoder struct {
	encoder *msgpack.Encoder
	decoder *msgpack.Decoder
}

func NewMsgpackEncoder() *msgpackEncoder {
	encoder := msgpack.NewEncoder(nil)
	encoder.SetCustomStructTag("json")
	decoder := msgpack.NewDecoder(nil)
	decoder.SetCustomStructTag("json")
	return &msgpackEncoder{
		encoder: encoder,
		decoder: decoder,
	}
}

func (js *msgpackEncoder) Encode(input interface{}) (string, error) {
	var buf bytes.Buffer
	js.encoder.ResetWriter(&buf)
	err := js.encoder.Encode(input)
	return buf.String(), err
}

func (js *msgpackEncoder) Decode(input string, output interface{}) error {
	js.decoder.ResetReader(bytes.NewReader([]byte(input)))
	return js.decoder.Decode(&output)
}
