package service

import (
	"os"
	"strings"
)

var (
	version string
	name    string
)

const (
	unversioned_value = "unversioned"
	path_separator    = "/"
)

func GetServiceName() string {
	return name
}

func GetVersion() string {
	return version
}

func initVersion() {
	if len(version) > 0 {
		return
	}
	version = unversioned_value
}

func initName() {
	if len(name) > 0 {
		return
	}
	parts := strings.Split(os.Args[0], path_separator)
	lenParts := len(parts)
	if lenParts <= 0 {
		return
	}
	name = parts[lenParts-1]
}

func init() {
	initVersion()
	initName()
}
