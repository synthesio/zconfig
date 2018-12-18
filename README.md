# zConfig [![Build Status](https://travis-ci.org/synthesio/zconfig.svg?branch=master)](https://travis-ci.org/synthesio/zconfig) [![Godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/synthesio/zconfig) [![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/synthesio/zconfig/master/LICENSE.md)

`zConfig` is an extensible, reflection based configuration and dependency injection tool.

## Usage

Configuration and injection are defined field by field using tags:

```go
type Database struct {
    Address string `key:"address" description:"Database address"`
    Port    int    `key:"port" default:"3306"`
}

type Repository struct {
	DB *Database `inject:"database"`
}

type Service struct {
	DB *Database `key:"db" inject-as:"database"`
	Repository Repository
}

func main() {
	s := new(Service)
	err := Configure(s)
	//...
}
```

Running the compiled binary with the argument `--help` will cause it to output the configuration reference and exit with
a 0 status code.
 
See the [examples section](#examples) for a more detailed look at zConfig capabilities and usage.

#### Configuration tags reference

- `key` defines a configuration key whose value will be injected into the current field;
- `description` (optional) an explicative text for the configuration key, it will be displayed when the configuration
  reference is printed to the screen;
- `default` (optional) a default value for the configuration key, it will be used if no value for  the configuration key
  was found in any of the repositories. When no default is defined and no value is found for the configuration key,
  calls to `Configure` will return an error. 

#### Dependency Injection tags reference

- `inject-as` declares the field as an injectable dependency, associating it to the given alias;
- `inject` declares that the dependency defined with the given alias should be injected into the field

## Extending (Hooks)

An extension point is provided under the form of hook functions, whose type is:
```go
type Hook func(field *Field) error
```

Hooks may be added by calling the `AddHooks(hooks ...Hook)` function. 
They will be executed after the configuration is applied and dependencies have been injected.
All defined hooks will be invoked on every exported struct field (in the example above, fields `Address`, `Port`, `DB`,
 `Repository`) and on the root (the `s` variable in the example).
Fields marked as injection targets will not be passed to hook functions.

## Examples

You can find some usage examples in the [examples package](../blob/master/examples).
If you are trying to create an extension, take a look at the Initialize hook defined [here](../blob/master/init.go).
