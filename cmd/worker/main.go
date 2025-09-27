package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/xxnuo/MTranCore/proto"
)

func main() {
	// Load configuration
	cfg := LoadConfig()

	// Initialize unified logger
	InitLogger(cfg.LogLevel)

	Info("==============================================")
	Info("Starting MTranCore Worker Service")
	Info("Log Level: %s", cfg.LogLevel)
	Info("Work Directory: %s", cfg.WorkDir)
	Info("==============================================")

	var wg sync.WaitGroup
	shutdownCh := make(chan struct{})

	// Start HTTP server if enabled
	var httpServer *Server
	if cfg.EnableHTTP {
		httpServer = NewServer(cfg)
		wg.Add(1)
		go func() {
			defer wg.Done()
			addr := fmt.Sprintf("%s:%s", cfg.HTTPHost, cfg.HTTPPort)
			Info("[HTTP] Starting server on %s", addr)
			Debug("[HTTP] Available endpoints:")
			Debug("  GET  /health   - Health check")
			Debug("  POST /poweron  - Load translation engine")
			Debug("  POST /poweroff - Shutdown server")
			Debug("  GET  /ready    - Check engine status")
			Debug("  POST /compute  - Translate text")

			if err := httpServer.Listen(addr); err != nil {
				Error("[HTTP] Server error: %v", err)
			}
		}()
	} else {
		Info("[HTTP] Disabled")
	}

	// Start WebSocket server if enabled
	var wsServer *WebSocketServer
	if cfg.EnableWebSocket {
		wsServer = NewWebSocketServer(cfg)
		wg.Add(1)
		go func() {
			defer wg.Done()
			addr := fmt.Sprintf("%s:%s", cfg.WebSocketHost, cfg.WebSocketPort)
			Info("[WebSocket] Starting server on %s", addr)
			Debug("[WebSocket] Available message types:")
			Debug("  - poweron  - Load translation engine")
			Debug("  - poweroff - Shutdown server")
			Debug("  - ready    - Check engine status")
			Debug("  - compute  - Translate text")

			if err := wsServer.Listen(addr); err != nil {
				Error("[WebSocket] Server error: %v", err)
			}
		}()
	} else {
		Info("[WebSocket] Disabled")
	}

	// Start gRPC server if enabled
	var grpcServerInstance *grpc.Server
	var grpcService *GRPCServer
	if cfg.EnableGRPC {
		grpcService = NewGRPCServer(cfg)
		grpcServerInstance = grpc.NewServer()
		pb.RegisterTranslatorServiceServer(grpcServerInstance, grpcService)
		reflection.Register(grpcServerInstance)

		wg.Add(1)
		go func() {
			defer wg.Done()
			addr := fmt.Sprintf("%s:%s", cfg.GRPCHost, cfg.GRPCPort)
			lis, err := net.Listen("tcp", addr)
			if err != nil {
				Fatal("[gRPC] Failed to listen: %v", err)
			}

			Info("[gRPC] Starting server on %s", addr)
			Debug("[gRPC] Service: TranslatorService")
			Debug("  - Health")
			Debug("  - Poweron")
			Debug("  - Poweroff")
			Debug("  - Ready")
			Debug("  - Compute")
			Debug("  - ComputeStream")

			if err := grpcServerInstance.Serve(lis); err != nil {
				Error("[gRPC] Server error: %v", err)
			}
		}()
	} else {
		Info("[gRPC] Disabled")
	}

	Info("==============================================")
	Info("All enabled services are running")
	Info("Press Ctrl+C to shutdown gracefully")
	Info("==============================================")

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sigCh:
		Info("\nReceived shutdown signal, shutting down gracefully...")
	case <-shutdownCh:
		Info("Shutdown initiated by service...")
	}

	// Shutdown all services
	shutdownCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if httpServer != nil {
		Info("[HTTP] Shutting down...")
		httpServer.app.ShutdownWithContext(shutdownCtx)
		httpServer.Close()
	}

	if wsServer != nil {
		Info("[WebSocket] Shutting down...")
		wsServer.app.ShutdownWithContext(shutdownCtx)
		wsServer.Close()
	}

	if grpcServerInstance != nil {
		Info("[gRPC] Shutting down...")
		grpcServerInstance.GracefulStop()
		if grpcService != nil {
			grpcService.Close()
		}
	}

	wg.Wait()
	Info("All services stopped. Goodbye!")
}
