package reduction_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/activatedio/go-reduction"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Counter struct {
	N int
}

type Add struct {
	By int
}

type Fail struct {
	reduction.Empty
}

func newCounter(store reduction.Store, fail error) reduction.Reduction {
	r := reduction.New(store)
	r.Builder().State(reflect.TypeFor[Counter]()).
		Action(reflect.TypeFor[Add](), func(ctx context.Context, s *Counter, a *Add) (*Counter, error) {
			if a.By <= 0 {
				return nil, reduction.Invalid("by", "must be more than 0")
			}
			return &Counter{N: s.N + a.By}, nil
		}).
		Action(reflect.TypeFor[Fail](), func(ctx context.Context, s *Counter, a *Fail) (*Counter, error) {
			return nil, fail
		})
	return r
}

func TestNew_MemoryStore_KeysAreSeparate(t *testing.T) {
	r := newCounter(reduction.NewMemoryStore(time.Minute), nil)
	alice := reduction.WithKey(context.Background(), "alice")
	bob := reduction.WithKey(context.Background(), "bob")

	_, err := r.Set(alice, reflect.TypeFor[Counter](), &Add{By: 2})
	require.NoError(t, err)
	_, err = r.Set(alice, reflect.TypeFor[Counter](), &Add{By: 3})
	require.NoError(t, err)

	got, err := r.Get(alice, reflect.TypeFor[Counter]())
	require.NoError(t, err)
	assert.Equal(t, &Counter{N: 5}, got.State)

	got, err = r.Get(bob, reflect.TypeFor[Counter]())
	require.NoError(t, err)
	assert.Equal(t, &Counter{}, got.State)
}

func TestNoKey(t *testing.T) {
	r := newCounter(reduction.NewMemoryStore(time.Minute), nil)
	_, err := r.Get(context.Background(), reflect.TypeFor[Counter]())
	assert.ErrorIs(t, err, reduction.ErrNoKey)
	_, err = r.Set(reduction.WithKey(context.Background(), ""), reflect.TypeFor[Counter](), &Add{By: 1})
	assert.ErrorIs(t, err, reduction.ErrNoKey)
}

func TestKeyFrom(t *testing.T) {
	_, ok := reduction.KeyFrom(context.Background())
	assert.False(t, ok)
	k, ok := reduction.KeyFrom(reduction.WithKey(context.Background(), "k"))
	assert.True(t, ok)
	assert.Equal(t, "k", k)
}

func TestErrors(t *testing.T) {
	var se reduction.StatusError

	err := fmt.Errorf("wrapped: %w", reduction.Invalid("name", "is required"))
	require.True(t, errors.As(err, &se))
	assert.Equal(t, http.StatusBadRequest, se.HTTPStatus())
	assert.Equal(t, "name: is required", se.Error())
	assert.Equal(t, "whole", reduction.Invalid("", "whole").Error())

	require.True(t, errors.As(reduction.NotAllowed("no"), &se))
	assert.Equal(t, http.StatusForbidden, se.HTTPStatus())

	require.True(t, errors.As(reduction.ErrNoKey, &se))
	assert.Equal(t, http.StatusUnauthorized, se.HTTPStatus())
}

// txStore is a memory store with a transaction: writes inside InTx are
// held and kept only when fn succeeds. It records what ran inside one.
type txStore struct {
	reduction.Store
	inTx    bool
	pending map[string]any
	calls   int
	setsIn  int
	setsOut int
}

type txKey struct{}

func (s *txStore) InTx(ctx context.Context, fn func(ctx context.Context) error) error {
	s.calls++
	s.inTx = true
	s.pending = map[string]any{}
	defer func() { s.inTx = false }()
	if err := fn(context.WithValue(ctx, txKey{}, true)); err != nil {
		return err // rolled back: pending dropped
	}
	for k, v := range s.pending {
		if err := s.Store.Set(ctx, k, v); err != nil {
			return err
		}
	}
	return nil
}

func (s *txStore) Set(ctx context.Context, key string, state any) error {
	if ctx.Value(txKey{}) == nil {
		s.setsOut++
		return s.Store.Set(ctx, key, state)
	}
	s.setsIn++
	s.pending[key] = state
	return nil
}

func TestTxStore_ReducerAndWriteAreOneUnit(t *testing.T) {
	store := &txStore{Store: reduction.NewMemoryStore(time.Minute)}
	var sawTx bool
	r := reduction.New(store)
	r.Builder().State(reflect.TypeFor[Counter]()).
		Action(reflect.TypeFor[Add](), func(ctx context.Context, s *Counter, a *Add) (*Counter, error) {
			sawTx = ctx.Value(txKey{}) != nil
			return &Counter{N: s.N + a.By}, nil
		})
	ctx := reduction.WithKey(context.Background(), "k")

	_, err := r.Set(ctx, reflect.TypeFor[Counter](), &Add{By: 4})
	require.NoError(t, err)
	assert.True(t, sawTx, "the reducer runs inside InTx")
	assert.Equal(t, 1, store.calls)
	assert.Equal(t, 1, store.setsIn)
	assert.Zero(t, store.setsOut)

	got, err := r.Get(ctx, reflect.TypeFor[Counter]())
	require.NoError(t, err)
	assert.Equal(t, &Counter{N: 4}, got.State)
	assert.Equal(t, 2, store.calls, "Get runs inside InTx too")
}

func TestTxStore_FailedReducerKeepsNothing(t *testing.T) {
	store := &txStore{Store: reduction.NewMemoryStore(time.Minute)}
	boom := errors.New("boom")
	r := newCounter(store, boom)
	ctx := reduction.WithKey(context.Background(), "k")

	_, err := r.Set(ctx, reflect.TypeFor[Counter](), &Add{By: 1})
	require.NoError(t, err)
	_, err = r.Set(ctx, reflect.TypeFor[Counter](), &Fail{})
	assert.ErrorIs(t, err, boom)

	got, err := r.Get(ctx, reflect.TypeFor[Counter]())
	require.NoError(t, err)
	assert.Equal(t, &Counter{N: 1}, got.State)
}

func TestFieldErrorFromReducer(t *testing.T) {
	r := newCounter(reduction.NewMemoryStore(time.Minute), nil)
	_, err := r.Set(reduction.WithKey(context.Background(), "k"), reflect.TypeFor[Counter](), &Add{By: 0})
	var fe *reduction.FieldError
	require.True(t, errors.As(err, &fe))
	assert.Equal(t, "by", fe.Field)
}
