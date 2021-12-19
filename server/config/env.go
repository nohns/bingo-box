package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func fromEnv(conf *Conf) {

	// Try to load from .env file, if it exists
	godotenv.Load()

	if val := os.Getenv(envVarName(ConfDBHost)); val != "" {
		conf.DB.Host = val
	}
	if val := os.Getenv(envVarName(ConfDBPort)); val != "" {
		conf.DB.Port = val
	}
	if val := os.Getenv(envVarName(ConfDBName)); val != "" {
		conf.DB.Name = val
	}
	if val := os.Getenv(envVarName(ConfDBSchema)); val != "" {
		conf.DB.Schema = val
	}
	if val := os.Getenv(envVarName(ConfDBUser)); val != "" {
		conf.DB.User = val
	}
	if val := os.Getenv(envVarName(ConfDBPass)); val != "" {
		conf.DB.Pass = val
	}

	if val := os.Getenv(envVarName(ConfDBMigrate)); val != "" {
		boolStr := strings.ToLower(val)
		if boolStr == "false" {
			conf.DB.Migrate = false
		} else if boolStr == "true" {
			conf.DB.Migrate = false
		}
	}
	if val := os.Getenv(envVarName(ConfDBMigrationsPath)); val != "" {
		conf.HTTP.Address = val
	}

	if val := os.Getenv(envVarName(ConfHTTPAddress)); val != "" {
		conf.HTTP.Address = val
	}
	if val := os.Getenv(envVarName(ConfHTTPPort)); val != "" {
		conf.HTTP.Port = val
	}
	if val := os.Getenv(envVarName(ConfHTTPJWTSecret)); val != "" {
		conf.HTTP.JWTSecret = val
	}
	if val := os.Getenv(envVarName(ConfHTTPAPIKey)); val != "" {
		conf.HTTP.APIKey = val
	}
}

func envVarName(field string) string {
	return strings.ToUpper(strings.ReplaceAll(field, " ", "_"))
}
