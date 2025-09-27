package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/gofiber/fiber/v3"
	"github.com/soheilhy/cmux"
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
	Info("Starting MTranCore Worker Service (Unified)")
	Info("Log Level: %s", cfg.LogLevel)
	Info("Work Directory: %s", cfg.WorkDir)
	Info("Server Address: %s:%s", cfg.ServerHost, cfg.ServerPort)
	Info("==============================================")

	// Create a TCP listener on the unified port
	addr := fmt.Sprintf("%s:%s", cfg.ServerHost, cfg.ServerPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		Fatal("Failed to listen on %s: %v", addr, err)
	}

	// Create a cmux multiplexer
	m := cmux.New(lis)

	var wg sync.WaitGroup
	shutdownCh := make(chan struct{})

	// Match gRPC connections
	grpcListener := m.MatchWithWriters(
		cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"),
	)

	// Match HTTP connections (including WebSocket upgrades)
	httpListener := m.Match(cmux.Any())

	// Create unified server for HTTP and WebSocket
	var unifiedServer *UnifiedServer
	var grpcServerInstance *grpc.Server
	var grpcService *GRPCServer

	enabledServices := []string{}

	if cfg.EnableHTTP || cfg.EnableWebSocket {
		unifiedServer = NewUnifiedServer(cfg)
		
		if cfg.EnableHTTP {
			enabledServices = append(enabledServices, "HTTP")
		}
		if cfg.EnableWebSocket {
			enabledServices = append(enabledServices, "WebSocket")
		}
	}

	// Start gRPC server if enabled
	if cfg.EnableGRPC {
		grpcService = NewGRPCServer(cfg)
		grpcServerInstance = grpc.NewServer()
		pb.RegisterTranslatorServiceServer(grpcServerInstance, grpcService)
		reflection.Register(grpcServerInstance)
		
		// Link gRPC service with unified server for shared state
		if unifiedServer != nil {
			unifiedServer.SetGRPCService(grpcService)
		}

		enabledServices = append(enabledServices, "gRPC")

		wg.Add(1)
		go func() {
			defer wg.Done()
			Info("[gRPC] Starting server on %s", addr)
			Debug("[gRPC] Service: TranslatorService")
			Debug("  - Health")
			Debug("  - Poweron")
			Debug("  - Poweroff")
			Debug("  - Ready")
			Debug("  - Compute")
			Debug("  - ComputeStream")

			if err := grpcServerInstance.Serve(grpcListener); err != nil {
				Error("[gRPC] Server error: %v", err)
			}
		}()
	}

	// Start HTTP/WebSocket server
	if unifiedServer != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if cfg.EnableHTTP {
				Info("[HTTP] Starting server on %s", addr)
				Debug("[HTTP] Available endpoints:")
				Debug("  GET  /health   - Health check")
				Debug("  POST /poweron  - Load translation engine")
				Debug("  POST /poweroff - Shutdown server")
				Debug("  GET  /ready    - Check engine status")
				Debug("  POST /compute  - Translate text")
			}
			if cfg.EnableWebSocket {
				Info("[WebSocket] Starting server on %s", addr)
				Debug("[WebSocket] Available at /ws")
				Debug("[WebSocket] Message types:")
				Debug("  - poweron  - Load translation engine")
				Debug("  - poweroff - Shutdown server")
				Debug("  - ready    - Check engine status")
				Debug("  - compute  - Translate text")
			}

			// Use the HTTP listener from cmux
			if err := unifiedServer.app.Listener(httpListener, fiber.ListenConfig{}); err != nil {
				Error("[HTTP/WebSocket] Server error: %v", err)
			}
		}()
	}

	// Start cmux
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := m.Serve(); err != nil {
			Error("[Multiplexer] Error: %v", err)
		}
	}()

	Info("==============================================")
	if len(enabledServices) > 0 {
		Info("Enabled services: %s", strings.Join(enabledServices, ", "))
		Info("All services running on port %s", cfg.ServerPort)
	} else {
		Warn("No services enabled!")
	}
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

	if unifiedServer != nil {
		Info("[HTTP/WebSocket] Shutting down...")
		unifiedServer.app.ShutdownWithContext(shutdownCtx)
		unifiedServer.Close()
	}

	if grpcServerInstance != nil {
		Info("[gRPC] Shutting down...")
		grpcServerInstance.GracefulStop()
		if grpcService != nil {
			grpcService.Close()
		}
	}

	// Close the multiplexer listener
	lis.Close()

	wg.Wait()
	Info("All services stopped. Goodbye!")
}
