package reduction

import (
	"context"
	"fmt"
	"github.com/activatedio/reduction/internal"
	"github.com/iancoleman/strcase"
	"reflect"
)

type stateEntry struct {
	access     internal.Access
	t          reflect.Type
	descriptor *StateDescriptor
	get        internal.GetInternal
	actions    map[string]*actionEntry
	init       internal.InitInternal
}

type actionEntry struct {
	set internal.SetInternal
}

func (s *stateEntry) Init(init any) StateBuilder {
	s.init = internal.ToInitInternal(s.access, s.t, init)
	return s
}

func (s *stateEntry) Action(t reflect.Type, reducer any) StateBuilder {

	name := toTypeName(t)

	d := &ActionDescriptor{
		Path: name,
	}

	s.actions[name] = &actionEntry{}

	s.descriptor.Actions = append(s.descriptor.Actions, d)

	return s
}

type reduction struct {
	access internal.Access
	states map[string]*stateEntry
}

func (r *reduction) State(t reflect.Type) StateBuilder {
	name := toTypeName(t)
	e := &stateEntry{
		descriptor: &StateDescriptor{
			Path: fmt.Sprintf("/%s", name),
		},
	}

	r.states[name] = e

	return e
}

func (r *reduction) GetStateDescriptors() []*StateDescriptor {
	var result []*StateDescriptor
	for _, entry := range r.states {
		result = append(result, entry.descriptor)
	}
	return result
}

func (r *reduction) Builder() Builder {
	return r
}

func (r *reduction) Set(ctx context.Context, stateType reflect.Type, action any) SetResult {
	//TODO implement me
	panic("implement me")
}

func (r *reduction) Get(ctx context.Context, stateType reflect.Type) GetResult {
	//TODO implement me
	panic("implement me")
}

func NewReduction() Reduction {
	return &reduction{}
}

func toTypeName(t reflect.Type) string {
	return strcase.ToSnake(t.Name())
}
