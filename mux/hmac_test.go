package mux_test

import (
	"github.com/activatedio/go-reduction/mux"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHMAC(t *testing.T) {

	key := "some-secret"

	unit := mux.NewHMAC(&mux.HMACConfig{Key: key})

	// No signature
	valid, payload := unit.ValidateAndExtract("payload")
	assert.False(t, valid)
	assert.Empty(t, payload)

	// Invalid signature
	valid, payload = unit.ValidateAndExtract("payload.invalid-sig")
	assert.False(t, valid)
	assert.Empty(t, payload)

	input := "test-input"
	result := unit.Sign(input)

	valid, payload = unit.ValidateAndExtract(result)
	assert.True(t, valid)
	assert.Equal(t, input, payload)

}
