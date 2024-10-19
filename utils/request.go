package utils

import (
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

func generateValidationParamErrorMsg(paramName string) string {
	return fmt.Sprintf("invalid param %s", paramName)
}

func generateValidationQueryErrorMsg(query string) string {
	return fmt.Sprintf("invalid query param %s", query)
}

func ValidateURLParamUUID(r *http.Request, paramName string, defaultValue ...uuid.UUID) uuid.UUID {
	param := chi.URLParam(r, paramName)

	uuid, err := uuid.Parse(param)
	if err != nil {
		if len(defaultValue) > 0 {
			uuid = defaultValue[0]
		} else {
			PanicIfError(CustomErrorWithTrace(err, generateValidationParamErrorMsg(paramName), 400))
		}
	}

	return uuid
}

func ValidateBodyPayload[T any](body io.ReadCloser, output *T) T {
	err := JSONiter().NewDecoder(body).Decode(output)
	PanicIfAppError(err, "failed when decode body payload", 400)

	ValidateStruct(output)
	return *output
}

func ValidateQueryParamInt(r *http.Request, queryName string, defaultValue ...int) int {
	var queryInt int
	var err error
	query := r.URL.Query().Get(queryName)

	if query != "" {
		queryInt, err = strconv.Atoi(query)
		if err != nil || queryInt < 0 {
			PanicIfError(CustomErrorWithTrace(err, generateValidationQueryErrorMsg(queryName), 400))
		}
	} else if len(defaultValue) > 0 {
		queryInt = defaultValue[0]
	}

	return queryInt
}

func ValidateStruct(data interface{}) {
	var validationErrors []ValidationError
	validate := validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

		if name == "-" {
			return ""
		}

		return name
	})
	errorValidate := validate.Struct(data)

	if errorValidate != nil {
		for _, err := range errorValidate.(validator.ValidationErrors) {
			var validationError ValidationError
			validationError.Message = strings.Split(err.Error(), "Error:")[1]
			validationError.Field = err.Field()
			validationError.Tag = err.Tag()
			validationErrors = append(validationErrors, validationError)
		}
		PanicValidationError(validationErrors, 400)
	}
}
