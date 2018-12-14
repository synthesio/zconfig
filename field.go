package zconfig

import (
	"reflect"

	"github.com/fatih/structtag"
	"github.com/pkg/errors"
)

const (
	tagInjectAs = "inject-as"
	tagInject   = "inject"
)

type Field struct {
	Value       reflect.Value
	StructField *reflect.StructField
	Parent      *Field
	Children    []*Field

	Path Path
	Tags *structtag.Tags
}

func (f *Field) InjectionSource() (InjectionKey, bool) {
	if f.Tags == nil {
		return "", false
	}
	tag, err := f.Tags.Get(tagInjectAs)
	if err != nil {
		return "", false
	}
	return InjectionKey(tag.Name), true
}

func (f *Field) InjectionTarget() (InjectionKey, bool) {
	if f.Tags == nil {
		return "", false
	}
	tag, err := f.Tags.Get(tagInject)
	if err != nil {
		return "", false
	}
	return InjectionKey(tag.Name), true
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
