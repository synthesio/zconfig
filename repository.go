package zconfig

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"sync"
)

// A Repository is list of configuration providers and hooks.
type Repository struct {
	lock      sync.Mutex
	providers []Provider
	parsers   []Parser
}

// Register a new Provider in this repository.
func (r *Repository) AddProviders(providers ...Provider) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.providers = append(r.providers, providers...)
	sort.Slice(r.providers, func(a, b int) bool {
		return r.providers[a].Priority() < r.providers[b].Priority()
	})
}

// Retrieve a key from the provider, by priority order.
func (r *Repository) Retrieve(key string) (value interface{}, provider string, found bool, err error) {
	for _, p := range r.providers {
		value, found, err = p.Retrieve(key)
		if err != nil {
			return nil, p.Name(), false, err
		}
		if found {
			return value, p.Name(), true, nil
		}
	}

	return nil, "", false, nil
}

var ErrNotParseable = errors.New("not parseable")

// Register allow anyone to add a custom parser to the list.
func (r *Repository) AddParsers(parsers ...Parser) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.parsers = append(r.parsers, parsers...)
}

// Parse the parameter depending on the kind of the field, returning an
// appropriately typed reflect.Value.
func (r *Repository) Parse(raw, res interface{}) (err error) {
	for _, p := range r.parsers {
		err = p(raw, res)
		if err == ErrNotParseable {
			continue
		}
		if err != nil {
			return fmt.Errorf("unable to parse %T: %s", res, err)
		}
		return nil
	}
	return fmt.Errorf("no parser for type %T", res)
}

func (r *Repository) Hook(f *Field) (err error) {
	if !f.Configurable {
		return nil
	}

	raw, _, found, err := r.Retrieve(f.ConfigurationKey)
	if err != nil {
		return fmt.Errorf("configuring field %s: retrieving key %s: %s", f.Path, f.ConfigurationKey, err)
	}

	if !found {
		def, ok := f.Tags.Lookup(TagDefault)
		if !ok {
			return fmt.Errorf("configuring field %s: missing key %s", f.Path, f.ConfigurationKey)
		}
		raw = def
	}

	var val = f.Value
	if val.Kind() != reflect.Ptr {
		val = val.Addr()
	}

	err = r.Parse(raw, val.Interface())
	if err != nil {
		return fmt.Errorf("configuring field %s: parsing value for key %s: %s", f.Path, f.ConfigurationKey, err)
	}

	return nil
}
