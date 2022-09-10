package zconfig

import (
	"context"
	"testing"
)

type injectionMock struct {
	called bool
}

func (n *injectionMock) call() {
	n.called = true
}

type caller interface {
	call()
}

func TestInject(t *testing.T) {
	t.Run("pointers", func(t *testing.T) {
		var s struct {
			Source *string `inject-as:"source"`
			Target *string `inject:"source"`
		}
		err := NewProcessor(Inject).Process(context.Background(), &s)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if s.Source != s.Target {
			t.Fatal("target was not injected with source")
		}
	})

	t.Run("values", func(t *testing.T) {
		const helloWorld = "hello, world"

		var s struct {
			Source string `inject-as:"source"`
			Target string `inject:"source"`
		}
		s.Source = helloWorld

		err := NewProcessor(Inject).Process(context.Background(), &s)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if s.Target != helloWorld {
			t.Fatal("target was not injected with source")
		}
	})

	t.Run("interface", func(t *testing.T) {
		var s struct {
			Source *injectionMock `inject-as:"source"`
			Target caller         `inject:"source"`
		}
		err := NewProcessor(Inject).Process(context.Background(), &s)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if s.Target == nil {
			t.Fatal("source was not injected into target")
		}

		if s.Source.called {
			t.Fatal("unexepected starting value for source.called field")
		}
		s.Target.call()
		if !s.Source.called {
			t.Fatal("method was not called on source")
		}
	})
}
