package mux

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
)

type hmacImpl struct {
	keyString string
}

func (h *hmacImpl) Sign(input string) string {
	return fmt.Sprintf("%s.%s", input, h.computeHmac(input))
}

func (h *hmacImpl) ValidateAndExtract(input string) (bool, string) {
	parts := strings.Split(input, ".")
	if len(parts) < 2 {
		return false, ""
	}
	inputSig := parts[len(parts)-1]
	payload := input[:len(input)-len(inputSig)-1]
	sig := h.computeHmac(payload)
	if inputSig != sig {
		return false, ""
	} else {
		return true, payload
	}
}

func (h *hmacImpl) computeHmac(input string) string {
	key := hmac.New(sha256.New, []byte(h.keyString))
	key.Write([]byte(input))
	return base64.RawURLEncoding.EncodeToString(key.Sum(nil))
}

type HMACConfig struct {
	Key string
}

func NewHMAC(config *HMACConfig) HMAC {

	return &hmacImpl{
		keyString: config.Key,
	}
}
