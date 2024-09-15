package gen

import (
	"context"
	"testing"
)

func TestGenerate(t *testing.T) {
	//r, err := Generate(context.Background(), "../../example/complex.go", "../../example/gen/complex.gen.go", nil)
	//if err != nil {
	//	t.Fatalf("Generate failed: %v", err)
	//}
	//if err = r.WriteOutput(); err != nil {
	//	t.Fatalf("WriteOutput failed: %v", err)
	//}
	r, err := Generate(context.Background(), "../../example/simple.go", "../../example/gen/simple.gen.go", nil)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if err = r.WriteOutput(); err != nil {
		t.Fatalf("WriteOutput failed: %v", err)
	}
}
