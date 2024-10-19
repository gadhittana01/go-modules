package utils

import (
	"fmt"
	"strconv"
	"strings"
)

type AppError struct {
	Message    string
	StatusCode int
	ErrorCode  int
}

func (ae *AppError) Error() string {
	return fmt.Sprintf("app error: status code %d, message %s", ae.StatusCode, ae.Message)
}

type ValidationError struct {
	Message string
	Field   string
	Tag     string
}

func (ve *ValidationError) Error() string {
	return fmt.Sprintf("validation error: message %s", ve.Message)
}

type ValidationErrors struct {
	Errors     []ValidationError
	StatusCode int
}

func CustomError(message string, statusCode int) error {
	return fmt.Errorf("%s|%s<->%d|", message, message, statusCode)
}

func CustomErrorWithTrace(err error, message string, statusCode int) error {
	return fmt.Errorf("%s|%s<->%d|", err.Error(), message, statusCode)
}

func PanicIfError(err error) {
	if err != nil {
		customError := strings.Split(err.Error(), "<->")
		message := customError[0]
		statusCode := 500

		if len(customError) > 1 {
			statusCode, _ = strconv.Atoi(strings.Split(customError[1], "|")[0])
		}

		appErr := AppError{
			Message:    message,
			StatusCode: statusCode,
		}
		panic(appErr)
	}
}

func PanicIfAppError(err error, message string, statusCode int) {
	if err != nil {
		customErr := CustomErrorWithTrace(err, message, statusCode)
		PanicIfError(customErr)
	}
}

func PanicAppError(message string, statusCode int) {
	customErr := CustomError(message, statusCode)
	PanicIfError(customErr)
}

func PanicValidationError(errors []ValidationError, statusCode int) {
	validationErrors := ValidationErrors{
		Errors:     errors,
		StatusCode: statusCode,
	}
	panic(validationErrors)
}
