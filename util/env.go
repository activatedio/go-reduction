package util

import (
	"fmt"
	"os"
)

func MustGetEnv(key string) string {
	v := os.Getenv(key)

	if v == "" {
		panic(fmt.Sprintf("environment variable %q not set", key))
	}

	return v
}

func GetEnv(key string, fallback ...string) string {

	v := os.Getenv(key)

	if v == "" && len(fallback) > 0 {
		return fallback[0]
	}

	return v
}
