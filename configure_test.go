package zconfig

import (
	"reflect"
	"strings"
	"testing"
)

type Service struct {
	Workers     int              `key:"workers"`
	Dependency  SimpleDependency `key:"dependency" inject-as:"dependency"`
	Injected    SimpleDependency `inject:"dependency"`
	notExported SimpleDependency
}

type SimpleDependency struct {
	Foo int `key:"foo"`
}

func TestWalk(t *testing.T) {
	var v = reflect.ValueOf(new(Service))
	root, err := walk(v, reflect.StructField{}, nil)
	if err != nil {
		t.Fatalf("walking service: %s", err)
	}

	expected := &Field{Path: "root", Children: []*Field{
		{Path: "root.Workers"},
		{Path: "root.Dependency", Children: []*Field{
			{Path: "root.Dependency.Foo"},
		}},
		{Path: "root.Injected"},
	}}

	displayGraph(t, root, 0)
	testGraph(t, root, expected, nil)
}

func displayGraph(t *testing.T, f *Field, l int) {
	t.Log(strings.Repeat("\t", l), f.Path)
	for _, c := range f.Children {
		displayGraph(t, c, l+1)
	}
}

func testGraph(t *testing.T, actual, expected, parent *Field) {
	if actual.Path != expected.Path {
		t.Fatalf("invalid value for path: expected %s, got %s", expected.Path, actual.Path)
	}

	if actual.Parent != parent {
		t.Fatalf("invalud parent for path %s: wanted %p, got %p", actual.Path, parent, actual.Parent)
	}

	if len(actual.Children) != len(expected.Children) {
		t.Fatalf("invalid number of children for path %s: wanted %d, got %d", actual.Path, len(expected.Children), len(actual.Children))
	}

	for i := 0; i < len(actual.Children); i++ {
		testGraph(t, actual.Children[i], expected.Children[i], actual)
	}
}

func TestResolve(t *testing.T) {
	var v = reflect.ValueOf(new(Service))
	root, err := walk(v, reflect.StructField{}, nil)
	if err != nil {
		t.Fatalf("walking service: %s", err)
	}

	fields, err := resolve(root)
	if err != nil {
		t.Fatalf("resolving graph: %s", err)
	}

	displayResolvedGraph(t, fields)
	testResolvedGraph(t, fields)
}

func displayResolvedGraph(t *testing.T, fields []*Field) {
	for _, f := range fields {
		t.Log(f.Path)
	}
}

func testResolvedGraph(t *testing.T, fields []*Field) {
	var found = make(map[Path]struct{})
	var injected = make(map[InjectionKey]struct{})

	for _, field := range fields {
		for _, c := range field.Children {
			if _, ok := found[c.Path]; !ok {
				t.Fatalf("unexpected field a path %s depends on unencountered dependency %s", field.Path, c.Path)
			}
		}

		if key, ok := field.InjectionTarget(); ok {
			if _, ok := injected[key]; !ok {
				t.Fatalf("unexpected field a path %s depends on unencountered injection %s", field.Path, key)
			}
		}

		found[field.Path] = struct{}{}
		if key, ok := field.InjectionSource(); ok {
			injected[key] = struct{}{}
		}
	}
}

type InjectionDependency struct {
	F string
}

type InjectionService struct {
	Source *InjectionDependency `inject-as:"source"`
	Target *InjectionDependency `inject:"source"`
}

func TestField_Inject(t *testing.T) {
	var s InjectionService
	root, err := walk(reflect.ValueOf(&s), reflect.StructField{}, nil)
	if err != nil {
		t.Fatalf("walking service: %s", err)
	}

	if len(root.Children) != 2 {
		t.Fatalf("unexpected number of children: wanted %d, got %d", 2, len(root.Children))
	}

	err = root.Children[1].Inject(root.Children[0])
	if err != nil {
		t.Fatalf("unable to inject source into target: %s", err)
	}

	if s.Source != s.Target {
		t.Fatalf("failed injection")
	}
}
