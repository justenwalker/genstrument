package example

import "context"

// SimpleService
//
// +genstrument:wrap
type SimpleService interface {
	// +genstrument:attr message message
	// +genstrument:attr result result
	// +genstrument:attr err err
	SayHello(ctx context.Context, message string) (result string, err error)
}
