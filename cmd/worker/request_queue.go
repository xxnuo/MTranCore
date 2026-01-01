package main

import (
	"context"
	engine "github.com/xxnuo/MTranCore/engine"
	"github.com/xxnuo/MTranCore/internal/logger"
	"sync"
)

// TranslationQueue manages sequential translation requests
type TranslationQueue struct {
	translator *engine.Translator
	reqChan    chan *translationRequest
	stopChan   chan struct{}
	wg         sync.WaitGroup
	mu         sync.RWMutex
	closeOnce  sync.Once
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
	logger.Debug("[DEBUG-QUEUE] SetTranslator: setting translator=%v", translator)
	q.mu.Lock()
	defer q.mu.Unlock()
	q.translator = translator
	logger.Debug("[DEBUG-QUEUE] SetTranslator: q.translator now=%v", q.translator)
}

// Translate submits a translation request and waits for the result
func (q *TranslationQueue) Translate(ctx context.Context, req engine.TranslationRequest) (string, error) {
	logger.Debug("[DEBUG-QUEUE] Translate: starting, text length=%d", len(req.Text))
	respChan := make(chan translationResponse, 1)
	translationReq := &translationRequest{
		ctx:      ctx,
		req:      req,
		respChan: respChan,
	}
	select {
	case <-q.stopChan:
		logger.Debug("[DEBUG-QUEUE] Translate: queue closed")
		return "", ErrQueueClosed
	case <-ctx.Done():
		logger.Debug("[DEBUG-QUEUE] Translate: context done")
		return "", ctx.Err()
	case q.reqChan <- translationReq:
		logger.Debug("[DEBUG-QUEUE] Translate: request submitted to queue")
	}
	// Wait for response
	select {
	case <-q.stopChan:
		logger.Debug("[DEBUG-QUEUE] Translate: queue closed while waiting")
		return "", ErrQueueClosed
	case <-ctx.Done():
		logger.Debug("[DEBUG-QUEUE] Translate: context done while waiting")
		return "", ctx.Err()
	case resp := <-respChan:
		logger.Debug("[DEBUG-QUEUE] Translate: got response, err=%v", resp.err)
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
	logger.Debug("[DEBUG-QUEUE] processRequest: starting")
	q.mu.RLock()
	translator := q.translator
	q.mu.RUnlock()
	logger.Debug("[DEBUG-QUEUE] processRequest: translator=%v", translator)
	var resp translationResponse
	if translator == nil {
		logger.Debug("[DEBUG-QUEUE] processRequest: translator is nil")
		resp.err = ErrTranslatorNotReady
	} else {
		logger.Debug("[DEBUG-QUEUE] processRequest: calling translator.Translate")
		resp.result, resp.err = translator.Translate(req.ctx, req.req)
		logger.Debug("[DEBUG-QUEUE] processRequest: translator.Translate returned, err=%v", resp.err)
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
	q.closeOnce.Do(func() {
		close(q.stopChan)
		q.wg.Wait()
	})
}

// IsReady checks if a translator is available
func (q *TranslationQueue) IsReady() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.translator != nil
}
