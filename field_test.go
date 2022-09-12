package zconfig

import (
	"reflect"
	"testing"
)

type InjectionDependency struct {
	F string
}

func (InjectionDependency) do() {}

type InjectionService struct {
	Source *InjectionDependency `inject-as:"source"`
	Target *InjectionDependency `inject:"source"`
}

func TestField_Inject(t *testing.T) {
	// nominal cases
	t.Run("struct pointer", func(t *testing.T) {
		s := InjectionService{
			Source: &InjectionDependency{
				F: "Foo",
			},
		}
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
	})

	t.Run("scalar pointer", func(t *testing.T) {
		var s struct {
			Source *string `inject-as:"source"`
			Target *string `inject:"source"`
		}
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
	})

	t.Run("interface", func(t *testing.T) {
		type doer interface {
			do()
		}

		s := struct {
			Source *InjectionDependency `inject-as:"source"`
			Target doer                 `inject:"source"`
		}{
			Source: &InjectionDependency{
				F: "Foo",
			},
		}
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
	})

	// errors
	for name, tCase := range map[string]reflect.Value{
		"pointer into value": reflect.ValueOf(new(struct {
			Source *InjectionDependency `inject-as:"source"`
			Target InjectionDependency  `inject:"source"`
		})),
		"value into pointer": reflect.ValueOf(new(struct {
			Source int  `inject-as:"source"`
			Target *int `inject:"source"`
		})),
		"type mismatch": reflect.ValueOf(new(struct {
			Source int    `inject-as:"source"`
			Target string `inject:"source"`
		})),
	} {
		t.Run(name, func(t *testing.T) {
			root, err := walk(tCase, reflect.StructField{}, nil)
			if err != nil {
				t.Fatalf("walking service: %s", err)
			}

			if len(root.Children) != 2 {
				t.Fatalf("unexpected number of children: wanted %d, got %d", 2, len(root.Children))
			}

			err = root.Children[0].Inject(root.Children[1])
			if err == nil {
				t.Fatal("expected an error, got nil")
			}
		})
	}
}
