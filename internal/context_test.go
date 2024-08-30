package internal_test

import (
	"context"
	"github.com/activatedio/go-reduction/internal"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWithSessionID_MustGetSessionID(t *testing.T) {

	ctx := context.Background()
	sID := "test-session-id"

	assert.Equal(t, sID, internal.MustGetSessionID(internal.WithSessionID(ctx, sID)))

}
