# zConfig [![Build Status](https://travis-ci.org/synthesio/zconfig.svg?branch=master)](https://travis-ci.org/synthesio/zconfig) [![Godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/synthesio/zconfig) [![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/synthesio/zconfig/master/LICENSE.md)

`zConfig` is a Golang, extensible, reflection-based configuration and
dependency injection tool whose goal is to get rid of the boilerplate code
needed to configure and initialize an application's dependencies.

## Usage

_zconfig_ primary feature is an extensible configuration repository. To use it,
simply define a configuration structure and feed it to the `Configure()`
method. You can use the `key`, `description`, `default` and `example` tags to define which
key to use.

```go
type Configuration struct {
	Addr string `key:"addr" description:"address the server should bind to" default:":80"`
	Name string `key:"name" description:"name displayed to the client" example:"zconfig"`
}

func main() {
	var c Configuration
	err := zconfig.Configure(&c)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	//...
	err := http.ListenAndServe(c.Addr, nil)
	//...
}
```

Once compiled, the special flag `help` can be passed to the binary to display a
list of the available configuration keys, in their cli and env form, as well as
their description and default/example values.

```shell
$ ./a.out --help
Keys:
addr	ADDR	address the server should bind to	(:80)
name	NAME	name displayed to the client	example: zconfig
```

Configurations can be nested into structs to improve usability, and the keys of
the final parameters are prefixed by the keys of all parents.

```go
type Configuration struct {
	Server struct{
		Addr string `key:"addr" description:"address the server should bind to" default:":80"`
	} `key:"server"`
}

func main() {
	var c Configuration
	err := zconfig.Configure(&c)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	//...
	err := http.ListenAndServe(c.Server.Addr, nil)
	//...
}
```

```shell
$ ./a.out --help
Keys:
server.addr	SERVER_ADDR	address the server should bind to	(:80)
```

The following types are handled by default by the library:

- `encoding.TextUnmarshaller`
- `encoding.BinaryUnmarshaller`
- `(u)?int(32|64)?`
- `float(32|64)`
- `string`
- `[]string`
- `bool`
- `time.Duration`
- `regexp.Regexp`

### Initialization

_zconfig_ does handle dependency initialization. Any reachable field of your
configuration struct (whatever the nesting level) that implements the
`Initializable` interface will be initialized during the configuration process.

Here is an example with our internal Redis wrapper.

```go
package zredis

import (
	"time"
	"github.com/go-redis/redis"
)

type Client struct {
	*redis.Client
	Address         string        `key:"address" description:"address and port of the redis"`
	ConnMaxLifetime time.Duration `key:"connection-max-lifetime" description:"maximum duration of open connection" default:"30s"`
	MaxOpenConns    int           `key:"max-open-connections" description:"maximum of concurrent open connections" default:"10"`

}

func (c *Client) Init() (err error) {
	c.Client = redis.NewClient(&redis.Options{
		Network:     "tcp",
		Addr:        c.Address,
		IdleTimeout: c.ConnMaxLifetime,
		PoolSize:    c.MaxOpenConns,
	})

	_, err = c.Ping()
	return err
}
```

Now, whenever we need to use a Redis database in our services, we can simply
declare the dependency in the configuration struct and go on without worrying
about initializing it, liberating your service from pesky initialization code.

```go
package main

import (
	"zredis"
	"github.com/synthesio/zconfig"
)

type Service struct {
	Redis *zredis.Client `key:"redis"`
}

func main() {
	var s Service
	err := zconfig.Configure(&s)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	res, err := s.Redis.Get("foo").Result()
	// ...
}
```

### Injection

The _zconfig_ processor understands a set of tags used for injecting one field
into another, thus sharing resources. The `inject-as` tag defines an _injection
source_ and the key used to identify it, while the `inject` tag defines an
_injection target_ and the key of the source to use.

Any type can be injected as long as the source is _assignable_ to the target.
This is especially useful to allow sharing common configuration fields or even
whole structs like a database handle.

```go
type Service struct {
	Foo struct{
		Root string `inject:"bar-root"`
	}
	Bar struct{
		Root string `inject-as:"bar-root"`
	}
}
```

Note that the injection system isn't tied to the configuration one: you don't
need your injection source or target to be part of a chain of `key`ed structs.

Also, _zconfig_ will return an error if given a struct with a cycle in it, the
same way the compiler will refuse to compile a type definition with cycles.

## How it works

