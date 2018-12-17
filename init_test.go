package zconfig

import "testing"

type initTest struct {
	Initialized bool
}

func (i *initTest) Init() error {
	i.Initialized = true
	return nil
}

func TestInitialize(t *testing.T) {
	initMe := new(initTest)
	err := NewRepository(Initialize).Configure(initMe)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !initMe.Initialized {
		t.Fatal("struct was not initialized as expected")
	}
}
