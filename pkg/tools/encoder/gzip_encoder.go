package encoder

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
)

type gzipEncoder struct{}

func NewGzipEncoder() *gzipEncoder {
	return &gzipEncoder{}
}

func (gz *gzipEncoder) Encode(input interface{}) (string, error) {
	jsonData, err := json.Marshal(input)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	_, err = gzipWriter.Write(jsonData)
	if err != nil {
		return "", err
	}
	gzipWriter.Close()

	return buf.String(), nil
}

func (gz *gzipEncoder) Decode(input string, output interface{}) error {
	buf := bytes.NewBufferString(input)
	gzipReader, err := gzip.NewReader(buf)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	decodedData, err := io.ReadAll(gzipReader)
	if err != nil {
		return err
	}

	return json.Unmarshal(decodedData, output)
}
