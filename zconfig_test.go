package zconfig

import (
	"testing"
	"time"
)

type S struct {
	Foo   *string   `key:"foo"`
	T     time.Time `key:"t"`
	E     E         `key:"e"`
	F     F         `key:"f"`
	babar string
	Bar   string
	EE    EE `key:"ee"`
	A     A  `key:"a" inject-as:"a"`
	L     A  `inject:"a"`
	M     M  `key:"m"`
}

type A struct {
	B struct {
		C struct {
			D int `key:"d"`
		} `key:"c"`
	} `key:"b"`
}

type F struct {
	B int `key:"b"`
}

type E struct {
	F
	R int `key:"r"`
}

type EE struct {
	E
}

type M struct {
	Foo int `key:"foo"`
	N   struct {
		unexp int
	}
}

func TestConfigure(t *testing.T) {
	var p = TestProvider{map[string]string{
		"foo":     "a",
		"t":       "2018-12-18T11:29:00+02:00",
		"e.r":     "1",
		"e.b":     "2",
		"f.b":     "3",
		"ee.r":    "4",
		"ee.b":    "5",
		"a.b.c.d": "6",
		"m.foo":   "7",
	}}
	AddProviders(p)
	var s S
	err := Configure(&s)
	if err != nil {
		t.Fatal(err)
	}
}
