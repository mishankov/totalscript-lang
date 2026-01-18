// Package main implements the TotalScript CLI interpreter.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/mishankov/totalscript-lang/internal/interpreter"
	"github.com/mishankov/totalscript-lang/internal/lexer"
	"github.com/mishankov/totalscript-lang/internal/parser"
	"github.com/mishankov/totalscript-lang/internal/stdlib"
)

const version = "0.1.0"

func main() {
	// Register methods for built-in types
	stdlib.RegisterMethods()

	// Define flags
	noWatch := flag.Bool("no-watch", false, "Disable live reloading (watch mode is enabled by default)")
	versionFlag := flag.Bool("version", false, "Show version")
	vFlag := flag.Bool("v", false, "Show version (shorthand)")
	helpFlag := flag.Bool("help", false, "Show help message")
	hFlag := flag.Bool("h", false, "Show help message (shorthand)")

	flag.Parse()

	// Handle flags
	if *versionFlag || *vFlag {
		fmt.Printf("TotalScript %s\n", version)
		os.Exit(0)
	}

	if *helpFlag || *hFlag {
		printUsage()
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	filename := args[0]

	// Get absolute path for the file
	absPath, err := filepath.Abs(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving file path: %v\n", err)
		os.Exit(1)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: file does not exist: %s\n", absPath)
		os.Exit(1)
	}

	// If watch mode is disabled, just run once
	if *noWatch {
		runFile(absPath, true)
		return
	}

	// Run with live reloading
	runWithWatch(absPath)
}

func printUsage() {
	fmt.Println("TotalScript - A scripting language with batteries included")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  tsl <file.tsl>        Run a TotalScript file with live reloading")
	fmt.Println("  tsl --no-watch <file> Run without live reloading")
	fmt.Println("  tsl --version         Show version")
	fmt.Println("  tsl --help            Show this help message")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --no-watch    Disable live reloading (enabled by default)")
	fmt.Println("  -v, --version Show version information")
	fmt.Println("  -h, --help    Show this help message")
	fmt.Println()
	fmt.Println("Live Reloading:")
	fmt.Println("  By default, tsl watches the main file and all imported files")
	fmt.Println("  for changes and automatically reloads when changes are detected.")
}

func runWithWatch(absPath string) {
	// Create file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating file watcher: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := watcher.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing watcher: %v\n", err)
		}
	}()

	// Add main file to watcher
	if err := watcher.Add(absPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error watching file: %v\n", err)
		// Close watcher before exit
		if closeErr := watcher.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Error closing watcher: %v\n", closeErr)
		}
		os.Exit(1)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Channel to debounce file changes
	debounce := time.NewTimer(0)
	debounce.Stop()

	// Track currently watched files
	watchedFiles := make(map[string]bool)
	watchedFiles[absPath] = true

	// Run the file initially
	clearScreen()
	fmt.Printf("Running %s (live reload enabled, press Ctrl+C to exit)\n", filepath.Base(absPath))
	fmt.Println("---")
	runFileAndUpdateWatcher(absPath, watcher, watchedFiles)

	// Watch for changes
	for {
		select {
		case <-sigChan:
			// Graceful shutdown
			fmt.Println("\nExiting...")
			return

		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			// Only respond to write and create events
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				// Debounce: reset timer on each event
				debounce.Reset(100 * time.Millisecond)
			}

		case <-debounce.C:
			// Debounce timer fired, reload the file
			clearScreen()
			fmt.Printf("File changed, reloading %s...\n", filepath.Base(absPath))
			fmt.Println("---")
			runFileAndUpdateWatcher(absPath, watcher, watchedFiles)

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Fprintf(os.Stderr, "Watcher error: %v\n", err)
		}
	}
}

func runFileAndUpdateWatcher(absPath string, watcher *fsnotify.Watcher, watchedFiles map[string]bool) {
	// Clear module cache before reloading to pick up changes in imported files
	interpreter.ClearModuleCache()

	// Run the file
	runFile(absPath, false)

	// Get all loaded file modules after execution
	loadedModules := interpreter.GetLoadedFileModules()

	// Add newly loaded modules to watcher
	for _, modulePath := range loadedModules {
		if !watchedFiles[modulePath] {
			if err := watcher.Add(modulePath); err != nil {
				// Don't fail on watcher errors, just log them
				fmt.Fprintf(os.Stderr, "Warning: could not watch imported file %s: %v\n", modulePath, err)
			} else {
				watchedFiles[modulePath] = true
			}
		}
	}

	// Remove modules from watcher that are no longer loaded
	loadedSet := make(map[string]bool)
	for _, modulePath := range loadedModules {
		loadedSet[modulePath] = true
	}
	for filePath := range watchedFiles {
		// Don't remove the main file from watcher
		if filePath == absPath {
			continue
		}
		// If file is no longer in loaded modules, remove from watcher
		if !loadedSet[filePath] {
			if err := watcher.Remove(filePath); err != nil {
				// Ignore errors when removing
				continue
			}
			delete(watchedFiles, filePath)
		}
	}
}

func runFile(absPath string, exitOnError bool) {
	// Read file
	input, err := os.ReadFile(absPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		if exitOnError {
			os.Exit(1)
		}
		return
	}

	// Lex
	l := lexer.New(string(input))

	// Parse
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		printParserErrors(p.Errors())
		if exitOnError {
			os.Exit(1)
		}
		return
	}

	// Interpret
	env := interpreter.NewEnvironment()
	env.SetCurrentFile(absPath) // Set current file for module resolution
	stdlib.RegisterBuiltins(env)
	result := interpreter.Eval(program, env)

	if result != nil && result.Type() == interpreter.ErrorObj {
		fmt.Fprintf(os.Stderr, "%s\n", result.Inspect())
		if exitOnError {
			os.Exit(1)
		}
		return
	}
}

func printParserErrors(errors []*parser.ParseError) {
	fmt.Fprintf(os.Stderr, "Parser errors:\n")
	for _, err := range errors {
		fmt.Fprintf(os.Stderr, "  %s\n", err.Error())
	}
}

func clearScreen() {
	// ANSI escape code to clear screen and move cursor to top-left
	fmt.Print("\033[H\033[2J")
}
