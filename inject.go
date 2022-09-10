package zconfig

import (
	"context"
)

// Inject is a Hook responsible for injecting source values into their targets.
// It must always be executed after any Repository hooks to ensure that
// source field values have already been populated.
func Inject(ctx context.Context, f *Field) error {
	if len(f.InjectionTargets) == 0 {
		return nil
	}

	for _, target := range f.InjectionTargets {
		target.Set(f.Value)
	}

	return nil
}
