package zconfig

import (
	"reflect"
	"regexp"
	"testing"
	"time"
)

func TestParseString(t *testing.T) {
	for _, c := range []struct {
		raw interface{}
		res interface{}
		err bool
	}{
		// Not Parseable
		{raw: 1, res: "", err: true},

		// Mismatched types
		{raw: "foo", res: float64(0), err: true},

		// Strings
		{raw: "foo", res: "foo", err: false},

		// String slices
		{raw: "foo", res: []string{"foo"}, err: false},
		{raw: "foo  ", res: []string{"foo"}, err: false},
		{raw: "foo , bar ", res: []string{"foo", "bar"}, err: false},
		{raw: "  baz,foo , bar ", res: []string{"baz", "foo", "bar"}, err: false},

		// Int slices
		{raw: "10", res: []int{10}, err: false},
		{raw: "10  ", res: []int{10}, err: false},
		{raw: "10 , 20 ", res: []int{10, 20}, err: false},
		{raw: "  10,20 , 30 ", res: []int{10, 20, 30}, err: false},

		// Int64 slices
		{raw: "10", res: []int64{10}, err: false},
		{raw: "10  ", res: []int64{10}, err: false},
		{raw: "10 , 20 ", res: []int64{10, 20}, err: false},
		{raw: "  10,20 , 30 ", res: []int64{10, 20, 30}, err: false},

		// Numeric types
		{raw: "1", res: int(1), err: false},
		{raw: "1", res: int8(1), err: false},
		{raw: "1", res: int16(1), err: false},
		{raw: "1", res: int32(1), err: false},
		{raw: "1", res: int64(1), err: false},
		{raw: "1", res: uint(1), err: false},
		{raw: "1", res: uint8(1), err: false},
		{raw: "1", res: uint16(1), err: false},
		{raw: "1", res: uint32(1), err: false},
		{raw: "1", res: uint64(1), err: false},
		{raw: "1", res: float32(1), err: false},
		{raw: "1", res: float64(1), err: false},

		// Boolean types
		{raw: "true", res: true, err: false},
		{raw: "T", res: true, err: false},
		{raw: "false", res: false, err: false},
		{raw: "F", res: false, err: false},

		// Regexp
		{raw: "/a/", res: regexp.MustCompile("/a/"), err: false},

		// Unmarshalers
		{raw: "2019-01-11T15:01:31Z", res: time.Date(2019, time.January, 11, 15, 01, 31, 000, time.UTC), err: false},
	} {
		var typ = reflect.TypeOf(c.res)
		var ptr = false
		if typ.Kind() == reflect.Ptr {
			ptr = true
			typ = typ.Elem()
		}
		var res = reflect.New(typ).Interface()

		err := ParseString(c.raw, res)
		if (err != nil) != c.err {
			if c.err {
				t.Errorf("ParseString(%+v): should fail", c.raw)
			} else {
				t.Errorf("ParseString(%+v): unexpected error %v", c.raw, err)
			}
			continue
		}

		if c.err {
			continue
		}

		if !ptr {
			res = reflect.ValueOf(res).Elem().Interface()
		}

		if !reflect.DeepEqual(res, c.res) {
			t.Errorf("ParseString(%+v): wanted %+v, got %+v", c.raw, c.res, res)
		}
	}
}
