package zconfig

import (
	"reflect"
)

var (
	defaultRepository Repository
	defaultProcessor  Processor
)

func init() {
	defaultRepository.AddProvider(NewEnvProvider())
	defaultRepository.AddProvider(NewArgsProvider())
	defaultRepository.AddParsers(DefaultParsers...)
	defaultProcessor.AddHooks(defaultRepository.Hook)
	defaultProcessor.AddHooks(Initialize)
}

// Configure a service using the default processor.
func Configure(s interface{}) error {
	return defaultProcessor.Process(s)
}

// A Hook can be used to act upon every field visited by the repository when
// configuring a service.
type Hook func(field *Field) error

// Add a hook to the default repository.
func AddHooks(hooks ...Hook) {
	defaultProcessor.AddHooks(hooks...)
}

// Provider is the interface implemented by all entity a configuration key can
// be retrieved from.
type Provider interface {
	Retrieve(key string) (value string, found bool, err error)
	Name() string
	Priority() int
}

// Add a provider to the default repository.
func AddProvider(p Provider) {
	defaultRepository.AddProvider(p)
}

// Parser is the interface implemented by a struct that can convert a raw
// string representation to a given type.
type Parser interface {
	Parse(reflect.Type, string) (reflect.Value, error)
	CanParse(reflect.Type) bool
}

// Add a parser to the default repository.
func AddParsers(parsers ...Parser) {
	defaultRepository.AddParsers(parsers...)
}
