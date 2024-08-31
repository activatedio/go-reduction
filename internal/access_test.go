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

	stateType := reflect.TypeFor[DummyState]()

	state1 := &DummyState{
		Value: "1",
	}

	state2 := &DummyState{
		Value: "2",
	}

	got, err := unit.Get(ctx, sID1, stateType)

	assert.Nil(t, got)
	assert.Nil(t, err)

	assert.NoError(t, unit.Set(ctx, sID1, state1))

	got, err = unit.Get(ctx, sID1, stateType)

	assert.Nil(t, err)
	assert.Equal(t, state1, got)

	got, err = unit.Get(ctx, sID2, stateType)

	assert.Nil(t, got)
	assert.Nil(t, err)

	assert.NoError(t, unit.Set(ctx, sID2, state2))

	got, err = unit.Get(ctx, sID1, stateType)

	assert.Nil(t, err)
	assert.Equal(t, state1, got)

	got, err = unit.Get(ctx, sID2, stateType)

	assert.Nil(t, err)
	assert.Equal(t, state2, got)
}
