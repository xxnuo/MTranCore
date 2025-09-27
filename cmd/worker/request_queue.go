package main

import (
	"context"
	"sync"

	engine "github.com/xxnuo/MTranCore/engine"
)

// TranslationQueue manages sequential translation requests
type TranslationQueue struct {
	translator *engine.Translator
	reqChan    chan *translationRequest
	stopChan   chan struct{}
	wg         sync.WaitGroup
	mu         sync.RWMutex
}

type translationRequest struct {
	ctx      context.Context
	req      engine.TranslationRequest
	respChan chan translationResponse
}

type translationResponse struct {
	result string
	err    error
}

// NewTranslationQueue creates a new translation queue
func NewTranslationQueue() *TranslationQueue {
	q := &TranslationQueue{
		reqChan:  make(chan *translationRequest, 100), // Buffer for 100 requests
		stopChan: make(chan struct{}),
	}

	// Start the worker goroutine
	q.wg.Add(1)
	go q.worker()

	return q
}

// SetTranslator updates the translator instance
func (q *TranslationQueue) SetTranslator(translator *engine.Translator) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.translator = translator
}

// Translate submits a translation request and waits for the result
func (q *TranslationQueue) Translate(ctx context.Context, req engine.TranslationRequest) (string, error) {
	respChan := make(chan translationResponse, 1)

	translationReq := &translationRequest{
		ctx:      ctx,
		req:      req,
		respChan: respChan,
	}

	select {
	case <-q.stopChan:
		return "", ErrQueueClosed
	case <-ctx.Done():
		return "", ctx.Err()
	case q.reqChan <- translationReq:
		// Request submitted successfully
	}

	// Wait for response
	select {
	case <-q.stopChan:
		return "", ErrQueueClosed
	case <-ctx.Done():
		return "", ctx.Err()
	case resp := <-respChan:
		return resp.result, resp.err
	}
}

// worker processes translation requests sequentially
func (q *TranslationQueue) worker() {
	defer q.wg.Done()

	for {
		select {
		case <-q.stopChan:
			return
		case req := <-q.reqChan:
			q.processRequest(req)
		}
	}
}

// processRequest handles a single translation request
func (q *TranslationQueue) processRequest(req *translationRequest) {
	q.mu.RLock()
	translator := q.translator
	q.mu.RUnlock()

	var resp translationResponse

	if translator == nil {
		resp.err = ErrTranslatorNotReady
	} else {
		resp.result, resp.err = translator.Translate(req.ctx, req.req)
	}

	// Send response
	select {
	case <-q.stopChan:
		return
	case req.respChan <- resp:
		// Response sent successfully
	}
}

// Close stops the queue worker
func (q *TranslationQueue) Close() {
	close(q.stopChan)
	q.wg.Wait()
}

// IsReady checks if a translator is available
func (q *TranslationQueue) IsReady() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.translator != nil
}
