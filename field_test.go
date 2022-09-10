package zconfig

import (
	"reflect"
	"testing"
)

type injectionDependency struct {
	F string
}

func (injectionDependency) do() {}

type doer interface {
	do()
}

func TestField_AddInjectionTarget(t *testing.T) {
	// nominal cases
	for name, tCase := range map[string]reflect.Value{
		"struct pointer": reflect.ValueOf(new(struct {
			Source *injectionDependency `inject-as:"source"`
			Target *injectionDependency `inject:"source"`
		})),
		"scalar pointer": reflect.ValueOf(new(struct {
			Source *int `inject-as:"source"`
			Target *int `inject:"source"`
		})),
		"scalar value": reflect.ValueOf(new(struct {
			Source string `inject-as:"source"`
			Target string `inject:"source"`
		})),
		"interface": reflect.ValueOf(new(struct {
			Source injectionDependency `inject-as:"source"`
			Target doer                `inject:"source"`
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

			err = root.Children[0].AddInjectionTarget(root.Children[1])
			if err != nil {
				t.Fatalf("unable to inject source into target: %s", err)
			}

			if len(root.Children[0].InjectionTargets) != 1 {
				t.Fatalf("unexpected number of injection targets: %d", len(root.Children[0].InjectionTargets))
			}

			if root.Children[0].InjectionTargets[0] != root.Children[1].Value {
				t.Fatalf("unexpected registered target value")
			}
		})
	}

	// errors
	for name, tCase := range map[string]reflect.Value{
		"pointer into value": reflect.ValueOf(new(struct {
			Source *injectionDependency `inject-as:"source"`
			Target injectionDependency  `inject:"source"`
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

			err = root.Children[0].AddInjectionTarget(root.Children[1])
			if err == nil {
				t.Fatal("expected an error, got nil")
			}
		})
	}
}
