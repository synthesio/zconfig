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
	DefaultProcessor.AddHooks(DefaultRepository.Hook, Initialize)
	DefaultProcessor.AddHooksEx(InitializeEx)
}

// Configure a service using the default processor.
func Configure(s interface{}) error {
	return DefaultProcessor.Process(s)
}

// A Hook can be used to act upon every field visited by the repository when
// configuring a service.
type Hook func(field *Field) error

type HookEx func(context.Context, *Field) error

// Add a hook to the default repository.
func AddHooks(hooks ...Hook) {
	DefaultProcessor.AddHooks(hooks...)
}

func AddHooksEx(hooksEx ...HookEx) {
	DefaultProcessor.AddHooksEx(hooksEx...)
}

// Provider is the interface implemented by all entity a configuration key can
// be retrieved from.
type Provider interface {
	Retrieve(key string) (value interface{}, found bool, err error)
	Name() string
	Priority() int
}

// Add a provider to the default repository.
func AddProviders(providers ...Provider) {
	DefaultRepository.AddProviders(providers...)
}

// Parser is the type of function that can convert a raw representation to a
// given type.
type Parser func(interface{}, interface{}) error

// Add a parser to the default repository.
func AddParsers(parsers ...Parser) {
	DefaultRepository.AddParsers(parsers...)
}
