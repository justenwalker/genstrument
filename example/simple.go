package example

import (
	"context"
)

// SimpleService
//
// +genstrument:wrap
type SimpleService interface {
	// +genstrument:op helloOp
	// +genstrument:attr message message
	// +genstrument:attr result result
	// +genstrument:attr err err
	SayHello(ctx context.Context, message string) (result string, err error)
}

// MyFunction
//
// +genstrument:wrap
// +genstrument:prefix Trace
// +genstrument:op helloOp
// +genstrument:attr message message
// +genstrument:attr result result
// +genstrument:attr err err
func SimpleFunction(message string) (result string, err error) {
	return message, nil
}
