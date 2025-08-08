package zconfig

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"unicode"

	"github.com/hchargois/flexwriter"
)

// A Processor handle the service processing and execute hooks on the resulting
// fields.
type Processor struct {
	hooks []Hook

	// Usage message to be displayed on error or when help is requested.
	// Deprecated: use UsageVal instead
	Usage func([]*Field)

	// UsageVal allows defining a custom usage function, i.e. a function that is
	// called when the --help flag is passed, and that is expected to print some
	// helpful message to the standard output. The argument value is the value
	// of the --help flag, e.g. --help=somevalue; if --help is passed without a
	// value then UsageVal will be called with an empty string as value.
	// If UsageVal is unset, then Usage is used. If Usage is unset too, then
	// DefaultUsageVal is used.
	UsageVal func(value string, fields []*Field)
}

func NewProcessor(hooks ...Hook) *Processor {
	return &Processor{
		hooks: hooks,
	}
}

func (p *Processor) Process(ctx context.Context, s interface{}) error {
	v := reflect.ValueOf(s)

	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("expected pointer to struct, %T given", s)
	}

	if v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("expected pointer to struct, %T given", s)
	}

	root, err := walk(v, reflect.StructField{}, nil)
	if err != nil {
		return fmt.Errorf("walking struct: %w", err)
	}

	fields, err := resolve(root)
	if err != nil {
		return fmt.Errorf("resolving struct: %w", err)
	}

	mark(root, "")

	if rawVal, ok, _ := Args.Retrieve("help"); ok {
		// we know rawVal is a string since it's coming from an ArgsProvider.
		val := rawVal.(string)

		var usage func(string, []*Field)
		switch {
		case p.UsageVal != nil:
			usage = p.UsageVal
		case p.Usage != nil:
			usage = func(_ string, fields []*Field) { p.Usage(fields) }
		default:
			usage = DefaultUsageVal
		}

		usage(val, fields)
		os.Exit(0)
	}

	for _, hook := range p.hooks {
		for _, field := range fields {
			err := hook(ctx, field)
			if err != nil {
				return fmt.Errorf("executing hook on field %s: %w", field.Path, err)
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

outer:
	for i := 0; i < v.Type().NumField(); i++ {
		structField := v.Type().Field(i)
		if !unicode.IsUpper([]rune(structField.Name)[0]) {
			continue
		}

		// Look for the field's own type in it's ancestry. If we find one,
		// consider this field as a leaf because it would otherwise end-up in
		// an infinite loop. See gorm.io/gorm.DB (in v1.22.4) for an example.
		// Fixes #46.
		fieldType := structField.Type
		for fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
		for ancestor := field; ancestor != nil; ancestor = ancestor.Parent {
			ancestorType := ancestor.Value.Type()
			for ancestorType.Kind() == reflect.Ptr {
				ancestorType = ancestorType.Elem()
			}

			if ancestorType == fieldType {
				continue outer
			}
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
			if e.Value.Kind() != reflect.Ptr {
				return nil, fmt.Errorf("cannot inject non pointer type %s, defined at path %s", e.Value.Type().Name(), e.Path)
			}

			if s, found := sources[key]; found {
				return nil, fmt.Errorf("injection source key %s already defined at path %s", key, s.Path)
			}
			sources[key] = e
		}

		if key, found := e.Tags.Lookup(TagInject); found {
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
			return nil, fmt.Errorf("injecting field %s into %s: %w", source.Path, target.Path, err)
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

func newCycleError(deps dependencies) error {
	var paths [][]string

	for path, dependencies := range deps {
		for depPath := range dependencies {
			paths = append(paths, []string{path, depPath})
		}
	}

	for {
		var next [][]string
		for _, path := range paths {
			for fieldPath := range deps[path[len(path)-1]] {
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

// DefaultUsageVal prints a usage message that lists the fields with their keys
// in CLI form (e.g. --foo) and environment variable form (e.g. FOO), as well as
// the fields descriptions and default values (if any).
//
// If called with the "cli" value, only the CLI form is printed, and if called
// with the "env" value, only the environment variable form is printed. Any
// other value (including an empty value) prints both forms.
func DefaultUsageVal(val string, fields []*Field) {
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

	required := flexwriter.New()
	optional := flexwriter.New()

	columns := []flexwriter.Column{
		flexwriter.Rigid{},      // CLI option name
		flexwriter.Rigid{},      // env variable name
		flexwriter.Shrinkable{}, // description
		flexwriter.Rigid{},      // default value
	}
	switch val {
	case "env":
		columns[0] = flexwriter.Omit{}
	case "cli":
		columns[1] = flexwriter.Omit{}
	}
	required.SetColumns(columns...)
	optional.SetColumns(columns...)

	for _, key := range keys {
		field := options[key]
		desc, _ := field.Tags.Lookup(TagDescription)

		row := []any{"--" + key, Env.FormatKey(key), desc}

		def, ok := field.Tags.Lookup(TagDefault)
		if ok {
			optional.WriteRow(append(row, "("+def+")")...)
		} else {
			required.WriteRow(row...)
		}
	}

	fmt.Printf("\nRequired parameters:\n")
	_ = required.Flush()

	fmt.Printf("\nOptional parameters:\n")
	_ = optional.Flush()
}

// DefaultUsage prints a usage message as DefaultUsageVal would with an empty
// value.
//
// Deprecated: use DefaultUsageVal.
func DefaultUsage(fields []*Field) {
	DefaultUsageVal("", fields)
}
