package main

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

func Configure(s interface{}) error {
	v := reflect.ValueOf(s)

	if v.Kind() != reflect.Ptr {
		return errors.Errorf("expected pointer to struct, %T given", s)
	}

	if v.Elem().Kind() != reflect.Struct {
		return errors.Errorf("expected pointer to struct, %T given", s)
	}

	root, err := walk(v, reflect.StructField{}, nil)
	if err != nil {
		return errors.Wrap(err, "walking struct")
	}

	DisplayField(root, 0)
	return nil
}

type Field struct {
	Value       reflect.Value
	StructField *reflect.StructField
	Parent      *Field
	Children    []*Field

	Path string
}

func DisplayField(f *Field, l int) {
	fmt.Println(strings.Repeat("\t", l), f.Path)
	for _, c := range f.Children {
		DisplayField(c, l+1)
	}
}

func walk(v reflect.Value, s reflect.StructField, p *Field) (field *Field, err error) {
	field = &Field{
		Value:  v,
		Parent: p,
	}

	if p == nil {
		field.Path = "root"
	} else {
		field.StructField = &s
		field.Path = fmt.Sprintf("%s.%s", p.Path, s.Name)
	}

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return field, nil
	}

	for i := 0; i < v.Type().NumField(); i++ {
		child, err := walk(v.Field(i), v.Type().Field(i), field)
		if err != nil {
			return nil, err
		}

		field.Children = append(field.Children, child)
	}

	return field, nil
}

func log(msg ...interface{}) {
	fmt.Println(msg...)
}
