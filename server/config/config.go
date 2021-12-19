package config

import (
	"fmt"
	"strconv"
	"strings"
)

type Conf struct {
	DB   dbConf
	HTTP httpConf
}

type dbConf struct {
	// DSN
	Host   string
	Port   string
	Name   string
	Schema string
	User   string
	Pass   string

	// Migrations
	Migrate        bool
	MigrationsPath string
}

type httpConf struct {
	Address   string
	Port      string
	JWTSecret string
	APIKey    string
}

// Root field name for config. Convention over configuration!
// E.g "db host" field as env var will look like DB_HOST.
const (
	ConfDBHost   = "db host"
	ConfDBPort   = "db port"
	ConfDBName   = "db name"
	ConfDBSchema = "db schema"
	ConfDBUser   = "db user"
	ConfDBPass   = "db pass"

	ConfDBMigrate        = "db migrate"
	ConfDBMigrationsPath = "db migrations path"

	ConfHTTPAddress   = "http address"
	ConfHTTPPort      = "http port"
	ConfHTTPJWTSecret = "http jwt secret"
	ConfHTTPAPIKey    = "http api key"
)

func (c Conf) Validate() error {

	valErr := newValidationError()

	// Test config fields for errors
	if c.DB.Host == "" {
		valErr.empty(ConfDBHost)
	}
	if c.DB.Port == "" {
		valErr.empty(ConfDBPort)
	}
	if _, err := strconv.Atoi(c.DB.Port); err != nil {
		valErr.finvalid(ConfDBPort, "has to be a number")
	}
	if c.DB.Name == "" {
		valErr.empty(ConfDBName)
	}
	if c.DB.Schema == "" {
		valErr.empty(ConfDBSchema)
	}
	if c.DB.User == "" {
		valErr.empty(ConfDBUser)
	}
	if c.DB.Pass == "" {
		valErr.empty(ConfDBPass)
	}
	if c.DB.MigrationsPath == "" {
		valErr.empty(ConfDBMigrationsPath)
	}
	if c.HTTP.Address == "" {
		valErr.empty(ConfHTTPAddress)
	}
	if c.HTTP.Port == "" {
		valErr.empty(ConfHTTPPort)
	}
	if _, err := strconv.Atoi(c.HTTP.Port); err != nil {
		valErr.finvalid(ConfHTTPPort, "has to be a number")
	}
	if c.HTTP.JWTSecret == "" {
		valErr.empty(ConfHTTPJWTSecret)
	}
	if c.HTTP.APIKey == "" {
		valErr.empty(ConfHTTPAPIKey)
	}

	// Final result
	if !valErr.isOk() {
		return valErr
	}

	return nil
}

// Construct DSN from config
func (c Conf) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", c.DB.User, c.DB.Pass, c.DB.Host, c.DB.Port, c.DB.Name)
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

func (e *ConfValidationError) invalid(field, reason string) {
	if _, ok := e.invalidFields[field]; !ok {
		e.invalidFields[field] = make([]string, 0, 1)
	}

	e.invalidFields[field] = append(e.invalidFields[field], reason)
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

			Migrate:        true,
			MigrationsPath: "file://postgres/migration",
		},
		HTTP: httpConf{
			Address: "0.0.0.0",
			Port:    "5001",
		},
	}

	// Go from least to most specific source. E.g flags override env vars and env vars override file config
	fromFile(&conf)
	fromEnv(&conf)
	fromFlags(&conf)

	// Validate the config, to make sure application can use it
	if err := conf.Validate(); err != nil {
		return Conf{}, err
	}

	return conf, nil
}
