package config

import (
	"os"
	"reflect"
	"strings"

	_ "github.com/joho/godotenv/autoload"
)

type fromEnvProvider valueProviderFunc

func (p fromEnvProvider) Provide(confField string, field reflect.StructField) (valuer, error) {
	return p(confField, field)
}

var fromEnv = fromEnvProvider(fromEnvValue)

func fromEnvValue(confField string, field reflect.StructField) (valuer, error) {
	return &stringValuer{
		input: os.Getenv(envVarName(confField)),
		kind:  field.Type.Kind(),
	}, nil
}

func envVarName(field string) string {
	return strings.ToUpper(strings.ReplaceAll(field, " ", "_"))
}
