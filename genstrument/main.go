package main

import (
	"context"
	"flag"
	"log"
	"os"
)

func main() {
	ctx := context.Background()
	var (
		inFile  string
		outFile string
	)
	flag.StringVar(&inFile, "input", "", "Input File to parse")
	flag.StringVar(&outFile, "output", "", "Output file to write generated code.")
	flag.Parse()
	if inFile == "" || outFile == "" {
		log.Println("Must provide an -input and -output flag")
		flag.Usage()
		os.Exit(1)
	}
	res, err := Generate(ctx, inFile, outFile, &Options{})
	if err != nil {
		log.Fatalf("Generate failed: %v", err)
	}
	if err = res.WriteOutput(); err != nil {
		log.Fatalf("write '%s' failed: %v", res.OutputFile, err)
	}
}
