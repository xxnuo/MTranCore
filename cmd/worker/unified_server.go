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
		app.Get("/health", server.health)
		app.Post("/exit", server.exit)
		app.Post("/trans", server.trans)
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
	isReady := s.engineManager.IsReady()
	return c.JSON(NewSuccessResponse(HealthResponse{Health: isReady}))
}

func (s *UnifiedServer) exit(c fiber.Ctx) error {
        var req ExitRequest
        if err := c.Bind().JSON(&req); err != nil {
                req.Time = 0
                req.Force = false
        }

        if req.Time < 0 {
                return c.Status(fiber.StatusBadRequest).JSON(
                        NewErrorResponse(CodeExitInvalidParams, "time must be non-negative"))
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
        return c.JSON(NewErrorResponse(CodeExitWaitingTask, "Server is shutting down, waiting for requests to complete"))
}
func (s *UnifiedServer) trans(c fiber.Ctx) error {
        atomic.AddInt32(&s.activeReqs, 1)
        defer atomic.AddInt32(&s.activeReqs, -1)

        if s.isShuttingDown.Load() {
                return c.Status(fiber.StatusServiceUnavailable).JSON(
                        NewErrorResponse(CodeTransInternalError, "Server is shutting down"))
        }

        var req TransRequest
        if err := c.Bind().JSON(&req); err != nil {
                return c.Status(fiber.StatusBadRequest).JSON(
                        NewErrorResponse(CodeTransInvalidParams, "Invalid JSON: "+err.Error()))
        }

                if req.Text == "" {

                        return c.Status(fiber.StatusBadRequest).JSON(

                                NewErrorResponse(CodeTransInvalidParams, "text is required"))

                }

        

                queue := s.engineManager.GetQueue()

                if queue == nil || !queue.IsReady() {

                        logger.Error("Engine not health")

                        return c.Status(fiber.StatusServiceUnavailable).JSON(

                                NewErrorResponse(CodeTransInvalidParams, "Engine not health"))

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
	                logger.Error("Translation failed: %v", err)
	                return c.Status(fiber.StatusInternalServerError).JSON(
	                        NewErrorResponse(CodeTransFailure, "Translation failed: "+err.Error()))
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
		                case "exit":
		                        response = s.handleWSExit(msg.Data)
		                case "trans":
		                        response = s.handleWSTrans(msg.Data)
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

		if msg.Type == "exit" {
			break
		}
	}
}

func (s *UnifiedServer) handleWSExit(data json.RawMessage) WSResponse {
        var req ExitRequest
        if err := json.Unmarshal(data, &req); err != nil {
                req.Time = 0
                req.Force = false
        }

        if req.Time < 0 {
                return WSResponse{
                        Type: "exit",
                        Code: int(CodeExitInvalidParams),
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
                        Type: "exit",
                        Code: int(CodeSuccess),
                        Msg:  "success",
                        Data: fiber.Map{"message": "Server is shutting down"},
                }
        }
        return WSResponse{
                Type: "exit",
                Code: int(CodeExitWaitingTask),
                Msg:  "Server is shutting down, waiting for requests to complete",
        }
}
func (s *UnifiedServer) handleWSTrans(data json.RawMessage) WSResponse {
        atomic.AddInt32(&s.activeReqs, 1)
        defer atomic.AddInt32(&s.activeReqs, -1)

        if s.isShuttingDown.Load() {
                return WSResponse{
                        Type: "trans",
                        Code: int(CodeTransInternalError),
                        Msg:  "Server is shutting down",
                }
        }

        var req TransRequest
        if err := json.Unmarshal(data, &req); err != nil {
                return WSResponse{
                        Type: "trans",
                        Code: int(CodeTransInvalidParams),
                        Msg:  "Invalid JSON: " + err.Error(),
                }
        }

                if req.Text == "" {

                        return WSResponse{

                                Type: "trans",

                                Code: int(CodeTransInvalidParams),

                                Msg:  "text is required",

                        }

                }

        

                queue := s.engineManager.GetQueue()

                if queue == nil || !queue.IsReady() {

                        logger.Error("Engine not health")

                        return WSResponse{

                                Type: "trans",

                                Code: int(CodeTransInvalidParams),

                                Msg:  "Engine not health",

                        }

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
	                logger.Error("Translation failed: %v", err)
	                return WSResponse{
	                        Type: "trans",
	                        Code: int(CodeTransFailure),
	                        Msg:  "Translation failed: " + err.Error(),
	                }
	        }
	
	        return WSResponse{		Type: "trans",
		Code: int(CodeSuccess),
		Msg:  "success",
		Data: TransResponse{TranslatedText: translatedText},
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
