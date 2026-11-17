// Package valkey is a reduction.Store over Valkey or Redis (anything that
// speaks the Redis protocol), for a program that runs more than one
// process and already has one. It is its own module, so a program that
// does not use it does not pull the client in.
//
// States are kept as JSON under <prefix><key>:<state type name>, each with
// the store's TTL from its last write.
package valkey

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/redis/go-redis/v9"
)

// DefaultTTL is how long a state is kept after its last write unless WithTTL says otherwise.
const DefaultTTL = 20 * time.Minute

// DefaultPrefix starts every key unless WithPrefix says otherwise.
const DefaultPrefix = "reduction:"

// Store implements reduction.Store.
type Store struct {
	client redis.UniversalClient
	prefix string
	ttl    time.Duration
}

// Option configures a Store.
type Option func(*Store)

// WithTTL keeps a state ttl after its last write; 0 keeps it for good.
func WithTTL(ttl time.Duration) Option {
	return func(s *Store) { s.ttl = ttl }
}

// WithPrefix starts every key with prefix, to share a server with others.
func WithPrefix(prefix string) Option {
	return func(s *Store) { s.prefix = prefix }
}

// New makes a Store over client: a *redis.Client, a cluster or a failover client.
func New(client redis.UniversalClient, opts ...Option) *Store {
	s := &Store{client: client, prefix: DefaultPrefix, ttl: DefaultTTL}
	for _, o := range opts {
		o(s)
	}
	return s
}

func (s *Store) key(sessionKey string, stateType reflect.Type) string {
	return fmt.Sprintf("%s%s:%s", s.prefix, sessionKey, stateType.Name())
}

// Get answers a pointer to a stateType read from the server, or nil, nil when there is none.
func (s *Store) Get(ctx context.Context, sessionKey string, stateType reflect.Type) (any, error) {
	b, err := s.client.Get(ctx, s.key(sessionKey, stateType)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	state := reflect.New(stateType).Interface()
	if err := json.Unmarshal(b, state); err != nil {
		return nil, fmt.Errorf("reading the %s state: %w", stateType.Name(), err)
	}
	return state, nil
}

// Set writes state (a pointer to a state) under the key and its type.
func (s *Store) Set(ctx context.Context, sessionKey string, state any) error {
	v := reflect.ValueOf(state)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return fmt.Errorf("a state is a non-nil pointer, not %T", state)
	}
	b, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, s.key(sessionKey, v.Elem().Type()), b, s.ttl).Err()
}
