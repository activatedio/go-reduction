package valkey_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/activatedio/go-reduction/store/valkey"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// store is reduction.Store's method set. This module does not import the
// library, so a program can take the store without a version tie.
type store interface {
	Get(ctx context.Context, sessionKey string, stateType reflect.Type) (any, error)
	Set(ctx context.Context, sessionKey string, state any) error
}

var _ store = (*valkey.Store)(nil)

type Cart struct {
	Items []string `json:"items"`
	Total int      `json:"total"`
}

type Other struct {
	Step string
}

func newStore(t *testing.T, opts ...valkey.Option) (*valkey.Store, *miniredis.Miniredis) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return valkey.New(client, opts...), mr
}

func TestStore_GetSet(t *testing.T) {
	s, _ := newStore(t)
	ctx := context.Background()
	cart := reflect.TypeFor[Cart]()

	got, err := s.Get(ctx, "alice", cart)
	require.NoError(t, err)
	assert.Nil(t, got)

	require.NoError(t, s.Set(ctx, "alice", &Cart{Items: []string{"a"}, Total: 3}))
	require.NoError(t, s.Set(ctx, "alice", &Other{Step: "two"}))

	got, err = s.Get(ctx, "alice", cart)
	require.NoError(t, err)
	assert.Equal(t, &Cart{Items: []string{"a"}, Total: 3}, got)

	got, err = s.Get(ctx, "alice", reflect.TypeFor[Other]())
	require.NoError(t, err)
	assert.Equal(t, &Other{Step: "two"}, got, "each state type has its own entry")

	got, err = s.Get(ctx, "bob", cart)
	require.NoError(t, err)
	assert.Nil(t, got, "each key has its own entry")
}

func TestStore_TTL(t *testing.T) {
	s, mr := newStore(t, valkey.WithTTL(time.Minute))
	ctx := context.Background()
	require.NoError(t, s.Set(ctx, "alice", &Cart{Total: 1}))
	assert.Equal(t, time.Minute, mr.TTL("reduction:alice:Cart"))

	mr.FastForward(2 * time.Minute)
	got, err := s.Get(ctx, "alice", reflect.TypeFor[Cart]())
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestStore_NoTTL(t *testing.T) {
	s, mr := newStore(t, valkey.WithTTL(0))
	require.NoError(t, s.Set(context.Background(), "alice", &Cart{}))
	assert.Zero(t, mr.TTL("reduction:alice:Cart"))
}

func TestStore_Prefix(t *testing.T) {
	s, mr := newStore(t, valkey.WithPrefix("portal:gs:"))
	require.NoError(t, s.Set(context.Background(), "alice", &Cart{Total: 2}))
	assert.True(t, mr.Exists("portal:gs:alice:Cart"))
	assert.False(t, mr.Exists("reduction:alice:Cart"))
}

func TestStore_RefusesNonPointer(t *testing.T) {
	s, _ := newStore(t)
	assert.Error(t, s.Set(context.Background(), "alice", Cart{}))
	var nilCart *Cart
	assert.Error(t, s.Set(context.Background(), "alice", nilCart))
}

func TestStore_UnreadableValue(t *testing.T) {
	s, mr := newStore(t)
	require.NoError(t, mr.Set("reduction:alice:Cart", "not json"))
	_, err := s.Get(context.Background(), "alice", reflect.TypeFor[Cart]())
	assert.Error(t, err)
}
