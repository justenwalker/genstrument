package main

import (
	"context"
	"fmt"
	"github.com/justenwalker/genstrument/example"
	"github.com/justenwalker/genstrument/example/types"
	"github.com/justenwalker/genstrument/example/types/dot"
	gopkg "github.com/justenwalker/genstrument/example/types/go-pkg"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var _ example.ComplexService = (*Service)(nil)
var _ example.SimpleService = (*Service)(nil)

type Service struct {
}

func (s *Service) SayHello(ctx context.Context, message string) (result string, err error) {
	return message, nil
}

func (s *Service) FuncArray(ctx context.Context, str string, st example.ServiceType) (res0 [32]byte, err error) {
	copy(res0[:], str)
	return
}

func (s *Service) FuncSlice(ctx context.Context, name example.Name, st example.ServiceType) ([]byte, error) {
	return []byte(name), nil
}

func (s *Service) FuncNoError(ctx context.Context) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("foo", "bar"))
}

func (s *Service) FuncGoPkg2(ctx context.Context, mt gopkg.GoType2) (bool, error) {
	return mt.Bar == "1", nil
}

func (s *Service) FuncPackageType(ctx context.Context, myType types.MyType) (int64, error) {
	return 123, nil
}

func (s *Service) FuncDotTypes(ctx context.Context, name example.Name, d1 dot.Type1Dot, d2 dot.Type2Dot) (string, error) {
	return string(name), fmt.Errorf("foo")
}

func (s *Service) FuncMyDupeType(ctx context.Context, myType types.MyType) (string, error) {
	return myType.Bar, nil
}
