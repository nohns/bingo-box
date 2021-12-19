package config

import (
	"flag"
	"fmt"
	"strings"
)

func fromFlags(conf *Conf) {
	flag.StringVar(&conf.DB.Host, flagName(ConfDBHost), conf.DB.Host, "")
	flag.StringVar(&conf.DB.Port, flagName(ConfDBPort), conf.DB.Port, "")
	flag.StringVar(&conf.DB.Name, flagName(ConfDBName), conf.DB.Name, "")
	flag.StringVar(&conf.DB.Schema, flagName(ConfDBSchema), conf.DB.Schema, "")
	flag.StringVar(&conf.DB.User, flagName(ConfDBUser), conf.DB.User, "")
	flag.StringVar(&conf.DB.Pass, flagName(ConfDBPass), conf.DB.Pass, "")

	flag.BoolVar(&conf.DB.Migrate, flagName(ConfDBMigrate), conf.DB.Migrate, "")
	flag.StringVar(&conf.DB.MigrationsPath, flagName(ConfDBMigrationsPath), conf.DB.MigrationsPath, "")

	flag.StringVar(&conf.HTTP.Address, flagName(ConfHTTPAddress), conf.HTTP.Address, "")
	flag.StringVar(&conf.HTTP.Port, flagName(ConfHTTPPort), conf.HTTP.Port, "")
	flag.StringVar(&conf.HTTP.JWTSecret, flagName(ConfHTTPJWTSecret), conf.HTTP.JWTSecret, "")
	flag.StringVar(&conf.HTTP.APIKey, flagName(ConfHTTPAPIKey), conf.HTTP.APIKey, "")
}

func flagName(field string) string {
	return strings.ReplaceAll(field, " ", "-")
}

func flagNameFull(field string) string {
	return fmt.Sprintf("--%s", flagName(field))
}
