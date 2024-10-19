package utils

import (
	"fmt"
	"net/http"
	"strings"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err != nil {
				generateErrorResp(err, w, r)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func getErrorMsg(message string) string {
	messages := strings.Split(message, "|")
	errMsg := messages[0]
	msg := messages[1]
	if !strings.EqualFold(errMsg, msg) {
		return fmt.Sprintf("%s: %s", msg, errMsg)
	}

	return errMsg
}

func generateErrorResp(err any, w http.ResponseWriter, r *http.Request) {
	var errorMsgs []map[string]interface{}
	var statusCode int
	var errorCode int

	appErr, isAppErr := err.(AppError)
	validationErr, isValidationErr := err.(ValidationErrors)

	if isAppErr {
		statusCode = appErr.StatusCode
		errorCode = appErr.ErrorCode
		messages := strings.Split(appErr.Message, "|")
		errorMsgs = []map[string]interface{}{
			{"message": messages[1]},
		}

		if appErr.StatusCode >= 500 {
			message := getErrorMsg(appErr.Message)
			LogError(fmt.Sprintf("APP ERROR (PANIC) %s", message))
		}

		if appErr.StatusCode >= 400 {
			LogWarning(fmt.Sprintf("APP ERROR (PANIC) %s", messages[0]))
		}

	} else if isValidationErr {
		LogWarning(fmt.Sprintf("VALIDATION ERROR (PANIC) %v", validationErr))

		for _, err := range validationErr.Errors {
			errorMsg := map[string]interface{}{
				"message": err.Message,
			}
			errorMsgs = append(errorMsgs, errorMsg)
		}

		statusCode = validationErr.StatusCode
	} else {
		LogError(fmt.Sprintf("UNKNOWN ERROR (PANIC) %v", err))
		errorMsgs = []map[string]interface{}{
			{"message": "internal server error"},
		}

		statusCode = 500
	}

	GenerateErrorResp(w, errorMsgs, statusCode, errorCode)
}
