package config

import (
	"flag"
	"fmt"
	"reflect"
	"strings"
)

type fromFlagsProvider valueProviderFunc

func (p fromFlagsProvider) Provide(confField string, field reflect.StructField) (valuer, error) {
	return p(confField, field)
}

func (p fromFlagsProvider) AfterProvide() error {
	flag.Parse()
	return nil
}

var fromFlags = fromFlagsProvider(fromFlagsValue)

func fromFlagsValue(confField string, field reflect.StructField) (valuer, error) {
	f := flagger{
		stringValuer: stringValuer{
			kind: field.Type.Kind(),
		},
	}

	flag.Var(&f, flagName(confField), field.Tag.Get("help"))
	return &f, nil
}

type flagger struct {
	stringValuer
}

func (f *flagger) IsBoolFlag() bool {
	return f.kind == reflect.Bool
}
func (f *flagger) String() string {
	return f.input
}
func (f *flagger) Set(val string) error {
	f.input = val
	return nil
}

func flagName(field string) string {
	return strings.ReplaceAll(field, " ", "-")
}

func flagNameFull(field string) string {
	return fmt.Sprintf("--%s", flagName(field))
}
