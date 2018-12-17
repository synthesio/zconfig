package zconfig

import (
	"reflect"
	"strings"

	"github.com/fatih/structtag"
	"github.com/pkg/errors"
)

const (
	tagInjectAs = "inject-as"
	tagInject   = "inject"
	tagKey      = "key"
	tagDefault  = "default"
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

func (f *Field) InjectionSource() (string, bool) {
	if f.Tags == nil {
		return "", false
	}
	tag, err := f.Tags.Get(tagInjectAs)
	if err != nil {
		return "", false
	}
	return tag.Name, true
}

func (f *Field) InjectionTarget() (string, bool) {
	if f.Tags == nil {
		return "", false
	}
	tag, err := f.Tags.Get(tagInject)
	if err != nil {
		return "", false
	}
	return tag.Name, true
}

func (f *Field) Default() (string, bool) {
	if f.Tags == nil {
		return "", false
	}
	tag, err := f.Tags.Get(tagDefault)
	if err != nil {
		return "", false
	}
	return tagValue(tag), true
}

func (f *Field) Inject(s *Field) (err error) {
	if !s.Value.Type().AssignableTo(f.Value.Type()) {
		return errors.Errorf("cannot inject %s into %s for field %s", s.Value.Type(), f.Value.Type(), f.Path)
	}

	if !f.Value.CanSet() {
		return errors.Errorf("cannot address %s for injection", f.Value.Type())
	}

	f.Value.Set(s.Value)
	return nil
}

func (f *Field) IsLeaf() bool {
	if _, isInjected := f.InjectionTarget(); isInjected {
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

// tagValue returns the full, raw value of a tag. Structtag splits a tag value
// into a "name" and "options", separated by commas, but some of our tags do
// not follow this convention (e.g. full-text description, can contain commas).
// This function puts name and options back together.
func tagValue(tag *structtag.Tag) string {
	return strings.Join(append([]string{tag.Name}, tag.Options...), ",")
}
