package postgres

import (
	"errors"
	"fmt"
	"os"
)

type dsnConfig struct {
	host   string
	port   string
	dbName string
	schema string
	user   string
	pass   string
}

var (
	ErrHostNotSet   = errors.New("postgres: host not set")
	ErrPortNotSet   = errors.New("postgres: port not set")
	ErrDBNameNotSet = errors.New("postgres: database name not set")
	ErrSchemaNotSet = errors.New("postgres: schema not set")
	ErrUserNotSet   = errors.New("postgres: username not set")
	ErrPassNotSet   = errors.New("postgres: password not set")
)

func (c *dsnConfig) From() string {
	return "environment"
}

func (c *dsnConfig) DSN() (string, error) {

	// Validate config
	switch {
	case c.host == "":
		return "", ErrHostNotSet
	case c.port == "":
		return "", ErrPortNotSet
	case c.dbName == "":
		return "", ErrDBNameNotSet
	case c.schema == "":
		return "", ErrSchemaNotSet
	case c.user == "":
		return "", ErrUserNotSet
	case c.pass == "":
		return "", ErrPassNotSet
	}

	// Create DSN
	return fmt.Sprintf("postgres://%s:%s@tcp(%s:%s)/%s?schema=%s", c.user, c.pass, c.host, c.port, c.dbName, c.schema), nil
}

func DSNFromEnv() *dsnConfig {
	return &dsnConfig{
		host:   os.Getenv("POSTGRES_HOST"),
		port:   os.Getenv("POSTGRES_PORT"),
		dbName: os.Getenv("POSTGRES_DB"),
		schema: os.Getenv("POSTGRES_SCHEMA"),
		user:   os.Getenv("POSTGRES_USER"),
		pass:   os.Getenv("POSTGRES_PASS"),
	}
}