Under the hood, the work is done by a
[`Processor`](https://godoc.org/github.com/synthesio/zconfig#Processor).
The _Processor_'s role is to construct a list of `Field` from the given
struct, and run a number of hooks on this list.

The `Field` struct is a graph representation of a single field of your
configuration struct, with pointers for parent and children. The list of fields
handled by the processor is ordered by deepest dependency first, meaning that
for any given hook, all children of a given field are processed by the hook
before the field itself. For the case of injection, the targets aren't
included in this list, but the sources are processed before the target's
branch.

For convenience, _zconfig_ provides a default processor already setup to use 2
hooks: the first is the one that do the actual configuration of the fields, and
the second do the initialization of the field. The global `Configure()` and
`AddHooks()` methods are shortcuts to the methods of this default processor.

### Hook

The `Hook` is a type for a function that takes a single pointer to a `Field` as
parameter, and returns an error if need be.

```go
type Hook func(field *Field) error
```

One good example of hook is the one used for initializing the fields:

```go
type Initializable interface {
	Init() error
}

// Used for type comparison.
var typeInitializable = reflect.TypeOf((*Initializable)(nil)).Elem()

func Initialize(field *Field) error {
	// Not initializable, nothing to do.
	if !field.Value.Type().Implements(typeInitializable) {
		return nil
	}

	// Initialize the element itself via the interface.
	err := field.Value.Interface().(Initializable).Init()
	if err != nil {
		return errors.Wrap(err, "initializing field")
	}

	return nil
}
```

### Repository

The first hook setup in the default processor, and the main feature of the
library, is the configuration repository hook. The `Repository` is a struct
holding a list of `Provider` interfaces and `Parser` function used by the hook
to set the values of the configuration struct.

#### Provider

A _provider_ is a simple interface for retrieving the value associated with a
key. Each provider listed by the repository is consulted until one of them
returns a result.

```go
type Provider interface {
	Retrieve(key string) (value interface{}, found bool, err error)
	Name() string
	Priority() int
}
```

The `Retrieve()` method gets a configuration _key_ and return the raw value
associated with it, whether it found something or not, and whether an error
occurred during the process.

The `Priority()` method defines the order in which the providers are checked by
the repository, the lowest going first. The `ArgsProvider` is priority 1, and
the `EnvProvider` is priority 2.

The `Name()` method helps to know which source provided the value, which can be
useful for various repository extensions.

A provider can be added to a repository using the `AddProviders()` method.

For example, the default repository has two providers registered: the
`ArgsProvider` that look on the CLI arguments and the `EnvProvider` that look
at the program's environment.

#### Parser

A _parser_ is a function for converting a raw value to another. The `dst`
parameter is always a pointer to the expected value.

```go
type Parser func(raw interface{}, dst interface{}) error
```

The default repository has a `ParseString` registered that handle the
convertion listed above from their matching string representation, obviously
intended to work with the values from the `Args` and `Env` providers.

## Frequently Asked Questions

### _How can I disable the CLI flags?_

Remove them from the providers of your repository. For the classic processor,
you want to do this:

```go
var repository zconfig.Repository
repository.AddProviders(zconfig.Env)
repository.AddParsers(zconfig.ParseString)

var processor zconfig.Processor
processor.AddHooks(repository.Hook, zconfig.Initialize)
```

### _I want to validate the values from the configuration before using them_

First obvious way would be to use custom types implementing the
`encoding.TextUnmarshaller` interface and do the check here. That would add
being explicit in the configuration by having the advantage of not allowing
inconsistent state. In the same web-form validation style, you could add
additional validation tags to your struct and create a hook to check that the
value matches the rules.

Another way would be to do it in the `Init()` method of your field, so the
initialization hook will handle the check. This has the advantage of not
forcing custom types for the runtime types, and having the ability to
cross-check multiple fields by using the parent's struct method.

### _Can I configure multiple structs during the program's lifetime?_

Of course. The `Processor.Process()` method is completely self-contained, and
doesn't use any state from the `Processor` except the list of hooks to apply.
Same thing goes for the `Repository` and the basic `Provider` of _zconfig_.

### _I want to read my configuration from "insert source name here"_

What you want is a custom provider. If the set provided by _zconfig_ itself
doesn't cover your way of defining configuration, you can always add one to the
default repository (or define your own).

Here is a quick-and-dirty example you can use as basis for a provider getting
its values from an arbitrary JSON file.

```go
import "github.com/tidwall/gjson"

type JSONProvider struct {
	raw gjson.Result
}

func (p JSONProvider) Retrieve(key string) (raw interface{}, found bool, err error) {
	field := p.raw.Get(key)
	return field.Value(), field.Exists(), nil
}

func (JSONProvider) Name() string {
	return "json"
}

func (JSONProvider) Priority() int {
	// args are 1, env are 2, we want both of them to override the
	// configuration file so we set this provider to be looked at after
	// them.
	return 3
}

func NewJSONProviderFromFile(path string) (*JSONProvider, err error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "reading field %s", path)
	}

	if !gjson.ValidBytes(raw) {
		return nil, errors.Errorf("invalid json file %s", path)
	}

	return  &JSONProvider{gjson.ParseBytes(raw)}, nil
}
```

The library `tidwall/gjson` is an alternative library for manipulating JSON
that fits this use case particularly well: it doesn't parse the whole string,
and only look for the field identified by the given key (whose format match the
_zconfig_ one.)

Amongst the improvements possible, this provider could be constructed using a
value retrieved from `zconfig.Args` or `zconfig.Env` so the path can be given
on the command-line of your program.

For example:

```go
func NewJSONProvider() (*JSONProvider, err error) {
	path, ok, err := zconfig.Args.Retrieve("configuration")
	if err != nil {
		return nil, err
	}

	if !ok {
		return &JSONProvider{}, nil
	}

	return NewJSONProviderFromFile(path)
}
```
