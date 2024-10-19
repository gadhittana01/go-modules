package utils

import (
	jsoniter "github.com/json-iterator/go"
)

var jsont jsoniter.API

func init() {
	jsont = jsoniter.ConfigCompatibleWithStandardLibrary
}

func Marshal(data any) ([]byte, error) {
	return jsont.Marshal(data)
}

func Unmarshal(data []byte, v any) error {
	return jsont.Unmarshal(data, v)
}

func UnmarshalT[T any](data []byte, v *T) (T, error) {
	err := jsont.Unmarshal(data, v)
	return *v, err
}

func JSONiter() jsoniter.API {
	return jsont
}
