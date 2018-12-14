package zconfig

import (
	"reflect"
	"testing"
)

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
