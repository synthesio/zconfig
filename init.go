package zconfig

import (
	"context"
	"fmt"
	"reflect"
)

type Initializable interface {
	Init() error
}

type InitializableEx interface {
	Init(context.Context) error
}

// Used for type comparison.
var typeInitializable = reflect.TypeOf((*Initializable)(nil)).Elem()
var typeInitializableEx = reflect.TypeOf((*InitializableEx)(nil)).Elem()

// DEPRECATED
func Initialize(field *Field) error {
	// Not initializable, nothing to do.
	if !field.Value.Type().Implements(typeInitializable) {
		return nil
	}

	// Initialize the element itself via the interface.
	err := field.Value.Interface().(Initializable).Init()
	if err != nil {
		return fmt.Errorf("initializing field: %s", err)
	}

	return nil
}

func InitializeEx(ctx context.Context, field *Field) error {
	var err error
	if field.Value.Type().Implements(typeInitializableEx) {
		err = field.Value.Interface().(InitializableEx).Init(ctx)
	} else if field.Value.Type().Implements(typeInitializable) {
		// Initialize the element itself via the interface.
		err = field.Value.Interface().(Initializable).Init()
	}
	if err != nil {
		return fmt.Errorf("initializing field: %s", err)
	}

	return nil
}
