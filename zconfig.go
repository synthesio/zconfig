package zconfig

import (
	"context"
)

var (
	DefaultRepository Repository
	DefaultProcessor  Processor
	Args              = NewArgsProvider()
	Env               = NewEnvProvider()
)

func init() {
	DefaultRepository.AddProviders(Args, Env)
	DefaultRepository.AddParsers(ParseString)
	DefaultProcessor.AddHooks(DefaultRepository.Hook, Inject, Initialize)
}

// Configure a service using the default processor.
func Configure(ctx context.Context, s interface{}) error {
	return DefaultProcessor.Process(ctx, s)
}

// A Hook can be used to act upon every field visited by the repository when
// configuring a service.
type Hook func(ctx context.Context, field *Field) error

// AddHooks Adds the given hooks to the default repository.
func AddHooks(hooks ...Hook) {
	DefaultProcessor.AddHooks(hooks...)
}

// Provider is the interface implemented by all entity a configuration key can
// be retrieved from.
type Provider interface {
	Retrieve(key string) (value interface{}, found bool, err error)
	Name() string
	Priority() int
}

// AddProviders Adds the given providers to the default repository.
func AddProviders(providers ...Provider) {
	DefaultRepository.AddProviders(providers...)
}

// Parser is the type of function that can convert a raw representation to a
// given type.
type Parser func(interface{}, interface{}) error

// AddParsers Adds the given parsers to the default repository.
func AddParsers(parsers ...Parser) {
	DefaultRepository.AddParsers(parsers...)
}
