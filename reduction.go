package reduction

import (
	"context"
	"fmt"
	"github.com/activatedio/go-reduction/internal"
	"github.com/go-errors/errors"
	"github.com/iancoleman/strcase"
	"go.uber.org/fx"
	"reflect"
)

type stateEntry struct {
	typeName   string
	access     internal.Access
	t          reflect.Type
	descriptor *StateDescriptor
	get        internal.GetInternal
	actions    map[string]*actionEntry
	init       internal.InitInternal
	refresh    internal.RefreshInternal
}

type actionEntry struct {
	set internal.SetInternal
}

func (s *stateEntry) Init(init any) StateBuilder {
	s.init = internal.ToInitInternal(s.access, s.t, init)
	return s
}

func (s *stateEntry) Refresh(init any) StateBuilder {
	s.refresh = internal.ToRefreshInternal(s.access, s.t, init)
	return s
}

func (s *stateEntry) Action(t reflect.Type, reducer any) StateBuilder {

	name := toTypeName(t)

	d := &ActionDescriptor{
		ActionType: t,
		Path:       name,
	}

	s.actions[name] = &actionEntry{
		set: internal.ToSetInternal(s.access, s.t, t, reducer),
	}

	s.descriptor.Actions = append(s.descriptor.Actions, d)

	return s
}

// TODO - write full unit tests for this
type reduction struct {
	access internal.Access
	states map[string]*stateEntry
}

func (r *reduction) State(t reflect.Type) StateBuilder {
	name := toTypeName(t)
	e := &stateEntry{
		typeName: name,
		access:   r.access,
		t:        t,
		descriptor: &StateDescriptor{
			StateType: t,
			Path:      fmt.Sprintf("/%s", name),
			Exporter:  internal.ToExportInternal(t),
		},
		get:     internal.ToGetInternal(r.access, t),
		actions: map[string]*actionEntry{},
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

func (r *reduction) Set(ctx context.Context, stateType reflect.Type, action any) (*SetResult, error) {
	actionName := toTypeName(reflect.TypeOf(reflect.ValueOf(action).Elem().Interface()))
	se, state, err := r.doGet(ctx, stateType)
	if err != nil {
		return nil, err
	}
	ae, ok := se.actions[actionName]
	if !ok {
		return nil, errors.New(fmt.Sprintf("action %s for state %s not found", actionName, se.typeName))
	}
	state, err = ae.set(ctx, state, action)
	if err != nil {
		return nil, err
	}
	return &SetResult{
		State: state,
	}, nil
}

// doGet returns an internal stateEntry for use by tother methods
func (r *reduction) doGet(ctx context.Context, stateType reflect.Type) (*stateEntry, any, error) {
	stateName := toTypeName(stateType)
	se, ok := r.states[stateName]
	if !ok {
		return nil, nil, errors.New(fmt.Sprintf("state %s not found", stateName))
	}
	state, err := se.get(ctx)
	if err != nil {
		return nil, nil, err
	}
	if state == nil {
		if se.init != nil {
			state, err = se.init(ctx)
			if err != nil {
				return nil, nil, err
			}
		} else {
			// TODO - test this
			state = reflect.New(se.t).Interface()
		}
	} else {
		if se.refresh != nil {
			state, err = se.refresh(ctx, state)
			if err != nil {
				return nil, nil, err
			}
		}
	}
	return se, state, nil
}

func (r *reduction) Get(ctx context.Context, stateType reflect.Type) (*GetResult, error) {
	_, state, err := r.doGet(ctx, stateType)
	if err != nil {
		return nil, err
	} else {
		return &GetResult{
			State: state,
		}, nil
	}
}

type reductionParams struct {
	Access internal.Access
}

func newReduction(params reductionParams) Reduction {
	return &reduction{
		access: params.Access,
		states: map[string]*stateEntry{},
	}
}

type factory struct {
	access internal.Access
}

func (f *factory) NewReduction() Reduction {
	return &reduction{
		access: f.access,
		states: map[string]*stateEntry{},
	}
}

type FactoryParams struct {
	fx.In
	Access internal.Access
}

func NewFactory(params FactoryParams) Factory {
	return &factory{
		access: params.Access,
	}
}

func toTypeName(t reflect.Type) string {
	return strcase.ToSnake(t.Name())
}
