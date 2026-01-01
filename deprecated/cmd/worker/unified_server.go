package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v3"
	fiberlog "github.com/gofiber/fiber/v3/log"
	"github.com/kiuber/gofiber3-contrib/websocket"

	engine "github.com/xxnuo/MTranCore/engine"
	"github.com/xxnuo/MTranCore/internal/logger"
)

type UnifiedServer struct {
	app            *fiber.App
	grpcService    *GRPCServer
	engineManager  *EngineManager
	shutdownCh     chan struct{}
	config         *Config
	activeReqs     int32
	isShuttingDown atomic.Bool
}

func NewUnifiedServer(cfg *Config) *UnifiedServer {
	if log := logger.GetLogger(); log != nil {
		fiberLogger := log.GetWriter(logger.LogLevelInfo)
		fiberlog.SetOutput(fiberLogger)
	} else {
		fiberlog.SetOutput(io.Discard)
	}

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(NewErrorResponse(CodeInternalError, err.Error()))
		},
	})

	server := &UnifiedServer{
		app:           app,
		config:        cfg,
		shutdownCh:    make(chan struct{}),
		engineManager: NewEngineManager(cfg),
	}

	app.Server().Logger = &logger.DiscardLogger{}

	if cfg.EnableHTTP {
		app.Get("/health", server.health)
		app.Post("/trans", server.trans)
		app.Post("/exit", server.exit)
	}

	if cfg.EnableWebSocket {
		app.Use("/ws", func(c fiber.Ctx) error {
			if websocket.IsWebSocketUpgrade(c) {
				return c.Next()
			}
			return fiber.ErrUpgradeRequired
		})
		app.Get("/ws", websocket.New(server.handleWebSocket))
	}

	return server
}

func (s *UnifiedServer) GetApp() *fiber.App {
	return s.app
}

func (s *UnifiedServer) health(c fiber.Ctx) error {
	ready := s.engineManager.IsReady()
	return c.JSON(StandardResponse{
		Code:    CodeSuccess,
		Message: "OK",
		Data: map[string]interface{}{
			"ready": ready,
		},
	})
}

func (s *UnifiedServer) trans(c fiber.Ctx) error {
	if s.isShuttingDown.Load() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(NewErrorResponse(CodeInternalError, "server is shutting down"))
	}

	atomic.AddInt32(&s.activeReqs, 1)
	defer atomic.AddInt32(&s.activeReqs, -1)

	type TransReq struct {
		Text string `json:"text"`
		HTML bool   `json:"html"`
	}

	var req TransReq
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(NewErrorResponse(CodeInvalidParams, "Invalid request body"))
	}

	if req.Text == "" {
		return c.Status(fiber.StatusBadRequest).JSON(NewErrorResponse(CodeInvalidParams, "text is required"))
	}

	queue := s.engineManager.GetQueue()
	if queue == nil || !queue.IsReady() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(NewErrorResponse(CodeNotReady, "Translation engine not ready"))
	}

	transReq := engine.TranslationRequest{
		Text: req.Text,
		Options: engine.TranslationOptions{
			HTML: req.HTML,
		},
	}

	result, err := queue.Translate(context.Background(), transReq)
	if err != nil {
		if isFatalWASMError(err) {
			logger.Error("Fatal WASM error detected, exiting process: %v", err)
			go func() {
				time.Sleep(100 * time.Millisecond)
				os.Exit(1)
			}()
		}
		return c.Status(fiber.StatusInternalServerError).JSON(NewErrorResponse(CodeTransFailure, fmt.Sprintf("Translation failed: %v", err)))
	}

	return c.SendString(result)
}

func (s *UnifiedServer) exit(c fiber.Ctx) error {
	type ExitReq struct {
		Time  int  `json:"time"`
		Force bool `json:"force"`
	}

	var req ExitReq
	if err := c.Bind().JSON(&req); err != nil {
		req.Time = 0
		req.Force = false
	}

	if req.Time < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(NewErrorResponse(CodeInvalidParams, "time must be non-negative"))
	}

	s.isShuttingDown.Store(true)

	go func() {
		if req.Time > 0 {
			time.Sleep(time.Duration(req.Time) * time.Second)
		}

		if !req.Force {
			s.engineManager.WaitForIdle(&s.activeReqs, 30*time.Second)
		}

		close(s.shutdownCh)
	}()

	return c.JSON(NewSuccessResponse(map[string]string{
		"message": "Shutdown initiated",
	}))
}

