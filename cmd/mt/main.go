package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	engine "github.com/xxnuo/MTranCore/engine"
	"github.com/xxnuo/MTranCore/internal/wasm"
)

var (
	modelDir  = flag.String("model", "", "Model directory path (required)")
	modelPath = flag.String("model-file", "", "Model file path (.bin)")
	shortlist = flag.String("shortlist", "", "Lexical shortlist file path (lex*.bin)")
	vocab     = flag.String("vocab", "", "Vocabulary file path (.spm)")
	vocabSrc  = flag.String("vocab-src", "", "Source vocabulary file path (.spm)")
	vocabTrg  = flag.String("vocab-trg", "", "Target vocabulary file path (.spm)")
	text      = flag.String("text", "", "Text to translate (required in non-REPL mode)")
	html      = flag.Bool("html", false, "Treat input as HTML")
	cacheSize = flag.Uint("cache", 1024, "Cache size for translation")
	repl      = flag.Bool("r", false, "Run in REPL (interactive) mode")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "A simple command-line translation tool using Bergamot translator.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Using model directory (auto-discover files):\n")
		fmt.Fprintf(os.Stderr, "  %s -model ./models/enzh -text \"Hello, world!\"\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Using individual files:\n")
		fmt.Fprintf(os.Stderr, "  %s -model-file model.bin -shortlist lex.bin \\\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "    -vocab-src src.spm -vocab-trg trg.spm -text \"Hello!\"\n\n")
		fmt.Fprintf(os.Stderr, "  # HTML translation:\n")
		fmt.Fprintf(os.Stderr, "  %s -model ./models/enzh -html -text \"<p>Hello</p>\"\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # REPL (interactive) mode:\n")
		fmt.Fprintf(os.Stderr, "  %s -model ./models/enzh -r\n\n", os.Args[0])
	}

	flag.Parse()

	// Validate required parameters
	if !*repl && *text == "" {
		fmt.Fprintf(os.Stderr, "Error: -text is required in non-REPL mode\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if *modelDir == "" && (*modelPath == "" || *shortlist == "" || *vocabSrc == "" || *vocabTrg == "" || *vocab == "") {
		fmt.Fprintf(os.Stderr, "Error: Either -model (model directory) or all of (-model-file, -shortlist, -vocab-src, -vocab-trg) must be specified\n\n")
		flag.Usage()
		os.Exit(1)
	}

	ctx := context.Background()

	// Create translator
	translator, cleanup, err := createTranslator(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating translator: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	// Choose mode: REPL or single translation
	if *repl {
		runREPL(ctx, translator)
	} else {
		// Translate
		req := engine.TranslationRequest{
			Text: *text,
			Options: engine.TranslationOptions{
				HTML: *html,
			},
		}

		result, err := translator.Translate(ctx, req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Translation error: %v\n", err)
			os.Exit(1)
		}

		// Output result
		fmt.Println(result)
	}
}

func createTranslator(ctx context.Context) (*engine.Translator, func(), error) {
	var actualModelPath, actualShortlist string
	var vocabPaths []string
	var err error

	// Auto-discover files if modelDir is provided
	if *modelDir != "" {
		actualModelPath, actualShortlist, vocabPaths, err = discoverModelFiles(*modelDir)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to discover model files: %w", err)
		}
	} else {
		actualModelPath = *modelPath
		actualShortlist = *shortlist
		vocabPaths = []string{*vocabSrc, *vocabTrg, *vocab}
	}

	// Open model file
	modelFile, err := os.Open(actualModelPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open model file: %w", err)
	}

	// Open shortlist file
	shortlistFile, err := os.Open(actualShortlist)
	if err != nil {
		modelFile.Close()
		return nil, nil, fmt.Errorf("failed to open shortlist file: %w", err)
	}

	// Open vocabulary files
	var vocabFiles []*os.File
	vocabularies := make([]io.Reader, len(vocabPaths))
	for i, vocabPath := range vocabPaths {
		vocabFile, err := os.Open(vocabPath)
		if err != nil {
			// Cleanup previously opened files
			modelFile.Close()
			shortlistFile.Close()
			for _, f := range vocabFiles {
				f.Close()
			}
			return nil, nil, fmt.Errorf("failed to open vocabulary file %s: %w", vocabPath, err)
		}
		vocabFiles = append(vocabFiles, vocabFile)
		vocabularies[i] = vocabFile
	}

	// Cleanup function
	cleanup := func() {
		modelFile.Close()
		shortlistFile.Close()
		for _, f := range vocabFiles {
			f.Close()
		}
	}

	// Create engine config
	cfg := engine.Config{
		CompileConfig: wasm.CompileConfig{
			Stderr: nil,
			Stdout: nil,
		},
		FilesBundle: engine.FilesBundle{
			Model:            modelFile,
			LexicalShortlist: shortlistFile,
			Vocabularies:     vocabularies,
		},
		CacheSize:       *cacheSize,
		BergamotOptions: engine.DefaultBergamotOptions(),
	}

	// Create translator
	translator, err := engine.New(ctx, cfg)
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("failed to create translator: %w", err)
	}

	// Wrap cleanup to also close translator
	fullCleanup := func() {
		if translator != nil {
			translator.Close(context.Background())
		}
		cleanup()
	}

	return translator, fullCleanup, nil
}

// discoverModelFiles automatically finds model files in a directory
func discoverModelFiles(dir string) (modelPath, shortlistPath string, vocabPaths []string, err error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", "", nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		fullPath := filepath.Join(dir, name)

		// Detect model file (.bin but not lex*.bin)
		if filepath.Ext(name) == ".bin" && !hasPrefix(name, "lex") {
			modelPath = fullPath
		}
		// Detect shortlist file (lex*.bin)
		if filepath.Ext(name) == ".bin" && hasPrefix(name, "lex") {
			shortlistPath = fullPath
		}
		// Detect vocabulary files (.spm)
		if filepath.Ext(name) == ".spm" {
			vocabPaths = append(vocabPaths, fullPath)
		}
	}

	if modelPath == "" {
		return "", "", nil, fmt.Errorf("model file (*.bin) not found in directory: %s", dir)
	}
	if shortlistPath == "" {
		return "", "", nil, fmt.Errorf("shortlist file (lex*.bin) not found in directory: %s", dir)
	}
	if len(vocabPaths) == 0 {
		return "", "", nil, fmt.Errorf("vocabulary files (*.spm) not found in directory: %s", dir)
	}

	return modelPath, shortlistPath, vocabPaths, nil
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

// runREPL runs the interactive REPL (Read-Eval-Print Loop) mode
func runREPL(ctx context.Context, translator *engine.Translator) {
	fmt.Println("Enter text to translate (type 'exit', 'quit', or press Ctrl+D to exit)")
	if *html {
		fmt.Println("HTML mode: enabled")
	}
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")

		// Read input
		if !scanner.Scan() {
			// EOF (Ctrl+D) or error
			break
		}

		input := strings.TrimSpace(scanner.Text())

		// Check for exit commands
		if input == "exit" || input == "quit" {
			fmt.Println("Goodbye!")
			break
		}

		// Skip empty input
		if input == "" {
			continue
		}

		// Translate
		req := engine.TranslationRequest{
			Text: input,
			Options: engine.TranslationOptions{
				HTML: *html,
			},
		}

		result, err := translator.Translate(ctx, req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Translation error: %v\n", err)
			continue
		}

		// Output result
		fmt.Println(result)
		fmt.Println()
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
	}
}
