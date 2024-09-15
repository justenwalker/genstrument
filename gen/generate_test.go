package main

import (
	"context"
	"github.com/sebdah/goldie/v2"
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
			inputFile:  "../example/simple.go",
			outputFile: "../example/gen/simple.gen.go",
		},
		{
			name:       "complex",
			inputFile:  "../example/complex.go",
			outputFile: "../example/gen/complex.gen.go",
		},
		{
			name:       "external",
			inputFile:  "../example/external/external.go",
			outputFile: "../example/external/external.gen.go",
		},
	}
	g := goldie.New(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r, err := Generate(context.Background(), tt.inputFile, tt.outputFile, nil)
			if err != nil {
				t.Fatalf("Generate failed: %v", err)
			}
			g.Assert(t, tt.name, r.Content)
		})
	}
}
