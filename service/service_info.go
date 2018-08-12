package service

import (
	"os"
	"strings"
)

var (
	version     string
	name        string
	environment string
)

const (
	unspecified_version = "unversioned"
	unspecified_env     = "unspecified"
	path_separator      = "/"
)

func GetServiceName() string {
	return name
}

func GetVersion() string {
	return version
}

func GetEnvironment() string {
	return environment
}

func initVersion() {
	if len(version) > 0 {
		return
	}
	version = unspecified_version
}

func initEnvironment() {
	if len(environment) > 0 {
		return
	}
	environment = unspecified_env
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
	initEnvironment()
}
