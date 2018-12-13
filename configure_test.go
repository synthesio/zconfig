package zconfig

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

type Service struct {
	Workers    int              `key:"workers"`
	Dependency SimpleDependency `key:"dependency"`
}

type SimpleDependency struct {
	Foo int `key:"foo"`
}

func TestWalk(t *testing.T) {
	var v = reflect.ValueOf(new(Service))
	root, err := walk(v, reflect.StructField{}, nil)
	if err != nil {
		t.Errorf("walking service: %s", err)
		return
	}

	var expected = &Field{Path: "root", Children: []*Field{
		{Path: "root.Workers"},
		{Path: "root.Dependency", Children: []*Field{
			{Path: "root.Dependency.Foo"},
		}},
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
		t.Errorf("invalid value for path: expected %s, got %s", expected.Path, actual.Path)
		return
	}

	if actual.Parent != parent {
		t.Errorf("invalud parent for path %s: wanted %p, got %p", actual.Path, parent, actual.Parent)
		return
	}

	if len(actual.Children) != len(expected.Children) {
		t.Errorf("invalid number of children for path %s: wanted %d, got %d", actual.Path, len(actual.Children), len(expected.Children))
		return
	}

	for i := 0; i < len(actual.Children); i++ {
		testGraph(t, actual.Children[i], expected.Children[i], actual)
	}
}

func TestResolve(t *testing.T) {
	var v = reflect.ValueOf(new(Service))
	root, err := walk(v, reflect.StructField{}, nil)
	if err != nil {
		t.Errorf("walking service: %s", err)
		return
	}

	fields, err := resolve(root)
	if err != nil {
		t.Errorf("resolving graph: %s", err)
		return
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
				t.Errorf("unexpected field a path %s depends on unencountered dependency %s", field.Path, c.Path)
				return
			}
		}

		if key, ok := field.InjectionTarget(); ok {
			if _, ok := injected[key]; !ok {
				t.Errorf("unexpected field a path %s depends on unencountered injection %s", field.Path, key)
				return
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
		t.Errorf("walking service: %s", err)
		return
	}

	if len(root.Children) != 2 {
		t.Errorf("unexpected number of children: wanted %d, got %d", 2, len(root.Children))
		return
	}

	err = root.Children[1].Inject(root.Children[0])
	if err != nil {
		t.Errorf("unable to inject source into target: %s", err)
		return
	}

	if s.Source != s.Target {
		t.Errorf("failed injection")
		return
	}
}

func TestNewCycleError(t *testing.T) {
	// This is the dependency map of the following structure, which is
	// valid Go.
	// type CycleService struct {
	// 	B *CycleDepB `inject-as:"B"`
	// 	C *CycleDepC `inject-as:"C"`
	// }
	// type CycleDepB struct {
	// 	D *CycleDepC `inject:"C"`
	// }
	// type CycleDepC struct {
	// 	E *CycleDepB `inject:"B"`
	// }
	var deps = map[Path][]*Field{
		"A": []*Field{{Path: "B"}, {Path: "C"}},
		"B": []*Field{{Path: "D"}},
		"C": []*Field{{Path: "E"}},
		"D": []*Field{{Path: "C"}},
		"E": []*Field{{Path: "B"}},
	}

	var err error
	var done = make(chan struct{})
	go func() {
		err = newCycleError(deps)
		close(done)
	}()

	select {
	case <-time.After(1 * time.Second):
		t.Fatalf("timeout computing cycle")
	case <-done:
	}

	t.Log(err)
}