func (s *UnifiedServer) handleWebSocket(c *websocket.Conn) {
	atomic.AddInt32(&s.activeReqs, 1)
	defer atomic.AddInt32(&s.activeReqs, -1)

	type WSRequest struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data,omitempty"`
	}

	type WSResponse struct {
		Type string      `json:"type"`
		Code ErrorCode   `json:"code"`
		Msg  string      `json:"msg,omitempty"`
		Data interface{} `json:"data,omitempty"`
	}

	for {
		var req WSRequest
		if err := c.ReadJSON(&req); err != nil {
			logger.Debug("[WebSocket] Read error: %v", err)
			break
		}

		logger.Debug("[WebSocket] Received message type: %s", req.Type)

		var resp WSResponse
		resp.Type = req.Type

		switch req.Type {
		case "health":
			ready := s.engineManager.IsReady()
			resp.Code = CodeSuccess
			resp.Msg = "OK"
			resp.Data = map[string]interface{}{
				"ready": ready,
			}

		case "trans":
			type TransData struct {
				Text string `json:"text"`
				HTML bool   `json:"html"`
			}
			var data TransData
			if err := json.Unmarshal(req.Data, &data); err != nil {
				resp.Code = CodeInvalidParams
				resp.Msg = "Invalid data"
				break
			}

			if data.Text == "" {
				resp.Code = CodeInvalidParams
				resp.Msg = "text is required"
				break
			}

			queue := s.engineManager.GetQueue()
			if queue == nil || !queue.IsReady() {
				resp.Code = CodeNotReady
				resp.Msg = "Translation engine not ready"
				break
			}

			transReq := engine.TranslationRequest{
				Text: data.Text,
				Options: engine.TranslationOptions{
					HTML: data.HTML,
				},
			}

			result, err := queue.Translate(context.Background(), transReq)
			if err != nil {
				if isFatalWASMError(err) {
					logger.Error("Fatal WASM error detected, exiting process: %v", err)
					go func() {
						time.Sleep(100 * time.Millisecond)
						os.Exit(1)
					}()
				}
				resp.Code = CodeTransFailure
				resp.Msg = fmt.Sprintf("Translation failed: %v", err)
				break
			}

			resp.Code = CodeSuccess
			resp.Msg = "OK"
			resp.Data = map[string]string{
				"translated_text": result,
			}

		case "exit":
			type ExitData struct {
				Time  int  `json:"time"`
				Force bool `json:"force"`
			}
			var data ExitData
			if err := json.Unmarshal(req.Data, &data); err != nil {
				data.Time = 0
				data.Force = false
			}

			if data.Time < 0 {
				resp.Code = CodeInvalidParams
				resp.Msg = "time must be non-negative"
				break
			}

			s.isShuttingDown.Store(true)

			go func() {
				if data.Time > 0 {
					time.Sleep(time.Duration(data.Time) * time.Second)
				}

				if !data.Force {
					s.engineManager.WaitForIdle(&s.activeReqs, 30*time.Second)
				}

				close(s.shutdownCh)
			}()

			resp.Code = CodeSuccess
			resp.Msg = "Shutdown initiated"

		default:
			resp.Code = CodeInvalidParams
			resp.Msg = fmt.Sprintf("Unknown message type: %s", req.Type)
		}

		if err := c.WriteJSON(resp); err != nil {
			logger.Debug("[WebSocket] Write error: %v", err)
			break
		}
	}
}

func (s *UnifiedServer) SetGRPCService(grpc *GRPCServer) {
	s.grpcService = grpc
	s.engineManager = grpc.engineManager
}

func (s *UnifiedServer) ShutdownChannel() <-chan struct{} {
	return s.shutdownCh
}

func (s *UnifiedServer) Close() error {
	if s.engineManager != nil {
		return s.engineManager.Close()
	}
	return nil
}
