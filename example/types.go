package example

type unexportedType struct{}

type Constraint[T any] interface {
	Func(T) error
}

type Name string

type ServiceType struct {
	FooBarBaz string
}
