package zconfig

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"unicode"

	"github.com/fatih/structtag"
	"github.com/pkg/errors"
)

// A Processor handle the service processing and execute hooks on the resulting
// fields.
type Processor struct {
	lock  sync.Mutex
	hooks []Hook
}

func NewProcessor(hooks ...Hook) *Processor {
	return &Processor{
		hooks: hooks,
	}
}

func (p *Processor) Process(s interface{}) error {
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

	mark(root, "")

	for _, hook := range p.hooks {
		for _, field := range fields {
			err := hook(field)
			if err != nil {
				return errors.Wrapf(err, "executing hook on field %s", field.Path)
			}
		}
	}

	return nil
}

func (p *Processor) AddHooks(hooks ...Hook) {
	p.hooks = append(p.hooks, hooks...)
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
		field.Path = fmt.Sprintf("%s.%s", p.Path, s.Name)
		field.Tags, err = structtag.Parse(string(s.Tag))
		if err != nil {
			return nil, errors.Wrapf(err, "invalid tag for field %s", field.Path)
		}

		keyTag, err := field.Tags.Get(tagKey)
		if err == nil {
			field.Key = keyTag.Name
			if field.Key == "" {
				return field, errors.Errorf("invalid empty key for field %s", field.Path)
			}
		}
	}

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			if !v.CanSet() {
				return nil, errors.Errorf("cannot address %s for path %s", v.Type(), field.Path)
			}
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}

	if field.IsLeaf() {
		return field, nil
	}

	for i := 0; i < v.Type().NumField(); i++ {
		structField := v.Type().Field(i)
		if !unicode.IsUpper([]rune(structField.Name)[0]) {
			continue
		}

		child, err := walk(v.Field(i), structField, field)
		if err != nil {
			return nil, err
		}

		field.Children = append(field.Children, child)
	}

	return field, nil
}

type dependencies map[string]map[string]struct{}

func (d dependencies) add(f *Field, deps ...*Field) {
	fDeps, ok := d[f.Path]
	if !ok {
		fDeps = make(map[string]struct{})
	}
	for _, dep := range deps {
		fDeps[dep.Path] = struct{}{}
	}
	d[f.Path] = fDeps
}

func (d dependencies) remove(path string) {
	delete(d, path)

	for p := range d {
		delete(d[p], path)
	}
}

func resolve(root *Field) (fields []*Field, err error) {
	// Do a stack-based depth-first-search to retrieve all fields, and
	// identify the dependencies of the various fields.
	var (
		stack        = []*Field{root}
		paths        = make(map[string]*Field)
		sources      = make(map[string]*Field)
		targets      = make(map[*Field]string)
		dependencies = make(dependencies)
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
		// later in the process by copying the slice right now.]
		dependencies.add(e, e.Children...)
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

		dependencies.add(target, source)
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
			// Remove the field from the other fields dependencies list
			dependencies.remove(path)
		}

		// If there was no resolved field, this means that there is a
		// circular dependency because all remaining fields are
		// dependent to at least another one.
		if len(resolved) == 0 {
			return nil, newCycleError(dependencies)
		}

		for _, res := range resolved {
			// Do not add injection targets to resolved fields because their sources will also be added
			if _, ok := res.InjectionTarget(); !ok {
				fields = append(fields, res)
			}
		}
	}

	return fields, nil
}

func newCycleError(dependencies dependencies) error {
	var paths [][]string

	for path, deps := range dependencies {
		for depPath := range deps {
			paths = append(paths, []string{path, depPath})
		}
	}

	for {
		var next [][]string
		for _, path := range paths {
			for fieldPath := range dependencies[path[len(path)-1]] {
				if fieldPath == path[0] {
					return errors.Errorf("cycle detected: %s", strings.Join(path, " -> "))
				}

				next = append(next, append(path, fieldPath))
			}
		}
		paths = next
	}
}

func mark(f *Field, key string) bool {
	// If the field has no key and isn't anonymous, we can safely mark it
	// no-configurable.
	if f.Key == "" && !f.IsAnonymous() {
		return false
	}

	if f.Key != "" {
		key = key + "." + f.Key
	}

	if len(f.Children) == 0 {
		f.Configurable = true
		f.ConfigurationKey = key[1:]
		return true
	}

	var children = 0
	for _, c := range f.Children {
		ok := mark(c, key)
		if ok {
			children += 1
		}
	}

	if children == 0 && key != "" {
		f.Configurable = true
		f.ConfigurationKey = key[1:]
	}

	return children > 0
}
