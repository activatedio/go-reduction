package reduction

import (
	"context"
	"time"

	"github.com/activatedio/go-reduction/internal"
)

// Store keeps a state per key and state type. Get answers nil, nil when the
// key holds no state of that type; a found state is a pointer to the state
// type. Set stores state (a pointer) under the key and its own type.
//
// Expiry is the store's own: NewMemoryStore takes one, the Valkey/Redis
// store (github.com/activatedio/go-reduction/store/valkey) has a TTL
// option, and a store over a program's database may keep states for good.
type Store = internal.Access

// TxStore is a Store that can make a read, a reducer and the write of its
// result one unit. When a Reduction's store is a TxStore, Get and Set run
// inside InTx, and the ctx handed to Init, Refresh and the reducers is the
// one InTx passes to fn, so a reducer's own writes and the saved state
// commit or roll back together. InTx returns fn's error.
type TxStore = internal.TxAccess

// NewMemoryStore keeps states in this process for expiration after their
// last write. For tests and a single process.
func NewMemoryStore(expiration time.Duration) Store {
	return internal.NewLocalAccess(&internal.LocalAccessConfig{ExpirationSeconds: int(expiration / time.Second)})
}

// New makes a Reduction over store. The fx module's Factory does the same
// over the in-memory store.
func New(store Store) Reduction {
	return newReduction(reductionParams{Access: store})
}

// WithKey names whose state ctx reads and writes: a person, a session.
// The mux's KeyFunc sets it per request; the cookie middleware sets a
// random one.
func WithKey(ctx context.Context, key string) context.Context {
	return internal.WithSessionID(ctx, key)
}

// KeyFrom answers the key in ctx, if any.
func KeyFrom(ctx context.Context) (string, bool) {
	return internal.SessionID(ctx)
}
