package reduction

import (
	"context"
	"reflect"
)

type Reducer[S any, A any] func(ctx context.Context, state *S, action *A) (*S, error)

type StateFactory[S any] func(ctx context.Context) (*S, error)

type SetResult struct {
	State any
}

type GetResult struct {
	State any
}

type StateBuilder interface {
	Init(init any) StateBuilder
	Action(t reflect.Type, reducer any) StateBuilder
}

type Builder interface {
	State(t reflect.Type) StateBuilder
}

type ActionDescriptor struct {
	Path string
}

type StateDescriptor struct {
	Path    string
	Actions []*ActionDescriptor
}

type Reduction interface {
	Builder() Builder
	GetStateDescriptors() []*StateDescriptor
	Set(ctx context.Context, stateType reflect.Type, action any) SetResult
	Get(ctx context.Context, stateType reflect.Type) GetResult
}
