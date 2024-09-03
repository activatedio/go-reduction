package internal_test

import "context"

type DummyState struct {
	Value string
}

type DummyAction struct {
	Action string
}

type DummyExportableState struct {
	Value string
}

func (d *DummyExportableState) Export(ctx context.Context) (*DummyExportedState, error) {

	return &DummyExportedState{Value: d.Value + "-exported"}, nil
}

type DummyExportedState struct {
	Value string
}
