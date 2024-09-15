// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"github.com/justenwalker/genstrument/example"
	"github.com/justenwalker/genstrument/example/gen"
	"github.com/justenwalker/genstrument/example/oteltracer"
	"github.com/justenwalker/genstrument/example/types"
	"github.com/justenwalker/genstrument/example/types/dot"
	gopkg "github.com/justenwalker/genstrument/example/types/go-pkg"
	"log"

	"github.com/go-logr/stdr"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

//go:generate go run github.com/justenwalker/genstrument/cmd/gen -input ../simple.go -output ../gen/simple.gen.go
//go:generate go run github.com/justenwalker/genstrument/cmd/gen -input ../complex.go -output ../gen/complex.gen.go
//go:generate go run github.com/justenwalker/genstrument/cmd/gen -input ../external/external.go -output ../external/external.gen.go

var (
	fooKey     = attribute.Key("ex.com/foo")
	barKey     = attribute.Key("ex.com/bar")
	anotherKey = attribute.Key("ex.com/another")
)

var tp *sdktrace.TracerProvider

// initTracer creates and registers trace provider instance.
func initTracer() error {
	exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return fmt.Errorf("failed to initialize stdouttrace exporter: %w", err)
	}
	bsp := sdktrace.NewBatchSpanProcessor(exp)
	tp = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tp)
	return nil
}

func main() {
	// Set logging level to info to see SDK status messages
	stdr.SetVerbosity(5)

	// initialize trace provider.
	if err := initTracer(); err != nil {
		log.Panic(err)
	}

	// Create a named tracer with package path as its name.
	tracer := tp.Tracer("go.opentelemetry.io/otel/example/namedtracer")
	ctx := context.Background()
	defer func() { _ = tp.Shutdown(ctx) }()

	m0, _ := baggage.NewMemberRaw(string(fooKey), "foo1")
	m1, _ := baggage.NewMemberRaw(string(barKey), "bar1")
	b, _ := baggage.New(m0, m1)
	ctx = baggage.ContextWithBaggage(ctx, b)
	t := &oteltracer.Tracer{
		Tracer: tracer,
	}
	svc := &Service{}
	cc := gen.InstrumentComplexService(t, svc)
	ss := gen.InstrumentSimpleService(t, svc)
	_, _ = ss.SayHello(ctx, "hello")
	cc.FuncNoError(ctx)
	_, _ = svc.FuncArray(ctx, "str1", example.ServiceType{
		FooBarBaz: "foobarbaz",
	})
	_, _ = svc.FuncSlice(ctx, "str2", example.ServiceType{
		FooBarBaz: "foobarbaz",
	})
	_, _ = svc.FuncGoPkg2(ctx, gopkg.GoType2{
		Bar: "bar",
	})
	_, _ = svc.FuncPackageType(ctx, types.MyType{
		Foo:    "foo",
		Bar:    "bar",
		Secret: "secret",
	})
	_, _ = svc.FuncDotTypes(ctx, "namename", dot.Type1Dot{
		Foo: "foo",
	}, dot.Type2Dot{
		Bar: "bar",
	})
	_, _ = svc.FuncMyDupeType(ctx, types.MyType{Bar: "1", Foo: "2"})
	_, _ = gen.TraceMyFunction(t)(ctx, example.ServiceType{FooBarBaz: "foo-bar-baz"}, dot.Type1Dot{Foo: "foo-1"}, dot.Type2Dot{Bar: "bar-1"}, types.MyType{
		Foo:    "foo",
		Bar:    "bar",
		Secret: "secret",
	})
	_, _ = gen.TraceGenericFunction[string, *string, int](t)(ctx, "", nil, nil, 1)
}
