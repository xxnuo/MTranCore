package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// getProjectRoot finds the project root directory
func getProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	// Check if we're already in the project root
	if _, err := os.Stat(filepath.Join(cwd, "models")); err == nil {
		return cwd, nil
	}
	// Search upward for the models directory
	for dir := cwd; dir != "/"; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "models")); err == nil {
			return dir, nil
		}
	}
	return "", errors.New("cannot find project root directory containing models folder")
}

// getTestModelPaths returns paths to test model files
func getTestModelPaths() (modelPath, shortlistPath string, vocabPaths []string, err error) {
	projectRoot, err := getProjectRoot()
	if err != nil {
		return "", "", nil, err
	}
	modelPath = filepath.Join(projectRoot, "models", "enzh", "model.enzh.intgemm.alphas.bin")
	shortlistPath = filepath.Join(projectRoot, "models", "enzh", "lex.50.50.enzh.s2t.bin")
	vocabPaths = []string{
		filepath.Join(projectRoot, "models", "enzh", "srcvocab.enzh.spm"),
		filepath.Join(projectRoot, "models", "enzh", "trgvocab.enzh.spm"),
	}
	// Verify files exist
	if _, err := os.Stat(modelPath); err != nil {
		return "", "", nil, fmt.Errorf("model file not found: %w", err)
	}
	if _, err := os.Stat(shortlistPath); err != nil {
		return "", "", nil, fmt.Errorf("shortlist file not found: %w", err)
	}
	for _, path := range vocabPaths {
		if _, err := os.Stat(path); err != nil {
			return "", "", nil, fmt.Errorf("vocab file not found: %w", err)
		}
	}
	return modelPath, shortlistPath, vocabPaths, nil
}
