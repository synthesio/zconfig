package zconfig

import (
	"encoding"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type ParserFunc struct {
	Type reflect.Type
	Func func(reflect.Type, string) (reflect.Value, error)
}

func (p ParserFunc) CanParse(typ reflect.Type) bool {
	return typ.AssignableTo(p.Type) || reflect.PtrTo(typ).AssignableTo(p.Type)
}

func (p ParserFunc) Parse(typ reflect.Type, raw string) (val reflect.Value, err error) {
	withValue := typ.AssignableTo(p.Type)
	withPointer := reflect.PtrTo(typ).AssignableTo(p.Type)

	if !withValue && !withPointer {
		return val, errors.Errorf("cannot parse %s", typ)
	}

	if withPointer {
		typ = reflect.PtrTo(typ)
	}

	val, err = p.Func(typ, raw)
	if err != nil {
		return val, errors.Wrapf(err, "unable to parse %s", typ)
	}

	if withPointer {
		val = val.Elem()
	}

	return val, nil
}

var DefaultParsers = []Parser{
	ParserFunc{reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem(), parseTextUnmarshaler},
	ParserFunc{reflect.TypeOf((*encoding.BinaryUnmarshaler)(nil)).Elem(), parseBinaryUnmarshaler},
	ParserFunc{reflect.TypeOf([]string(nil)), parseStringSlice},
	ParserFunc{reflect.TypeOf(string("")), parseString},
	ParserFunc{reflect.TypeOf(bool(false)), parseBool},
	ParserFunc{reflect.TypeOf(float32(0)), parseFloat},
	ParserFunc{reflect.TypeOf(float64(0)), parseFloat},
	ParserFunc{reflect.TypeOf(uint(0)), parseUint},
	ParserFunc{reflect.TypeOf(uint8(0)), parseUint},
	ParserFunc{reflect.TypeOf(uint16(0)), parseUint},
	ParserFunc{reflect.TypeOf(uint32(0)), parseUint},
	ParserFunc{reflect.TypeOf(uint64(0)), parseUint},
	ParserFunc{reflect.TypeOf(int(0)), parseInt},
	ParserFunc{reflect.TypeOf(int8(0)), parseInt},
	ParserFunc{reflect.TypeOf(int16(0)), parseInt},
	ParserFunc{reflect.TypeOf(int32(0)), parseInt},
	ParserFunc{reflect.TypeOf(int64(0)), parseInt},
	ParserFunc{reflect.TypeOf(time.Duration(0)), parseDuration},
	ParserFunc{reflect.TypeOf(new(regexp.Regexp)), parseRegexp},
}

func parseString(typ reflect.Type, parameter string) (val reflect.Value, err error) {
	return reflect.ValueOf(parameter), nil
}

func parseStringSlice(_ reflect.Type, parameter string) (val reflect.Value, err error) {
	var out []string
	for _, r := range strings.Split(parameter, ",") {
		r = strings.TrimSpace(r)
		if len(r) == 0 {
			continue
		}
		out = append(out, r)
	}
	return reflect.ValueOf(out), nil
}

func parseTextUnmarshaler(typ reflect.Type, parameter string) (val reflect.Value, err error) {
	var i interface{}
	if typ.Kind() == reflect.Ptr {
		i = reflect.New(typ.Elem()).Interface()
	} else {
		i = reflect.New(typ).Elem().Interface()
	}
	err = i.(encoding.TextUnmarshaler).UnmarshalText([]byte(parameter))
	return reflect.ValueOf(i), err
}

func parseBinaryUnmarshaler(typ reflect.Type, parameter string) (val reflect.Value, err error) {
	var i interface{}
	if typ.Kind() == reflect.Ptr {
		i = reflect.New(typ.Elem()).Interface()
	} else {
		i = reflect.New(typ).Elem().Interface()
	}
	err = i.(encoding.BinaryUnmarshaler).UnmarshalBinary([]byte(parameter))
	return reflect.ValueOf(i), err
}

func parseUint(typ reflect.Type, parameter string) (val reflect.Value, err error) {
	v, err := strconv.ParseUint(parameter, 10, typ.Bits())
	if err != nil {
		return val, err
	}
	return reflect.ValueOf(v).Convert(typ), nil
}

func parseInt(typ reflect.Type, parameter string) (val reflect.Value, err error) {
	v, err := strconv.ParseInt(parameter, 10, typ.Bits())
	if err != nil {
		return val, err
	}
	return reflect.ValueOf(v).Convert(typ), nil
}

func parseFloat(typ reflect.Type, parameter string) (val reflect.Value, err error) {
	v, err := strconv.ParseFloat(parameter, typ.Bits())
	if err != nil {
		return val, err
	}
	return reflect.ValueOf(v).Convert(typ), nil
}

func parseBool(typ reflect.Type, parameter string) (val reflect.Value, err error) {
	if parameter == "" {
		return reflect.ValueOf(true), nil
	}

	v, err := strconv.ParseBool(parameter)
	if err != nil {
		return val, err
	}
	return reflect.ValueOf(v), nil
}

func parseDuration(typ reflect.Type, parameter string) (val reflect.Value, err error) {
	duration, err := time.ParseDuration(parameter)
	if err != nil {
		return val, err
	}

	return reflect.ValueOf(duration).Convert(typ), nil
}

func parseRegexp(typ reflect.Type, parameter string) (val reflect.Value, err error) {
	if parameter == "" {
		return val, nil
	}

	regex, err := regexp.Compile(parameter)
	if err != nil {
		return val, err
	}

	return reflect.ValueOf(regex).Convert(typ), nil
}
