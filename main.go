package main

type Service struct {
	Workers int        `key:"workers"`
	Dep     Dependency `key:"dep"`
}

type Dependency struct {
	Foo int `key:"foo"`
}

func main() {
	var s Service
	err := Configure(&s)
	if err != nil {
		panic(err)
	}
}
