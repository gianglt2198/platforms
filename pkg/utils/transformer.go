package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

func MapToStruct[T any](m map[string]interface{}) (*T, error) {
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	var res T

	err = json.Unmarshal(bytes, &res)

	return &res, err
}

func StructToMap[T any](d T) (*map[string]interface{}, error) {
	bytes, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}

	var m map[string]interface{}

	err = json.Unmarshal(bytes, &m)

	return &m, err
}

func TransformToType[T any](payload []byte) (*T, error) {

	var res T
	err := json.Unmarshal(payload, &res)

	return &res, err
}

func TransformToByteArray[T any](payload T) ([]byte, error) {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func TransformDateStrToDate(s string) (*time.Time, error) {
	date, err := time.Parse("2006-01-02", s)
	if err != nil || date.IsZero() {
		return nil, errors.New("invalid date string")
	}

	return &date, nil
}

func TransformISOTimeStrToTime(s string) (*time.Time, error) {
	date, err := time.Parse(time.RFC3339, s)

	if err != nil || date.IsZero() {
		return nil, errors.New("invalid iso time string")
	}

	return &date, nil
}

func JoinIntArrayToStr(a *[]int) *string {
	if a != nil && len(*a) > 0 {
		v := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(*a)), ","), "[]")
		return &v
	}
	return nil
}

func StructToStruct[T any](from interface{}) (*T, error) {
	var to T

	b, err := json.Marshal(from)
	if err != nil {
		return nil, err
	}

	_ = json.Unmarshal(b, &to)

	return &to, nil
}
