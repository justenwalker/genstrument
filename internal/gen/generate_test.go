package gen

import (
	"context"
	"testing"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name       string
		inputFile  string
		outputFile string
	}{
		{
			name:       "simple",
			inputFile:  "../../example/simple.go",
			outputFile: "../../example/gen/simple.gen.go",
		},
		{
			name:       "complex",
			inputFile:  "../../example/complex.go",
			outputFile: "../../example/gen/complex.gen.go",
		},
		{
			name:       "external",
			inputFile:  "../../example/external/external.go",
			outputFile: "../../example/external/external.gen.go",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := Generate(context.Background(), tt.inputFile, tt.outputFile, nil)
			if err != nil {
				t.Fatalf("Generate failed: %v", err)
			}
			if err = r.WriteOutput(); err != nil {
				t.Fatalf("WriteOutput failed: %v", err)
			}
		})
	}
}
