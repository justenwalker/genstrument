package genstrument

import (
	"context"
	"fmt"
)

type AttributeSetter interface {
	Error(err error)
	Stringer(v fmt.Stringer)
	String(v string)
	Int64(v int64)
	Bool(v bool)
	Float64(v float64)
	StringSlice(v []string)
	BoolSlice(v []bool)
	Float64Slice(v []float64)
	Int64Slice(v []int64)
	Attribute(key string) AttributeSetter
}

type Span interface {
	EndSuccess(ctx context.Context)
	EndError(err error)
	Attribute(key string) AttributeSetter
}

type Tracer interface {
	StartSpan(ctx context.Context, operationName string) (context.Context, Span)
}
