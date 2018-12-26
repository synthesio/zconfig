# zConfig [![Build Status](https://travis-ci.org/synthesio/zconfig.svg?branch=master)](https://travis-ci.org/synthesio/zconfig) [![Godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/synthesio/zconfig) [![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/synthesio/zconfig/master/LICENSE.md)

`zConfig` is a Golang, extensible, reflection-based configuration and
dependency injection tool.

## Usage

_zconfig_ primary feature is an extensible configuration repository. To use it,
simply define a configuration struture and feed it to the `zconfig.Configure`
method. You can use the `key`, `description`and `default` tags to define which
key to use.

```go
type Configuration struct {
	Addr string `key:"addr" description:"address the server should bind to" default:":80"`
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
their description and default values.

```bash
$ ./a.out --help
Keys:
addr	ADDR	address the server should bind to	(:80)
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

```bash
$ ./a.out --help
Keys:
server.addr	SERVER_ADDR	address the server should bind to	(:80)
```

### Initialization

_zconfig_ does handle dependency initialization. Any reachable field of your
configuration struct (whatever the nesting level) that implements the
`zconfig.Initializable` interface will be `Init`ed during the configuration
process.

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
same way the compiler will refuse compiling a type definition with cycles.

## Repository

The configuration of the fields by the processor is deleguated to a
configuration repository. A repository is a list of _providers_ and _parsers_
than respectively retrieve configuration key's raw values and transform them
into the expected value. The configuration itself is done via a _hook_ added to
the default provider.

### Providers

The repository is based around a slice of structs implementing the
`zconfig.Provider` interface. Each provider has a name and priority and they
are requested in priority order when a key is needed.

The `args` and `env` providers are registered on the default provider, and
instances are exposed as `zconfig.Args` and `zconfig.Env` for convenience so
you can use them for other purposes, like getting the path for a `YAMLProvider`
file for example.

You can use the `zconfig.Repository.AddProviders(...zconfig.Provider)` method
to add a new provider to a given repository, or the
`zconfig.AddProviders(...zconfig.Provider)` shortcut to add one to the default
repository.

### Parsers

Parsing the raw string values into the actual variables for the configuration
struct is handled by the configuration repository. The repository holds a slice
of `zconfig.Parser` called in order until one of them is able to handle the
type of the destination field.

The following types actually have parsers in the default repository:

* `encoding.TextUnmarshaller`
* `encoding.BinaryUnmarshaller`
* `(u)?int(32|64)?`
* `float(32|64)`
* `string`
* `[]string`
* `bool`
* `time.Duration`
* `regexp.Regexp`

You can use the `zconfig.Repository.AddParsers(...zconfig.Parser)` method to
add a new parser to a given repository, or the
`zconfig.AddParsers(...zconfig.Parser)` shortcut to add to the default
repository.

For convenience, the various parsers of the default repository are available as
`zconfig.DefaultParsers`.

## Hooks

The standard way to extend the behavior of _zconfig_ is to implement a new
_hook_. The default `Processor` uses two of them: the `zconfig.Repository.Hook`
and the `zconfig.Initialize` hook.

Once the configuration struct is parsed and verified, the processor runs each
hook registered on the fields of the struct, allowing custom behavior to be
implemented for your systems. The interface for a hook is really simple:

```go
type Hook func(field *Field) error
```

Each encountered field will be passed to the hook, lowest dependencies first,
meaning that you can assume that the hook was executed on any child of the
current field before the field itself. The fields that have a configuration key
can be distinguished by the `Configurable` bool attribute being `true`. Note that
the hooks aren't executed on the injection targets to avoid  processing them
twice.

You can use the `zconfig.Processor.AddHooks(...zconfig.Hook)` method to add a
new hook to a given processor, or the `zconfig.AddHooks(...zconfig.Hook)`
shortcut to add one to the default processor.
