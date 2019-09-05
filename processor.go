package zconfig

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"sync"
	"text/tabwriter"
	"unicode"
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
		return fmt.Errorf("expected pointer to struct, %T given", s)
	}

	if v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("expected pointer to struct, %T given", s)
	}

	root, err := walk(v, reflect.StructField{}, nil)
	if err != nil {
		return fmt.Errorf("walking struct: %s", err)
	}

	fields, err := resolve(root)
	if err != nil {
		return fmt.Errorf("resolving struct: %s", err)
	}

	mark(root, "")

	if _, ok, _ := Args.Retrieve("help"); ok {
		usage(fields)
		os.Exit(0)
	}

	for _, hook := range p.hooks {
		for _, field := range fields {
			err := hook(field)
			if err != nil {
				return fmt.Errorf("executing hook on field %s: %s", field.Path, err)
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
		field.Path = "$"
		field.Anonymous = true
	} else {
		field.Path = fmt.Sprintf("%s.%s", p.Path, s.Name)
		field.Anonymous = s.Anonymous
		field.Tags = s.Tag

		key, ok := field.Tags.Lookup(TagKey)
		if ok {
			field.Key = key
			if field.Key == "" {
				return field, fmt.Errorf("invalid empty key for field %s", field.Path)
			}
		}
	}

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			if !v.CanSet() {
				return nil, fmt.Errorf("cannot address %s for path %s", v.Type(), field.Path)
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

		if key, ok := e.Tags.Lookup(TagInjectAs); ok {
			if s, ok := sources[key]; ok {
				return nil, fmt.Errorf("injection source key %s already defined at path %s", key, s.Path)
			}
			sources[key] = e
		}

		if key, ok := e.Tags.Lookup(TagInject); ok {
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
			return nil, fmt.Errorf("injection source key %s undefined for path %s", key, target.Path)
		}

		err := target.Inject(source)
		if err != nil {
			return nil, fmt.Errorf("injecting field %s into %s: %s", source.Path, target.Path, err)
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
			if _, ok := res.Tags.Lookup(TagInject); !ok {
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
					return fmt.Errorf("cycle detected: %s", strings.Join(path, " -> "))
				}

				next = append(next, append(path, fieldPath))
			}
		}
		paths = next
	}
}

// Mark the configurable fields and compute their configuration key. Return
// true if the current field or one of its children is configurable (used for
// recursion.)
func mark(f *Field, key string) bool {
	if f.Key == "" {
		// A field with no key and that is not anonymous isn't
		// configurable.
		if !f.Anonymous {
			return false
		}

		// A field with no key, anonymous but without exported children
		// isn't configurable either.
		if len(f.Children) == 0 {
			return false
		}
	}

	// Derive the key if needed.
	if f.Key != "" {
		key = key + "." + f.Key
	}

	// Mark the children and count the number of marked children.
	var children = 0
	for _, c := range f.Children {
		ok := mark(c, key)
		if ok {
			children += 1
		}
	}

	// A field with no key at this point is anonymous. It can't be
	// configured, but should return whether one of his children can be.
	if f.Key == "" {
		return children > 0

	}

	// If the field has no marked children at this point, mark it.
	if children == 0 {
		f.Configurable = true
		f.ConfigurationKey = key[1:]
	}

	// A field with a key should always return true.
	return true
}

func usage(fields []*Field) {
	var keys []string
	var options = make(map[string]*Field)
	for _, f := range fields {
		if !f.Configurable {
			continue
		}
		keys = append(keys, f.ConfigurationKey)
		options[f.ConfigurationKey] = f
	}
	sort.Strings(keys)

	fmt.Printf("\nKeys:\n")
	w := tabwriter.NewWriter(os.Stdout, 2, 2, 2, ' ', 0)
	for _, key := range keys {
		field := options[key]

		def, ok := field.Tags.Lookup(TagDefault)
		desc, _ := field.Tags.Lookup(TagDescription)
		if ok {
			fmt.Fprintf(w, "%s\t%s\t%s\t(%s)\n", key, Env.FormatKey(key), desc, def)
		} else if ex, ok := field.Tags.Lookup(TagExample); ok {
			fmt.Fprintf(w, "%s\t%s\t%s\texample: %s\n", key, Env.FormatKey(key), desc, ex)
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\t\n", key, Env.FormatKey(key), desc)
		}
	}
	_ = w.Flush()
}
