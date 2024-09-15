# genstrument

genstrument is a Go Code Generator for instrumenting interfaces and functions.

It takes interface and function definitions with special doc comments and creates wrapper implementations
which add instrumentation.

## Example

### Source

```go
// SimpleService
//
// +genstrument:wrap
type SimpleService interface {
    // +genstrument:attr message message
    // +genstrument:attr result result
    // +genstrument:attr err err
    SayHello(ctx context.Context, message string) (result string, err error)
}
```

### Generated

```go
// InstrumentSimpleService adds APM traces around the wrapped example.SimpleService using the provided tracer.
func InstrumentSimpleService(tracer genstrument.Tracer, wrapped example.SimpleService) example.SimpleService {
	return &instrumentedSimpleService{
		tracer:  tracer,
		wrapped: wrapped,
	}
}

type instrumentedSimpleService struct {
	wrapped example.SimpleService
	tracer  genstrument.Tracer
}

func (w *instrumentedSimpleService) SayHello(ctx context.Context, message string) (result string, err error) {
	// Start Span
	var span genstrument.Span
	ctx, span = w.tracer.StartSpan(ctx, "example.SimpleService:SayHello")
	// Set Input Attributes
	genstrument.SetStringAttribute(message, span.Attribute("message"))

// call Wrapped Function
	result, err = w.wrapped.SayHello(ctx, message)
	// Finish Span with Error
	if err != nil {
		span.EndError(err)
		return
	}
	// Set Return Attributes
	genstrument.SetStringAttribute(result, span.Attribute("result"))
	genstrument.SetErrorAttribute(err, span.Attribute("err"))

	// Finish Span with Success
	span.EndSuccess(ctx)
	return
}
```

For a full example, see the [Example Code](./example).