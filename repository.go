package zconfig

import (
	"reflect"
	"sort"
	"sync"

	"github.com/pkg/errors"
)

// A Repository is list of configuration providers and hooks.
type Repository struct {
	lock      sync.Mutex
	providers []Provider
	parsers   []Parser
}

// Register a new Provider in this repository.
func (r *Repository) AddProvider(p Provider) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.providers = append(r.providers, p)
	sort.Slice(r.providers, func(a, b int) bool {
		return r.providers[a].Priority() < r.providers[b].Priority()
	})
}

// Retrieve a key from the provider, by priority order.
func (r *Repository) Retrieve(key string) (value, provider string, found bool, err error) {
	for _, p := range r.providers {
		value, found, err = p.Retrieve(key)
		if err != nil {
			return "", "", false, err
		}
		if found {
			return value, p.Name(), true, nil
		}
	}

	return "", "", false, nil
}

// Register allow anyone to add a custom parser to the list.
func (r *Repository) AddParsers(parsers ...Parser) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.parsers = append(r.parsers, parsers...)
}

// Parse the parameter depending on the kind of the field, returning an
// appropriately typed reflect.Value.
func (r *Repository) Parse(typ reflect.Type, parameter string) (val reflect.Value, err error) {
	for _, p := range r.parsers {
		if !p.CanParse(typ) {
			continue
		}

		val, err = p.Parse(typ, parameter)
		if err != nil {
			return val, errors.Wrapf(err, "unable to parse %s", typ)
		}

		return val, nil
	}

	return val, errors.Errorf("no parser for type %s", typ)
}

func (r *Repository) Hook(f *Field) (err error) {
	if !f.Configurable {
		return nil
	}

	raw, _, found, err := r.Retrieve(f.ConfigurationKey)
	if err != nil {
		return errors.Wrapf(err, "configuring field %s: retrieving key %s", f.Path, f.ConfigurationKey)
	}

	if !found {
		def, ok := f.FullTag(TagDefault)
		if !ok {
			return errors.Errorf("configuring field %s: missing key %s", f.Path, f.ConfigurationKey)
		}
		raw = def
	}

	res, err := r.Parse(f.Value.Type(), raw)
	if err != nil {
		return errors.Wrapf(err, "configuring field %s: parsing value for key %s", f.Path, f.ConfigurationKey)
	}

	if !f.Value.CanSet() {
		return errors.Errorf("configuring field %s: can't set value", f.Path)
	}

	f.Value.Set(res)
	return nil
}
