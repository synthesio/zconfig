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
	TagExample     = "example"
)

type Field struct {
	Value     reflect.Value
	Path      string
	Anonymous bool
	Tags      reflect.StructTag

	Parent   *Field
	Children []*Field

	Key              string
	Configurable     bool
	ConfigurationKey string
}

func (f *Field) Inject(s *Field) (err error) {
	if !s.Value.Type().AssignableTo(f.Value.Type()) {
		return fmt.Errorf("cannot inject %s into %s for field %s", s.Value.Type(), f.Value.Type(), f.Path)
	}

	if !f.Value.CanSet() {
		return fmt.Errorf("cannot address %s for injection", f.Value.Type())
	}

	f.Value.Set(s.Value)
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
