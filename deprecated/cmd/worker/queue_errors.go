package main

import "errors"

var (
	// ErrQueueClosed indicates the translation queue has been closed
	ErrQueueClosed = errors.New("translation queue is closed")

	// ErrTranslatorNotReady indicates no translator is available
	ErrTranslatorNotReady = errors.New("translator is not ready")
)
