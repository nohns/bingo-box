package config

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// TODO: overhaul config system with reflection instead of constants
type Conf struct {
	DB   dbConf   `conf:"db" validate:"required"`
	HTTP httpConf `conf:"http" validate:"required"`
	Mail mailConf `conf:"mail" validate:"required"`
}

type dbConf struct {
	// DSN
	Host   string `validate:"required" help:"Host of PostgreSQL database"`
	Port   string `validate:"required" help:"Port of PostgreSQL database"`
	Name   string `validate:"required" help:"Database name used"`
	Schema string `help:"Schema used in the database"`
	User   string `validate:"required" help:"Username of PostgreSQL user"`
	Pass   string `validate:"required" help:"Password of PostgreSQL user"`

	// Migrations
	Migrate        bool   `help:"Set this flag to false to disable migrations"`
	MigrationsPath string `validate:"required" help:"A path to the folder containing migrations"`
}

type httpConf struct {
	Address   string `validate:"required" help:"Address for HTTP server to listen on"`
	Port      string `validate:"required" help:"Port for HTTP server to listen on"`
	JWTSecret string `conf:"jwt secret" validate:"required" help:"JWT secret used for signing access tokens"`
	APIKey    string `conf:"api key" validate:"required" help:"(deprecated) API key used for remote autorized access"`
}

type mailConf struct {
	DLLinkBase string `conf:"dl link base" validate:"required" help:""`

	// Mail gun credentials
	MGDomain string `conf:"mg domain" validate:"required"`
	MGAPIKey string `conf:"mg api key" validate:"required"`
}

func (c Conf) Validate(confNamespaces map[string]string) error {

	valErr := newValidationError()

	v := validator.New()
	err := v.Struct(c)
	if err != nil {
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			return err
		}

		for _, err := range errs {
			confField := confNamespaces[err.StructNamespace()]
			switch err.Tag() {
			case "required":
				valErr.empty(confField)
			default:
				valErr.invalid(confField, "validation with tag '%s' failed", err.Tag())
			}
		}
	}

	// Final result
	if !valErr.isOk() {
		return valErr
	}

	return nil
}

// Construct DSN from config
func (c Conf) ConnURI() string {
	return fmt.Sprintf("mongodb://%s:%s@%s:%s/%s", c.DB.User, c.DB.Pass, c.DB.Host, c.DB.Port, c.DB.Name)
}

// Construct http listener address from config
func (c Conf) HTTPListenAddr() string {
	return fmt.Sprintf("%s:%s", c.HTTP.Address, c.HTTP.Port)
}

type ConfValidationError struct {
	// invalid config fields. Key = field, value = reason
	invalidFields map[string][]string
}

func (e *ConfValidationError) empty(field string) {
	e.invalid(field, fmt.Sprintf("%s cannot be empty %s", field, e.confHelp(field)))
}

func (e *ConfValidationError) confHelp(field string) string {
	flag, envVar, fileKey := flagNameFull(field), envVarName(field), fileKeyName(field)
	return fmt.Sprintf("(set flag %s, env var %s or key %s in config file)", flag, envVar, fileKey)
}

func (e *ConfValidationError) finvalid(field, reason string) {
	e.invalid(field, fmt.Sprintf("%s %s", field, reason))
}

func (e *ConfValidationError) invalid(field, reasonFmt string, v ...interface{}) {
	if _, ok := e.invalidFields[field]; !ok {
		e.invalidFields[field] = make([]string, 0, 1)
	}

	e.invalidFields[field] = append(e.invalidFields[field], fmt.Sprintf(reasonFmt, v...))
}

func (e *ConfValidationError) isOk() bool {
	return len(e.invalidFields) == 0
}

func (e *ConfValidationError) Error() string {

	fieldReasons := make([]string, 0, len(e.invalidFields))
	for field, reasons := range e.invalidFields {
		reason := strings.Join(reasons, ", ")
		fieldReasons = append(fieldReasons, fmt.Sprintf("  - %s: %s", field, reason))
	}

	invalids := strings.Join(fieldReasons, "\n")
	return fmt.Sprintf("config: following fields were invalid\n\n%v", invalids)
}

func newValidationError() *ConfValidationError {
	return &ConfValidationError{
		invalidFields: make(map[string][]string),
	}
}

func Read() (Conf, error) {

	// Defaults
	conf := Conf{
		DB: dbConf{
			Host:   "127.0.0.1",
			Port:   "5432",
			Name:   "bingo_box_db",
			Schema: "public",

			Migrate:        false,
			MigrationsPath: "file://postgres/migration",
		},
		HTTP: httpConf{
			Address: "0.0.0.0",
			Port:    "5001",
		},
	}

	// Go from least to most specific source. E.g file config is overridden by env, and env is overridden by flags
	refl, err := reflectVals(&conf, fromEnv, fromFlags /*, fromFile*/)
	if err != nil {
		return Conf{}, err
	}
	// Validate the config, to make sure application can use it
	if err := conf.Validate(refl.namespaces); err != nil {
		return Conf{}, err
	}

	return conf, nil
}
