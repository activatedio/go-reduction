package internal_test

import (
	"context"
	"github.com/activatedio/go-reduction/internal"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestLocalAccess_Get_Set(t *testing.T) {

	unit := internal.NewLocalAccess(&internal.LocalAccessConfig{ExpirationSeconds: 60})

	ctx := context.Background()

	sID1 := "session-id-1"
	sID2 := "session-id-2"

	state1 := &DummyState{
		Value: "1",
	}

	state2 := &DummyState{
		Value: "2",
	}

	got, err := unit.Get(ctx, sID1, reflect.TypeOf(state1))

	assert.Nil(t, got)
	assert.Nil(t, err)

	assert.NoError(t, unit.Set(ctx, sID1, state1))

	got, err = unit.Get(ctx, sID1, reflect.TypeOf(state1))

	assert.Nil(t, err)
	assert.Equal(t, state1, got)

	got, err = unit.Get(ctx, sID2, reflect.TypeOf(state1))

	assert.Nil(t, got)
	assert.Nil(t, err)

	assert.NoError(t, unit.Set(ctx, sID2, state2))

	got, err = unit.Get(ctx, sID1, reflect.TypeOf(state1))

	assert.Nil(t, err)
	assert.Equal(t, state1, got)

	got, err = unit.Get(ctx, sID2, reflect.TypeOf(state2))

	assert.Nil(t, err)
	assert.Equal(t, state2, got)
}
