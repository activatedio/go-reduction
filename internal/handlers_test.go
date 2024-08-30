package internal_test

import (
	"context"
	"github.com/activatedio/reduction/internal"
	"github.com/activatedio/reduction/internal/mock_internal"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestToInitInternal(t *testing.T) {

	type s struct {
		arrange func() (reflect.Type, any)
		assert  func(got internal.InitInternal)
	}

	var aMock *mock_internal.Access

	state := &DummyState{
		Value: "value1",
	}

	refCtx := context.Background()
	refErr := errors.New("test")

	cases := map[string]s{
		"success": {
			arrange: func() (reflect.Type, any) {
				return reflect.TypeFor[DummyState](), func(ctx context.Context) (*DummyState, error) {
					assert.Equal(t, refCtx, ctx)
					return state, nil
				}
			},
			assert: func(got internal.InitInternal) {
				result, err := got(refCtx)
				assert.Nil(t, err)
				assert.Same(t, state, result)
			},
		},
		"error": {
			arrange: func() (reflect.Type, any) {
				return reflect.TypeFor[DummyState](), func(ctx context.Context) (*DummyState, error) {
					assert.Equal(t, refCtx, ctx)
					return nil, refErr
				}
			},
			assert: func(got internal.InitInternal) {
				result, err := got(refCtx)
				assert.Nil(t, result)
				assert.Equal(t, refErr, err)
			},
		},
	}

	for k, v := range cases {
		t.Run(k, func(t *testing.T) {

			aMock = mock_internal.NewAccess(t)
			st, i := v.arrange()
			v.assert(internal.ToInitInternal(aMock, st, i))
		})
	}
}

func TestToGetInternal(t *testing.T) {

	type s struct {
		arrange func(ctx context.Context) (context.Context, reflect.Type)
		assert  func(ctx context.Context, got internal.GetInternal)
	}

	var aMock *mock_internal.Access

	refErr := errors.New("test")
	state := &DummyState{
		Value: "value1",
	}
	stateType := reflect.TypeOf(state)
	sessionID := "test-session-id"

	cases := map[string]s{
		"success": {
			arrange: func(ctx context.Context) (context.Context, reflect.Type) {
				ctx = internal.WithSessionID(ctx, sessionID)
				aMock.EXPECT().Get(ctx, sessionID, stateType).Return(state, nil)
				return ctx, stateType
			},
			assert: func(ctx context.Context, got internal.GetInternal) {
				result, err := got(ctx)
				assert.Nil(t, err)
				assert.Same(t, state, result)
			},
		},
		"error": {
			arrange: func(ctx context.Context) (context.Context, reflect.Type) {
				ctx = internal.WithSessionID(ctx, sessionID)
				aMock.EXPECT().Get(ctx, sessionID, stateType).Return(nil, refErr)
				return ctx, stateType
			},
			assert: func(ctx context.Context, got internal.GetInternal) {
				result, err := got(ctx)
				assert.Nil(t, result)
				assert.Equal(t, refErr, err)
			},
		},
	}

	for k, v := range cases {
		t.Run(k, func(t *testing.T) {

			aMock = mock_internal.NewAccess(t)
			ctx, st := v.arrange(context.Background())
			v.assert(ctx, internal.ToGetInternal(aMock, st))

		})
	}

}

func TestToSetInternal(t *testing.T) {

	type s struct {
		arrange func(ctx context.Context) (context.Context, reflect.Type, reflect.Type, any)
		assert  func(ctx context.Context, got internal.SetInternal)
	}

	var aMock *mock_internal.Access

	refErr := errors.New("test")
	state := &DummyState{
		Value: "value1",
	}
	reducedState := &DummyState{
		Value: "value2",
	}
	stateType := reflect.TypeOf(state)
	action := &DummyAction{
		Action: "action1",
	}
	actionType := reflect.TypeOf(action)
	sessionID := "test-session-id"

	makeReducer := func(err error) func(ctx context.Context, state *DummyState, action *DummyAction) (*DummyState, error) {
		return func(ctx context.Context, state *DummyState, action *DummyAction) (*DummyState, error) {
			if err != nil {
				return nil, err
			} else {
				return reducedState, err
			}
		}
	}

	cases := map[string]s{
		"success": {
			arrange: func(ctx context.Context) (context.Context, reflect.Type, reflect.Type, any) {
				ctx = internal.WithSessionID(ctx, sessionID)
				aMock.EXPECT().Set(ctx, sessionID, reducedState).Return(nil)
				return ctx, stateType, actionType, makeReducer(nil)
			},
			assert: func(ctx context.Context, got internal.SetInternal) {
				result, err := got(ctx, state, action)
				assert.Nil(t, err)
				assert.Equal(t, reducedState, result)
			},
		},
		"reducer error": {
			arrange: func(ctx context.Context) (context.Context, reflect.Type, reflect.Type, any) {
				ctx = internal.WithSessionID(ctx, sessionID)
				return ctx, stateType, actionType, makeReducer(refErr)
			},
			assert: func(ctx context.Context, got internal.SetInternal) {
				result, err := got(ctx, state, action)
				assert.Nil(t, result)
				assert.Equal(t, refErr, err)
			},
		},
		"access error": {
			arrange: func(ctx context.Context) (context.Context, reflect.Type, reflect.Type, any) {
				ctx = internal.WithSessionID(ctx, sessionID)
				aMock.EXPECT().Set(ctx, sessionID, reducedState).Return(refErr)
				return ctx, stateType, actionType, makeReducer(nil)
			},
			assert: func(ctx context.Context, got internal.SetInternal) {
				result, err := got(ctx, state, action)
				assert.Nil(t, result)
				assert.Equal(t, refErr, err)
			},
		},
	}

	for k, v := range cases {

		t.Run(k, func(t *testing.T) {

			aMock = mock_internal.NewAccess(t)
			ctx, st, at, r := v.arrange(context.Background())
			v.assert(ctx, internal.ToSetInternal(aMock, st, at, r))
		})
	}
}

func TestToRefreshInternal(t *testing.T) {

	type s struct {
		arrange func(ctx context.Context) (context.Context, reflect.Type, any)
		assert  func(ctx context.Context, got internal.RefreshInternal)
	}

	var aMock *mock_internal.Access

	refErr := errors.New("test")
	state := &DummyState{
		Value: "value1",
	}
	reducedState := &DummyState{
		Value: "value2",
	}
	stateType := reflect.TypeOf(state)
	sessionID := "test-session-id"

	makeRefresher := func(err error) func(ctx context.Context, state *DummyState) (*DummyState, error) {
		return func(ctx context.Context, state *DummyState) (*DummyState, error) {
			if err != nil {
				return nil, err
			} else {
				return reducedState, err
			}
		}
	}

	cases := map[string]s{
		"success": {
			arrange: func(ctx context.Context) (context.Context, reflect.Type, any) {
				ctx = internal.WithSessionID(ctx, sessionID)
				aMock.EXPECT().Set(ctx, sessionID, reducedState).Return(nil)
				return ctx, stateType, makeRefresher(nil)
			},
			assert: func(ctx context.Context, got internal.RefreshInternal) {
				result, err := got(ctx, state)
				assert.Nil(t, err)
				assert.Equal(t, reducedState, result)
			},
		},
		"refresher error": {
			arrange: func(ctx context.Context) (context.Context, reflect.Type, any) {
				ctx = internal.WithSessionID(ctx, sessionID)
				return ctx, stateType, makeRefresher(refErr)
			},
			assert: func(ctx context.Context, got internal.RefreshInternal) {
				result, err := got(ctx, state)
				assert.Nil(t, result)
				assert.Equal(t, refErr, err)
			},
		},
		"access error": {
			arrange: func(ctx context.Context) (context.Context, reflect.Type, any) {
				ctx = internal.WithSessionID(ctx, sessionID)
				aMock.EXPECT().Set(ctx, sessionID, reducedState).Return(refErr)
				return ctx, stateType, makeRefresher(nil)
			},
			assert: func(ctx context.Context, got internal.RefreshInternal) {
				result, err := got(ctx, state)
				assert.Nil(t, result)
				assert.Equal(t, refErr, err)
			},
		},
	}

	for k, v := range cases {

		t.Run(k, func(t *testing.T) {

			aMock = mock_internal.NewAccess(t)
			ctx, st, r := v.arrange(context.Background())
			v.assert(ctx, internal.ToRefreshInternal(aMock, st, r))
		})
	}
}
