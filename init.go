package zconfig

import (
	"context"
	"fmt"
	"reflect"
)

type Initializable interface {
	Init(context.Context) error
}

type initializableDeprecated interface {
	Init() error
}

// Used for type comparison.
var typeInitializable = reflect.TypeOf((*Initializable)(nil)).Elem()
var typeInitializableDeprecated = reflect.TypeOf((*initializableDeprecated)(nil)).Elem()

func Initialize(ctx context.Context, field *Field) error {
	// Not initializable, nothing to do.
	if field.Value.Type().Implements(typeInitializable) {

		// Initialize the element itself via the interface.
		err := field.Value.Interface().(Initializable).Init(ctx)
		if err != nil {
			return fmt.Errorf("initializing field: %s", err)
		}

		return nil
	}

	if field.Value.Type().Implements(typeInitializableDeprecated) {

		// Initialize the element itself via the interface.
		err := field.Value.Interface().(initializableDeprecated).Init()
		if err != nil {
			return fmt.Errorf("initializing field: %s", err)
		}

		return nil
	}

	return nil
}
