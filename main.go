package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/apparentlymart/go-indeps/indeps"
)

func main() {
	err := realMain(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func realMain(args []string) error {
	if len(args) != 2 {
		return errors.New("Usage: go-indeps <pkg>")
	}

	pkgPath := args[1]
	graph, err := indeps.AnalyzePackagePath(pkgPath)
	if err != nil {
		return err
	}

	os.Stdout.WriteString("digraph G {\n")
	os.Stdout.WriteString("    node [shape=box,fontname=Helvetica,fontsize=10,margin=\"0.11,0\"]\n")
	for _, node := range graph.Nodes() {
		fmt.Fprintf(os.Stdout, "    %q;\n", node)
	}
	for _, edge := range graph.Edges() {
		fmt.Fprintf(os.Stdout, "    %q -> %q;\n", edge.From, edge.To)
	}
	os.Stdout.WriteString("}\n")

	return nil
}
