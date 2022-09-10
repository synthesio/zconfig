package zconfig

import (
	"fmt"
	"reflect"
)

const (
	TagInjectAs    = "inject-as"
	TagInject      = "inject"
	TagKey         = "key"
	TagDefault     = "default"
	TagDescription = "description"
)

type Field struct {
	Value     reflect.Value
	Path      string
	Anonymous bool
	Tags      reflect.StructTag

	Parent   *Field
	Children []*Field

	Key              string
	Provider         string
	Configurable     bool
	ConfigurationKey string

	// InjectionTargets contains a list of values registered for the injection
	// of this field value, once it has been set.
	InjectionTargets []reflect.Value
}

// AddInjectionTarget registers the value of the given target field for injection.
// The method returns an error if the target type is not compatible with the receiver field type or
// if the target value cannot be addressed.
func (f *Field) AddInjectionTarget(target *Field) (err error) {
	if !f.Value.Type().AssignableTo(target.Value.Type()) {
		return fmt.Errorf("cannot inject %s into %s for field %s", f.Value.Type(), target.Value.Type(), target.Path)
	}

	if !target.Value.CanSet() {
		return fmt.Errorf("cannot address %s for injection", target.Value.Type())
	}

	f.InjectionTargets = append(f.InjectionTargets, target.Value)

	return nil
}

func (f *Field) IsLeaf() bool {
	if _, ok := f.Tags.Lookup(TagInject); ok {
		return true
	}

	// The field is a leaf if it is a type different than a struct or a
	// pointer to a struct.
	if f.Value.Kind() != reflect.Ptr {
		return f.Value.Kind() != reflect.Struct
	}

	return f.Value.Type().Elem().Kind() != reflect.Struct
}
