package main

import (
	"context"
	"encoding/json"
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
	engineManager  *EngineManager
	shutdownCh     chan struct{}
	config         *Config
	activeReqs     int32
	isShuttingDown atomic.Bool
}

func NewUnifiedServer(cfg *Config) *UnifiedServer {
	return NewUnifiedServerWithEngine(cfg, NewEngineManager(cfg))
}

func NewUnifiedServerWithEngine(cfg *Config, em *EngineManager) *UnifiedServer {
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
			return c.Status(code).JSON(NewErrorResponse(CodePoweronUnknownError, err.Error()))
		},
	})

	server := &UnifiedServer{
		app:           app,
		config:        cfg,
		shutdownCh:    make(chan struct{}),
		engineManager: em,
	}

	app.Server().Logger = &logger.DiscardLogger{}

	if cfg.EnableHTTP {
		app.Get("/ready", server.ready)
		app.Post("/poweroff", server.poweroff)
		app.Post("/compute", server.compute)
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

func (s *UnifiedServer) ready(c fiber.Ctx) error {
	isReady := s.engineManager.IsReady()
	return c.JSON(NewSuccessResponse(ReadyResponse{Ready: isReady}))
}

func (s *UnifiedServer) poweroff(c fiber.Ctx) error {
	var req PoweroffRequest
	if err := c.Bind().JSON(&req); err != nil {
		req.Time = 0
		req.Force = false
	}

	if req.Time < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(
			NewErrorResponse(CodePoweroffInvalidParams, "time must be non-negative"))
	}

	s.isShuttingDown.Store(true)

	go func() {
		if req.Time > 0 {
			time.Sleep(time.Duration(req.Time) * time.Second)
		}

		if !req.Force {
			s.engineManager.WaitForIdle(&s.activeReqs, 30*time.Second)
		}

		if err := s.app.Shutdown(); err != nil {
			logger.Error("Error during shutdown: %v", err)
		}
		close(s.shutdownCh)
	}()

	if req.Force {
		return c.JSON(NewSuccessResponse(fiber.Map{"message": "Server is shutting down"}))
	}
	return c.JSON(NewErrorResponse(CodePoweroffWaitingTask, "Server is shutting down, waiting for requests to complete"))
}

func (s *UnifiedServer) compute(c fiber.Ctx) error {
	atomic.AddInt32(&s.activeReqs, 1)
	defer atomic.AddInt32(&s.activeReqs, -1)

	if s.isShuttingDown.Load() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(
			NewErrorResponse(CodeComputeInternalError, "Server is shutting down"))
	}

	var req ComputeRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			NewErrorResponse(CodeComputeInvalidParams, "Invalid JSON: "+err.Error()))
	}

	if req.Text == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			NewErrorResponse(CodeComputeInvalidParams, "text is required"))
	}

	queue := s.engineManager.GetQueue()
	if queue == nil || !queue.IsReady() {
		logger.Fatal("Engine not ready. Worker should auto-load on startup")
	}

	translationReq := engine.TranslationRequest{
		Text: req.Text,
		Options: engine.TranslationOptions{
			HTML: req.HTML,
		},
	}

	ctx := context.Background()
	translatedText, err := queue.Translate(ctx, translationReq)
	if err != nil {
		logger.Fatal("Translation failed: %v", err)
	}

	c.Set("Content-Type", "text/plain; charset=utf-8")
	return c.SendString(translatedText)
}

func (s *UnifiedServer) handleWebSocket(c *websocket.Conn) {
	defer c.Close()

	for {
		var msg WSMessage
		err := c.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("[WebSocket] Error reading message: %v", err)
			}
			break
		}

		var response WSResponse
		switch msg.Type {
		case "poweroff":
			response = s.handleWSPoweroff(msg.Data)
		case "compute":
			response = s.handleWSCompute(msg.Data)
		default:
			response = WSResponse{
				Type: msg.Type,
				Code: int(CodePoweronUnknownError),
				Msg:  "Unknown message type: " + msg.Type,
			}
		}

		if err := c.WriteJSON(response); err != nil {
			logger.Error("[WebSocket] Error sending response: %v", err)
			break
		}

		if msg.Type == "poweroff" {
			break
		}
	}
}

func (s *UnifiedServer) handleWSPoweroff(data json.RawMessage) WSResponse {
	var req PoweroffRequest
	if err := json.Unmarshal(data, &req); err != nil {
		req.Time = 0
		req.Force = false
	}

	if req.Time < 0 {
		return WSResponse{
			Type: "poweroff",
			Code: int(CodePoweroffInvalidParams),
			Msg:  "time must be non-negative",
		}
	}

	s.isShuttingDown.Store(true)

	go func() {
		if req.Time > 0 {
			time.Sleep(time.Duration(req.Time) * time.Second)
		}

		if !req.Force {
			s.engineManager.WaitForIdle(&s.activeReqs, 30*time.Second)
		}

		if err := s.app.Shutdown(); err != nil {
			logger.Error("Error during shutdown: %v", err)
		}
		close(s.shutdownCh)
	}()

	if req.Force {
		return WSResponse{
			Type: "poweroff",
			Code: int(CodeSuccess),
			Msg:  "success",
			Data: fiber.Map{"message": "Server is shutting down"},
		}
	}
	return WSResponse{
		Type: "poweroff",
		Code: int(CodePoweroffWaitingTask),
		Msg:  "Server is shutting down, waiting for requests to complete",
	}
}

func (s *UnifiedServer) handleWSCompute(data json.RawMessage) WSResponse {
	atomic.AddInt32(&s.activeReqs, 1)
	defer atomic.AddInt32(&s.activeReqs, -1)

	if s.isShuttingDown.Load() {
		return WSResponse{
			Type: "compute",
			Code: int(CodeComputeInternalError),
			Msg:  "Server is shutting down",
		}
	}

	var req ComputeRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return WSResponse{
			Type: "compute",
			Code: int(CodeComputeInvalidParams),
			Msg:  "Invalid JSON: " + err.Error(),
		}
	}

	if req.Text == "" {
		return WSResponse{
			Type: "compute",
			Code: int(CodeComputeInvalidParams),
			Msg:  "text is required",
		}
	}

	queue := s.engineManager.GetQueue()
	if queue == nil || !queue.IsReady() {
		logger.Fatal("Engine not ready. Worker should auto-load on startup")
	}

	translationReq := engine.TranslationRequest{
		Text: req.Text,
		Options: engine.TranslationOptions{
			HTML: req.HTML,
		},
	}

	ctx := context.Background()
	translatedText, err := queue.Translate(ctx, translationReq)
	if err != nil {
		logger.Fatal("Translation failed: %v", err)
	}

	return WSResponse{
		Type: "compute",
		Code: int(CodeSuccess),
		Msg:  "success",
		Data: ComputeResponse{TranslatedText: translatedText},
	}
}

func (s *UnifiedServer) ShutdownChannel() <-chan struct{} {
	return s.shutdownCh
}

func (s *UnifiedServer) Close() error {
	return s.engineManager.Close()
}

func (s *UnifiedServer) Listen(addr string) error {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.app.Listen(addr, fiber.ListenConfig{
			DisableStartupMessage: true,
		})
	}()

	time.Sleep(100 * time.Millisecond)

	w.Close()
	os.Stdout = oldStdout

	io.Copy(io.Discard, r)
	r.Close()

	return <-errCh
}
