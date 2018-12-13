package zconfig

import (
	"fmt"
	"reflect"

	"github.com/fatih/structtag"
	"github.com/pkg/errors"
)

func Configure(s interface{}) error {
	v := reflect.ValueOf(s)

	if v.Kind() != reflect.Ptr {
		return errors.Errorf("expected pointer to struct, %T given", s)
	}

	if v.Elem().Kind() != reflect.Struct {
		return errors.Errorf("expected pointer to struct, %T given", s)
	}

	root, err := walk(v, reflect.StructField{}, nil)
	if err != nil {
		return errors.Wrap(err, "walking struct")
	}

	fields, err := resolve(root)
	if err != nil {
		return errors.Wrap(err, "resolving struct")
	}

	_ = fields

	return nil
}

type (
	Path         string
	Key          string
	InjectionKey string
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
	return nil
}

func walk(v reflect.Value, s reflect.StructField, p *Field) (field *Field, err error) {
	field = &Field{
		Value:  v,
		Parent: p,
	}

	if p == nil {
		field.Path = "root"
	} else {
		field.StructField = &s
		field.Path = Path(fmt.Sprintf("%s.%s", p.Path, s.Name))
		field.Tags, err = structtag.Parse(string(s.Tag))
		if err != nil {
			return nil, errors.Wrapf(err, "invalid tag for field %s", field.Path)
		}
	}

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return field, nil
	}

	for i := 0; i < v.Type().NumField(); i++ {
		child, err := walk(v.Field(i), v.Type().Field(i), field)
		if err != nil {
			return nil, err
		}

		field.Children = append(field.Children, child)
	}

	return field, nil
}

func resolve(root *Field) (fields []*Field, err error) {
	// Do a stack-based depth-first-search to retrieve all fields, and
	// identify the dependencies of the various fields.
	var (
		stack        = []*Field{root}
		paths        = make(map[Path]*Field)
		sources      = make(map[InjectionKey]*Field)
		targets      = make(map[*Field]InjectionKey)
		dependencies = make(map[Path][]*Field)
	)
	for len(stack) != 0 {
		e := stack[len(stack)-1]
		stack = append(stack[:len(stack)-1], e.Children...)

		paths[e.Path] = e

		if key, ok := e.InjectionSource(); ok {
			if s, ok := sources[key]; ok {
				return nil, errors.Errorf("injection source key %s already defined at path %s", key, s.Path)
			}
			sources[key] = e
		}

		if key, ok := e.InjectionTarget(); ok {
			targets[e] = key
		}

		// Safeguard against children and/or dependency modification
		// later in the process by copying the slice right now.
		dependencies[e.Path] = append([]*Field(nil), e.Children...)
	}

	// Inject the sources into the targets and add them to the dependencies
	// of the targets.
	for target, key := range targets {
		source, ok := sources[key]
		if !ok {
			return nil, errors.Errorf("injection source key %s undefined for path %s", key, target.Path)
		}

		err := target.Inject(source)
		if err != nil {
			return nil, errors.Wrapf(err, "injecting field %s into %s", source.Path, target.Path)
		}

		dependencies[target.Path] = append(dependencies[target.Path], source)
	}

	// Resolve the dependency graph by finding fields that have no
	// dependency and removing them from the graph and the dependencies of
	// the other fields. Iterate until the graph is empty, in which case we
	// obtain a resolved set of fields.
	for len(dependencies) != 0 {
		var resolved = make([]*Field, 0)
		for path, deps := range dependencies {
			if len(deps) != 0 {
				continue
			}

			resolved = append(resolved, paths[path])
			delete(dependencies, path)

			// Remove the field from the other fields dependencies
			// list. We find the index in the list and remove this
			// index only (a field can be a dependency of another
			// only once).
			for fieldPath, deps := range dependencies {
				var idx int = -1
				for i, dep := range deps {
					if dep.Path == path {
						idx = i
						break
					}
				}
				if idx != -1 {
					dependencies[fieldPath] = append(deps[:idx], deps[idx+1:]...)
				}
			}
		}

		// If there was no resolved field, this means that there is a
		// circular dependency because all remaining fields are
		// dependent to at least another one.
		// TODO Find the circular circuit and improve the error
		// message.
		if len(resolved) == 0 {
			return nil, errors.New("circular dependency found")
		}

		// Resolved fields can be added to the result as is.
		fields = append(fields, resolved...)
	}

	return fields, nil
}
