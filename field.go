package zconfig

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/fatih/structtag"
)

const (
	TagInjectAs    = "inject-as"
	TagInject      = "inject"
	TagKey         = "key"
	TagDefault     = "default"
	TagDescription = "description"
)

type Field struct {
	StructField *reflect.StructField
	Value       reflect.Value
	Path        string

	Parent   *Field
	Children []*Field

	Tags             *structtag.Tags
	Key              string
	Configurable     bool
	ConfigurationKey string
}

func (f *Field) Tag(name string) (string, bool) {
	if f.Tags == nil {
		return "", false
	}
	tag, err := f.Tags.Get(name)
	if err != nil {
		return "", false
	}
	return tag.Name, true
}

func (f *Field) FullTag(name string) (string, bool) {
	if f.Tags == nil {
		return "", false
	}
	tag, err := f.Tags.Get(name)
	if err != nil {
		return "", false
	}
	return strings.Join(append([]string{tag.Name}, tag.Options...), ","), true
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
	if _, isInjected := f.Tag(TagInject); isInjected {
		return true
	}

	// The field is a leaf if it is a type different than a struct or a
	// pointer to a struct.
	if f.Value.Kind() != reflect.Ptr {
		return f.Value.Kind() != reflect.Struct
	}

	return f.Value.Type().Elem().Kind() != reflect.Struct
}

func (f *Field) IsAnonymous() bool {
	if f.StructField == nil {
		return true
	}

	return f.StructField.Anonymous
}
