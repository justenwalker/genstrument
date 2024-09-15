package example

import (
	"cmp"
	"context"
	"fmt"
	"genstrument/example/types"
	dupe "genstrument/example/types"
	. "genstrument/example/types/dot"
	"genstrument/example/types/go-pkg"
	"github.com/justenwalker/genstrument"
)

// GenericService
//
// +genstrument:wrap
// +genstrument:constructor Trace
// +genstrument:prefix traced
type GenericService[T any, PT cmp.Ordered] interface {
	FuncIsGeneric(ctx context.Context, t T) (PT, error)
}

// ComplexService
//
// +genstrument:wrap
// +genstrument:constructor Instrument
// +genstrument:prefix instrumented
type ComplexService interface {
	FuncNoError(ctx context.Context)
	// +genstrument:attr key1 str StringAttributeSetter
	// +genstrument:attr key2 st ServiceTypeSetter
	// +genstrument:attr result res0 AnyTypeSetter
	// +genstrument:attr error err AnyTypeSetter
	FuncArray(ctx context.Context, str string, st ServiceType) (res0 [32]byte, err error)
	// +genstrument:attr key1 name StringAttributeSetter
	// +genstrument:attr key2 st ServiceTypeSetter
	FuncSlice(ctx context.Context, name Name, st ServiceType) ([]byte, error)
	// +genstrument:op goPkg2
	// +genstrument:attr key1 name gopkg.GoType2Attr
	FuncGoPkg2(ctx context.Context, mt gopkg.GoType2) (bool, error)
	// +genstrument:op packageType
	// +genstrument:attr type myType types.MyTypeAttr
	FuncPackageType(ctx context.Context, myType types.MyType) (int64, error)
	// +genstrument:op dots
	// +genstrument:attr name name
	// +genstrument:attr dot1 d1 Type1Attr
	// +genstrument:attr dot2 d2 Type2Attr
	FuncDotTypes(ctx context.Context, name Name, d1 Type1Dot, d2 Type2Dot) (string, error)
	// +genstrument:op dupes
	// +genstrument:attr mine myType dupe.MyTypeAttr
	FuncMyDupeType(ctx context.Context, myType dupe.MyType) (string, error)
}

// MyFunction
//
// +genstrument:wrap
// +genstrument:prefix Trace
// +genstrument:op func1
// +genstrument:attr key1 s AnyTypeSetter
// +genstrument:attr key2 d1 AnyTypeSetter
// +genstrument:attr key3 d2 AnyTypeSetter
// +genstrument:attr key4 myType AnyTypeSetter
func MyFunction(ctx context.Context, s ServiceType, d1 Type1Dot, d2 Type2Dot, myType types.MyType) ([]byte, error) {
	return nil, nil
}

// GenericFunction
//
// +genstrument:wrap
// +genstrument:prefix Trace
// +genstrument:attr key1 t AnyTypeSetter
// +genstrument:attr key2 tr AnyTypeSetter
// +genstrument:attr key3 pt AnyTypeSetter
// +genstrument:attr key4 err AnyTypeSetter
func GenericFunction[T ~string, PT *T, PTT cmp.Ordered](ctx context.Context, t T, tr PT, pt PT, err PTT) (ServiceType, error) {
	return ServiceType{}, nil
}

// GenericTypeConstraints
//
// +genstrument:wrap
// +genstrument:prefix Observe
func GenericTypeConstraints[P any, S interface{ ~[]byte | string }, ES ~[]E, E any, C Constraint[int], O cmp.Ordered](ctx context.Context, p P, es ES, e E, c C, o O) (S, error) {
	var zero S
	return zero, nil
}

func StringAttributeSetter[S ~string](s S, attr genstrument.AttributeSetter) {
	attr.String(string(s))
}

func ServiceTypeSetter(st ServiceType, attr genstrument.AttributeSetter) {
	attr.Attribute("foobarbaz").String(st.FooBarBaz)
}

func AnyTypeSetter(v any, attr genstrument.AttributeSetter) {
	attr.String(fmt.Sprintf("%v", v))
}
