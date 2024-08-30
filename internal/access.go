package internal

import (
	"context"
	"fmt"
	"github.com/activatedio/go-reduction/config"
	"github.com/activatedio/go-reduction/util"
	"github.com/patrickmn/go-cache"
	"reflect"
	"strconv"
	"time"
)

type localAccess struct {
	cache *cache.Cache
}

func (l *localAccess) Get(ctx context.Context, sessionKey string, stateType reflect.Type) (any, error) {
	key := makeKey(sessionKey, stateType)
	got, ok := l.cache.Get(key)
	if ok {
		return got, nil
	} else {
		return nil, nil
	}
}

func (l *localAccess) Set(ctx context.Context, sessionKey string, state any) error {
	key := makeKey(sessionKey, reflect.TypeOf(state))
	l.cache.Set(key, state, cache.DefaultExpiration)
	return nil
}

// NewLocalAccess creates an access implementation which is appropriate for single node deployments
func NewLocalAccess(config *LocalAccessConfig) Access {
	return &localAccess{
		cache: cache.New(time.Duration(config.ExpirationSeconds)*time.Second, 60*time.Second),
	}
}

type LocalAccessConfig struct {
	ExpirationSeconds int
}

func NewLocalAccessConfig() *LocalAccessConfig {
	expirationStr := util.GetEnv(config.ReductionKeyLocalAccessExpirationSeconds, "1200")
	expirationSeconds, err := strconv.Atoi(expirationStr)
	if err != nil {
		panic(err)
	}
	return &LocalAccessConfig{
		ExpirationSeconds: expirationSeconds,
	}
}

func makeKey(sessionID string, stateType reflect.Type) string {
	return fmt.Sprintf("%s_%s", sessionID, stateType.Name())
}
