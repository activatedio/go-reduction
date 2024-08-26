package util

import (
	"fmt"
	"os"
	"strconv"
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

// GetEnvBool returns an boolean for an environment variable.
// Valid vlaues are "true" and "false"
// If the environment variable is not found and a fallback is not provided, the call panics
func GetEnvBool(key string, fallback ...bool) bool {

	v := os.Getenv(key)

	if v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			panic(err)
		}
		return b
	} else {
		if len(fallback) > 0 {
			return fallback[0]
		} else {
			panic(fmt.Sprintf("environment variable %q not set and no fallback provided", key))
		}
	}
}
