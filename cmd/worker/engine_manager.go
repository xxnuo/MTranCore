package main

import (
	"context"

	engine "github.com/xxnuo/MTranCore/engine"
)

// Type aliases for backward compatibility
type EngineConfig = engine.EngineConfig
type LoadedFiles = engine.LoadedFiles

// CreateTranslator creates a new translator from the engine configuration
// This is a wrapper around engine.CreateTranslator for backward compatibility
func CreateTranslator(ctx context.Context, cfg EngineConfig) (*engine.Translator, *LoadedFiles, error) {
	return engine.CreateTranslator(ctx, cfg)
}
