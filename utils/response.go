package utils

import (
	"net/http"
	"strconv"
)

type SuccessResponse[T any] struct {
	Success    bool `json:"success"`
	StatusCode int  `json:"statusCode"`
	Data       T    `json:"data"`
}

type FailedResponse[T any] struct {
	Success    bool `json:"success"`
	StatusCode int  `json:"statusCode"`
	ErrorCode  int  `json:"errorCode"`
	Errors     T    `json:"errors"`
}

// FOR TESTING PURPOSE
type ResponseMap struct {
	Success    bool                   `json:"success"`
	StatusCode int                    `json:"statusCode"`
	Data       map[string]interface{} `json:"data"`
}

func MustParseStringToInt(str string) int {
	val, err := strconv.Atoi(str)
	if err != nil {
		panic(err)
	}

	return val
}

func GenerateSuccessResp[T any](w http.ResponseWriter, data T, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := SuccessResponse[T]{
		Success:    true,
		StatusCode: statusCode,
		Data:       data,
	}

	responseEncode, err := Marshal(response)
	if err != nil {
		panic(err)
	}

	_, err = w.Write(responseEncode)
	if err != nil {
		panic(err)
	}
}

func GenerateErrorResp[T any](w http.ResponseWriter, data T, statusCode int, errorCode ...int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := FailedResponse[T]{
		Success:    false,
		StatusCode: statusCode,
		Errors:     data,
	}

	responseEncode, err := Marshal(response)
	if err != nil {
		panic(err)
	}

	_, err = w.Write(responseEncode)
	if err != nil {
		panic(err)
	}
}

func GenerateDefaultResp(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	responseEncode, err := Marshal(data)
	if err != nil {
		panic(err)
	}

	_, err = w.Write(responseEncode)
	if err != nil {
		panic(err)
	}
}
