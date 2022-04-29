package zconfig

import (
	"context"
	"testing"
)

type initTest struct {
	Initialized bool
}

func (i *initTest) Init(ctx context.Context) error {
	i.Initialized = true
	return nil
}

type initTestDeprecated struct {
	init initTest
}

func (i *initTestDeprecated) Init() error {
	return i.init.Init(context.Background())
}

func TestInitialize(t *testing.T) {
	t.Run("with context", func(t *testing.T) {
		initMe := new(initTest)
		err := NewProcessor(Initialize).Process(context.Background(), initMe)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !initMe.Initialized {
			t.Fatal("struct was not initialized as expected")
		}
	})

	t.Run("without context", func(t *testing.T) {
		initMe := new(initTestDeprecated)
		err := NewProcessor(Initialize).Process(context.Background(), initMe)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !initMe.init.Initialized {
			t.Fatal("struct was not initialized as expected")
		}
	})
}
