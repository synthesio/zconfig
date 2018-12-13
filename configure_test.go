package zconfig

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
)

type Service struct {
	Workers        int              `key:"workers"`
	Dependency     SimpleDependency `key:"dependency" inject-as:"dependency"`
	Injected       SimpleDependency `inject:"dependency"`
	notExported    SimpleDependency
	notExportedPtr *SimpleDependency
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

	expected := map[Path]struct{}{
		"root":                {},
		"root.Workers":        {},
		"root.Dependency":     {},
		"root.Dependency.Foo": {},
	}

	for _, field := range fields {
		if _, ok := expected[field.Path]; ok {
			delete(expected, field.Path)
		} else {
			t.Errorf("got unexpected field %s", field.Path)
		}
	}

	for path := range expected {
		t.Errorf("expected field %s not found", path)
	}
}

func displayResolvedGraph(t *testing.T, fields []*Field) {
	for _, f := range fields {
		t.Log(f.Path)
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
	var deps = dependencies{
		"A": {"B": struct{}{}, "C": struct{}{}},
		"B": {"D": struct{}{}},
		"C": {"E": struct{}{}},
		"D": {"C": struct{}{}},
		"E": {"B": struct{}{}},
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

func TestDependencies(t *testing.T) {
	var deps = dependencies{}
	a := &Field{
		Path: "A",
	}
	b := &Field{
		Path: "B",
	}
	deps.add(a, b)
	deps.add(b, a)

	aDeps, ok := deps[a.Path]
	if !ok {
		t.Fatal("a path not found")
	}

	if _, ok := aDeps[b.Path]; !ok {
		t.Fatal("b dependency not found")
	}

	deps.remove(a.Path)

	if _, ok = deps[a.Path]; ok {
		t.Fatal("a path still found")
	}

	bDeps, ok := deps[b.Path]
	if !ok {
		t.Fatal("b path not found")
	}
	if _, ok := bDeps[a.Path]; ok {
		t.Fatal("a dependency found")
	}
}

func TestHooks_Execution(t *testing.T) {
	executed := false

	testHook := func(field *Field) error {
		executed = true
		return nil
	}

	testRepo := Repository{
		hooks: []Hook{testHook},
	}

	err := testRepo.Configure(new(Service))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !executed {
		t.Fatal("hook was not executed")
	}
}

func TestHooks_AllFields(t *testing.T) {
	fields := map[Path]bool{
		"root":                false,
		"root.Workers":        false,
		"root.Dependency":     false,
		"root.Dependency.Foo": false,
	}

	testHook := func(field *Field) error {
		visited, ok := fields[field.Path]
		if !ok {
			t.Fatalf("unexpected field: %s", field.Path)
		}
		if visited {
			t.Fatalf("field %s already visited", field.Path)
		}

		fields[field.Path] = true

		return nil
	}

	testRepo := Repository{
		hooks: []Hook{testHook},
	}

	err := testRepo.Configure(new(Service))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHooks_Error(t *testing.T) {
	expected := errors.New("an error")
	testHook := func(field *Field) error {
		return expected
	}

	testRepo := Repository{
		hooks: []Hook{testHook},
	}

	err := testRepo.Configure(new(Service))
	if err == nil {
		t.Fatalf("expected an error, got nil")
	}
	if errors.Cause(err) != expected {
		t.Fatalf("unexpected error: %v", err)
	}
}
