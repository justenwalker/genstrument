package oteltracer

import (
	"context"
	"fmt"
	"github.com/justenwalker/genstrument"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"strings"
)

type Tracer struct {
	Tracer trace.Tracer
}

type wrappedSpan struct {
	span trace.Span
}

func (s *wrappedSpan) Attribute(key string) genstrument.AttributeSetter {
	return &keyValue{key: key, span: s.span}
}

func (s *wrappedSpan) EndSuccess(_ context.Context) {
	s.span.SetStatus(codes.Ok, "")
	s.span.End()
}

func (s *wrappedSpan) EndError(err error) {
	s.span.RecordError(err)
	s.span.SetStatus(codes.Error, err.Error())
	s.span.End()
}

func (t *Tracer) StartSpan(ctx context.Context, operationName string) (context.Context, genstrument.Span) {
	ctx, span := t.Tracer.Start(ctx, operationName)
	return ctx, &wrappedSpan{span: span}
}

type keyValue struct {
	key  string
	span trace.Span
}

func (k *keyValue) Attribute(key string) genstrument.AttributeSetter {
	return &keyValue{
		key:  strings.Join([]string{k.key, key}, "."),
		span: k.span,
	}
}

func (k *keyValue) Error(err error) {
	k.span.RecordError(err)
	k.span.SetAttributes(attribute.String(k.key, err.Error()))
}

func (k *keyValue) Stringer(v fmt.Stringer) {
	k.span.SetAttributes(attribute.Stringer(k.key, v))
}

func (k *keyValue) String(v string) {
	k.span.SetAttributes(attribute.String(k.key, v))
}

func (k *keyValue) Int64(v int64) {
	k.span.SetAttributes(attribute.Int64(k.key, v))
}

func (k *keyValue) Bool(v bool) {
	k.span.SetAttributes(attribute.Bool(k.key, v))
}

func (k *keyValue) Float64(v float64) {
	k.span.SetAttributes(attribute.Float64(k.key, v))
}

func (k *keyValue) StringSlice(v []string) {
	k.span.SetAttributes(attribute.StringSlice(k.key, v))
}

func (k *keyValue) BoolSlice(v []bool) {
	k.span.SetAttributes(attribute.BoolSlice(k.key, v))
}

func (k *keyValue) Float64Slice(v []float64) {
	k.span.SetAttributes(attribute.Float64Slice(k.key, v))
}

func (k *keyValue) Int64Slice(v []int64) {
	k.span.SetAttributes(attribute.Int64Slice(k.key, v))
}

var _ genstrument.Tracer = (*Tracer)(nil)
var _ genstrument.Span = (*wrappedSpan)(nil)
var _ genstrument.AttributeSetter = (*keyValue)(nil)
