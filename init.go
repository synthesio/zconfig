package zconfig

import (
	"context"
	"fmt"
	"reflect"
)

type Initializable interface {
	Init(context.Context) error
}

// Used for type comparison.
var typeInitializable = reflect.TypeOf((*Initializable)(nil)).Elem()

func Initialize(ctx context.Context, field *Field) error {
	// Not initializable, nothing to do.
	if !field.Value.Type().Implements(typeInitializable) {
		return nil
	}

	// Initialize the element itself via the interface.
	err := field.Value.Interface().(Initializable).Init(ctx)
	if err != nil {
		return fmt.Errorf("initializing field: %s", err)
	}

	return nil
}
