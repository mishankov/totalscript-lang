// Package main implements the TotalScript CLI interpreter.
package main

import (
	"fmt"
	"os"

	"github.com/mishankov/totalscript-lang/internal/interpreter"
	"github.com/mishankov/totalscript-lang/internal/lexer"
	"github.com/mishankov/totalscript-lang/internal/parser"
	"github.com/mishankov/totalscript-lang/internal/stdlib"
)

const version = "0.1.0"

func main() {
	// Register methods for built-in types
	stdlib.RegisterMethods()

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	arg := os.Args[1]

	// Handle flags
	if arg == "--version" || arg == "-v" {
		fmt.Printf("TotalScript %s\n", version)
		os.Exit(0)
	}

	if arg == "--help" || arg == "-h" {
		printUsage()
		os.Exit(0)
	}

	// Run file
	runFile(arg)
}

func printUsage() {
	fmt.Println("TotalScript - A scripting language with batteries included")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  tsl <file.tsl>    Run a TotalScript file")
	fmt.Println("  tsl --version     Show version")
	fmt.Println("  tsl --help        Show this help message")
}

func runFile(filename string) {
	// Read file
	input, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Lex
	l := lexer.New(string(input))

	// Parse
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		printParserErrors(p.Errors())
		os.Exit(1)
	}

	// Interpret
	env := interpreter.NewEnvironment()
	stdlib.RegisterBuiltins(env)
	result := interpreter.Eval(program, env)

	if result != nil && result.Type() == interpreter.ERROR_OBJ {
		fmt.Fprintf(os.Stderr, "%s\n", result.Inspect())
		os.Exit(1)
	}
}

func printParserErrors(errors []*parser.ParseError) {
	fmt.Fprintf(os.Stderr, "Parser errors:\n")
	for _, err := range errors {
		fmt.Fprintf(os.Stderr, "  %s\n", err.Error())
	}
}
