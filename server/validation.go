package bingo

import (
	"fmt"
)

type validationFieldErrs map[string][]validationFieldErr

type ValidationErr struct {
	FieldErrs validationFieldErrs

	original error
	prefix   string
}

type validationFieldErr struct {
	Kind string
	Msg  string
}

func (vr ValidationErr) withPrefix(format string, v ...interface{}) ValidationErr {
	return ValidationErr{
		FieldErrs: vr.copyErrs(),
		original:  vr,
		prefix:    fmt.Sprintf(format, v...),
	}
}

func (vr ValidationErr) copyErrs() validationFieldErrs {
	fieldErrs := make(validationFieldErrs)
	for field, errs := range vr.FieldErrs {
		fieldErrs[field] = errs
	}

	return fieldErrs
}

func (vr ValidationErr) withFieldErr(field, kind, format string, v ...interface{}) ValidationErr {
	fieldErrs := vr.copyErrs()
	fieldErrs[field] = append(fieldErrs[field], validationFieldErr{kind, fmt.Sprintf(format, v...)})

	return ValidationErr{
		original:  vr,
		FieldErrs: fieldErrs,
	}
}

func (vr ValidationErr) Error() string {

	s := vr.prefix + "\n"
	for field, valErrs := range vr.FieldErrs {
		for _, valErr := range valErrs {
			s += fmt.Sprintf("\n - %s: (%s) %s", field, valErr.Kind, valErr.Msg)
		}
	}

	return s
}

func (vr ValidationErr) Unwrap() error {
	return vr.original
}

func NewValErr(prefix string) ValidationErr {
	return ValidationErr{
		FieldErrs: make(validationFieldErrs),
		original:  nil,
		prefix:    prefix,
	}
}
