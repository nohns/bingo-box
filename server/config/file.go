package config

import "strings"

func fromFile(conf *Conf) {
	// Missing implementation for now
}

func fileKeyName(field string) string {
	return strings.ReplaceAll(field, " ", ".")
}
