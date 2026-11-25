package mux

import "net/http"

type Error struct {
	Error string `json:"error,omitempty"`
	// Field names the action's field a 400 is about, when it is about one.
	Field string `json:"field,omitempty"`
	// Ref is a server fault's reference, under which its detail is logged.
	Ref string `json:"ref,omitempty"`
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
