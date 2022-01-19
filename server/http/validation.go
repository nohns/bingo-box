package http

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	bingo "github.com/nohns/bingo-box/server"
)

type validationFieldError struct {
	Reason  string `json:"reason"`
	Message string `json:"message,omitempty"`
}

type validationField struct {
	FieldName string                 `json:"fieldName"`
	Value     interface{}            `json:"value"`
	Errors    []validationFieldError `json:"errors"`
}

type validationData map[string]*validationField

// Add error to validation data
func (vd validationData) err(field, reason, message string, value interface{}) {
	if _, ok := vd[field]; !ok {
		vd[field] = &validationField{
			FieldName: field,
			Value:     value,
			Errors:    make([]validationFieldError, 0, 1),
		}
	}

	vd[field].Errors = append(vd[field].Errors, validationFieldError{
		Reason:  reason,
		Message: message,
	})
}

func (vd validationData) Error() string {
	fields := make([]string, 0, len(vd))
	for field := range vd {
		fields = append(fields, field)
	}
	return fmt.Sprintf("Validation failed for fields: %s", strings.Join(fields, ", "))
}

func (vd validationData) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{})
	for k, v := range vd {
		m[k] = v
	}
	return json.Marshal(map[string]interface{}{"fieldErrors": m})
}

func (vd validationData) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &vd)
}

// Validate a given request body. If validation succeeds, then 2nd return value, ok, will be true but otherwise false
func validateBody(body interface{}) (error, bool) {
	v := validator.New()
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

		if name == "-" {
			return ""
		}

		return name
	})
	err := v.Struct(body)
	if errs, ok := err.(validator.ValidationErrors); ok {
		validationData := make(validationData)
		for _, err := range errs {
			validationData.err(err.Field(), err.Tag(), err.Error(), err.Value())
		}

		return validationData, false
	}

	return err, err == nil
}

func translateBingoValidationErr(err bingo.ValidationErr) validationData {
	valData := make(validationData)
	for f, fErrs := range err.FieldErrs {
		lowerF := strings.ToLower(f[:1]) + f[1:]

		vFieldErrs := make([]validationFieldError, 0, len(fErrs))
		for _, fe := range fErrs {
			vFieldErrs = append(vFieldErrs, validationFieldError{
				Reason:  fe.Kind,
				Message: fe.Msg,
			})
		}

		valData[lowerF] = &validationField{
			FieldName: lowerF,
			Errors:    vFieldErrs,
		}
	}

	return valData
}
