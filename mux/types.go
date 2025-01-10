package mux

import "net/http"

type Error struct {
	Error string `json:"error,omitempty"`
}

type Middleware interface {
	Handle(next http.Handler) http.Handler
}

type HMAC interface {
	// Takes an input, calculate the hmac signature and return with signature appended
	Sign(input string) string
	// Takes a signed string, checks the hmac signature, and returns
	ValidateAndExtract(input string) (bool, string)
}
