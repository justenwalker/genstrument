package external

import (
	"cmp"
	"context"
	"github.com/justenwalker/genstrument/example"
	"github.com/justenwalker/genstrument/example/types"
	"github.com/justenwalker/genstrument/example/types/dot"
)

// SimpleService
//
// +genstrument:wrap
// +genstrument:external example.SimpleService
type SimpleService interface {
	// +genstrument:op goPkg2
	// +genstrument:attr key1 name gopkg.GoType1Attr
	SayHello(ctx context.Context, message string) (result string, err error)
}

// GenericService
//
// +genstrument:wrap
// +genstrument:external example.GenericService
// +genstrument:constructor Trace
// +genstrument:prefix traced
type GenericService[T any, PT cmp.Ordered] interface {
	FuncIsGeneric(ctx context.Context, t T) (PT, error)
}

// MyFunction
//
// +genstrument:wrap
// +genstrument:external example.MyFunction
// +genstrument:prefix Trace
// +genstrument:op func1
// +genstrument:attr key1 s example.AnyTypeSetter
// +genstrument:attr key2 d1 example.AnyTypeSetter
// +genstrument:attr key3 d2 example.AnyTypeSetter
// +genstrument:attr key4 myType example.AnyTypeSetter
func MyFunction(ctx context.Context, s example.ServiceType, d1 dot.Type1Dot, d2 dot.Type2Dot, myType types.MyType) ([]byte, error) {
	return example.MyFunction(ctx, s, d1, d2, myType) // compile-time check
}

// GenericFunction
//
// +genstrument:wrap
// +genstrument:external example.GenericFunction
// +genstrument:prefix Trace
// +genstrument:attr key1 t example.AnyTypeSetter
// +genstrument:attr key2 tr example.AnyTypeSetter
// +genstrument:attr key3 pt example.AnyTypeSetter
// +genstrument:attr key4 err example.AnyTypeSetter
func GenericFunction[T ~string, PT *T, PTT cmp.Ordered](ctx context.Context, t T, tr PT, pt PT, err PTT) (example.ServiceType, error) {
	return example.GenericFunction[T, PT](ctx, t, tr, pt, err)
}
