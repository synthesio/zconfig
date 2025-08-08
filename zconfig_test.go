package zconfig

import (
	"context"
	"testing"
	"time"
)

type S struct {
	Foo *string   `key:"foo"`
	Baz string    `key:"baz"`
	Quz string    `key:"quz" default:"some"`
	T   time.Time `key:"t"`
	E   E         `key:"e"`
	F   F         `key:"f"`
	Bar string
	EE  EE `key:"ee"`
	A   *A `key:"a" inject-as:"a"`
	L   *A `inject:"a"`
	M   M  `key:"m"`
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

type Tester struct {
	t *testing.T
}

func (tester *Tester) Hook(ctx context.Context, field *Field) error {
	expectedProviders := map[string]string{
		"foo": "test",
		"baz": "test2",
		"quz": ProviderDefault,
	}

	expectedProvider, found := expectedProviders[field.Key]
	if !found {
		return nil
	}

	if field.Provider != expectedProvider {
		tester.t.Fatalf("unexpected provider for key %q, got %q, expected %q", field.Key, field.Provider, expectedProvider)
	}

	return nil
}

func TestConfigure(t *testing.T) {
	tester := Tester{t}

	var p = TestProvider{"test", map[string]string{
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
	var p2 = TestProvider{
		"test2",
		map[string]string{
			"baz": "baz",
		}}
	AddHooks(tester.Hook)
	AddProviders(p, p2)
	var s S
	err := Configure(context.Background(), &s)
	if err != nil {
		t.Fatal(err)
	}
}
