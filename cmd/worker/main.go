package main

import (
	"github.com/xxnuo/MTranCore/internal/logger"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	pb "github.com/xxnuo/MTranCore/proto"
)

// isClosedConnectionError checks if an error is related to a closed network connection
func isClosedConnectionError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	return strings.Contains(errMsg, "use of closed network connection") ||
		strings.Contains(errMsg, "mux: server closed") ||
		errors.Is(err, net.ErrClosed) ||
		strings.Contains(errMsg, "Server closed")
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "[PANIC] main: %v\n", r)
		}
	}()

	// Load configuration
	cfg := GetConfig()

	// Initialize unified logger
	logger.InitLogger(cfg.LogLevel)

	logger.Debug("[DEBUG-MAIN] Configuration loaded")
	logger.Info("Starting MTranCore Worker Service (Unified)")
	logger.Info("Log Level: %s", cfg.LogLevel)
	if cfg.ModelDir != "" {
		logger.Info("Model Directory: %s", cfg.ModelDir)
	}
	
	logger.Info("Server Address: %s:%s", cfg.ServerHost, cfg.ServerPort)
	// logger.Info("==============================================")

	// Create a TCP listener on the unified port
	addr := fmt.Sprintf("%s:%s", cfg.ServerHost, cfg.ServerPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatal("Failed to listen on %s: %v", addr, err)
	}

	// Create a cmux multiplexer
	m := cmux.New(lis)

	var wg sync.WaitGroup

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
		
		// Configure gRPC server with performance optimizations
		grpcServerInstance = grpc.NewServer(
			// Keepalive settings - keep connections alive to reduce handshake overhead
			grpc.KeepaliveParams(keepalive.ServerParameters{
				MaxConnectionIdle:     15 * time.Minute, // Close idle connections after 15 minutes
				MaxConnectionAge:      30 * time.Minute, // Force reconnect after 30 minutes
				MaxConnectionAgeGrace: 5 * time.Second,  // Allow 5s for pending RPCs to complete
				Time:                  5 * time.Second,  // Send keepalive ping every 5s if no activity
				Timeout:               1 * time.Second,  // Wait 1s for ping ack before closing
			}),
			grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
				MinTime:             5 * time.Second, // Allow pings every 5s minimum
				PermitWithoutStream: true,            // Allow pings even when no streams are active
			}),
			// Connection options
			grpc.MaxConcurrentStreams(1000),        // Allow up to 1000 concurrent streams per connection
			grpc.MaxRecvMsgSize(4 * 1024 * 1024),   // 4MB max receive message size
			grpc.MaxSendMsgSize(4 * 1024 * 1024),   // 4MB max send message size
			// Buffer sizes for better throughput
			grpc.ReadBufferSize(32 * 1024),  // 32KB read buffer
			grpc.WriteBufferSize(32 * 1024), // 32KB write buffer
			// Initial window size for flow control
			grpc.InitialWindowSize(1 << 20),     // 1MB initial window size (per stream)
			grpc.InitialConnWindowSize(1 << 20), // 1MB initial connection window size
		)
		
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
			logger.Info("[gRPC] Starting server on %s", addr)
			logger.Debug("[gRPC] Service: TranslatorService")
			logger.Debug("  - Health")
			logger.Debug("  - Poweron")
			logger.Debug("  - Poweroff")
			logger.Debug("  - Reboot")
			logger.Debug("  - Ready")
			logger.Debug("  - Compute")
			logger.Debug("  - TransStream")

			if err := grpcServerInstance.Serve(grpcListener); err != nil {
				if isClosedConnectionError(err) {
					logger.Debug("[gRPC] Server stopped: %v", err)
				} else {
					logger.Error("[gRPC] Server error: %v", err)
				}
			}
		}()

		// Start gRPC Unix socket server if configured
		if cfg.GRPCUnixSocket != "" {
			// Remove existing socket file if it exists
			os.Remove(cfg.GRPCUnixSocket)

			unixListener, err := net.Listen("unix", cfg.GRPCUnixSocket)
			if err != nil {
				logger.Fatal("Failed to listen on Unix socket %s: %v", cfg.GRPCUnixSocket, err)
			}

			// Set socket permissions to allow connections
			if err := os.Chmod(cfg.GRPCUnixSocket, 0666); err != nil {
				logger.Warn("Failed to set Unix socket permissions: %v", err)
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				defer unixListener.Close()
				defer os.Remove(cfg.GRPCUnixSocket)

				logger.Info("[gRPC Unix] Starting server on %s", cfg.GRPCUnixSocket)
				logger.Debug("[gRPC Unix] Using same service instance as TCP gRPC")

				if err := grpcServerInstance.Serve(unixListener); err != nil {
					if isClosedConnectionError(err) {
						logger.Debug("[gRPC Unix] Server stopped: %v", err)
					} else {
						logger.Error("[gRPC Unix] Server error: %v", err)
					}
				}
			}()

			logger.Info("[gRPC Unix] Unix socket enabled: %s", cfg.GRPCUnixSocket)
		}
	}

	// Start HTTP/WebSocket server
	if unifiedServer != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if cfg.EnableHTTP {
				logger.Info("[HTTP] Starting server on %s", addr)
				logger.Debug("[HTTP] Available endpoints:")
				logger.Debug("  GET  /health   - Health check")
				logger.Debug("  POST /health - Health check engine")
				logger.Debug("  POST /trans  - Translate text server")
				logger.Debug("  POST /exit   - Shutdown server translation engine")
				logger.Debug("  GET  / status")
				logger.Debug("  POST / text")
			}
			if cfg.EnableWebSocket {
				logger.Info("[WebSocket] Starting server on %s", addr)
				logger.Debug("[WebSocket] Available at /ws")
				logger.Debug("[WebSocket] Message types:")
				logger.Debug("  - health - Health check engine")
				logger.Debug("  - trans  - Translate text server")
				logger.Debug("  - exit   - Shutdown server translation engine")
				logger.Debug("  -  status")
				logger.Debug("  -  text")
			}

			// Use the HTTP listener from cmux
			// Temporarily redirect stdout to suppress Fiber's startup banner
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Start listening
			errCh := make(chan error, 1)
			go func() {
				errCh <- unifiedServer.app.Listener(httpListener, fiber.ListenConfig{
					DisableStartupMessage: true,
				})
			}()

			// Wait a bit for startup messages to be written
			time.Sleep(100 * time.Millisecond)

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout

			// Discard the captured output
			io.Copy(io.Discard, r)
			r.Close()

			// Get any error
			if err := <-errCh; err != nil {
				if isClosedConnectionError(err) {
					logger.Debug("[HTTP/WebSocket] Server stopped: %v", err)
				} else {
					logger.Error("[HTTP/WebSocket] Server error: %v", err)
				}
			}
		}()
	}

	// Start cmux
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := m.Serve(); err != nil {
			if isClosedConnectionError(err) {
				logger.Debug("[Multiplexer] Stopped: %v", err)
			} else {
				logger.Error("[Multiplexer] Error: %v", err)
			}
		}
	}()

	// logger.Info("==============================================")
	if len(enabledServices) > 0 {
		logger.Info("Enabled services: %s", strings.Join(enabledServices, ", "))
		logger.Info("All services running on port %s", cfg.ServerPort)
	} else {
		logger.Warn("No services enabled!")
	}
	logger.Info("Press Ctrl+C to shutdown gracefully")
	// logger.Info("==============================================")

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Collect shutdown channels from all services
	shutdownChannels := make([]<-chan struct{}, 0)
	if unifiedServer != nil {
		shutdownChannels = append(shutdownChannels, unifiedServer.ShutdownChannel())
	}
	if grpcService != nil {
		shutdownChannels = append(shutdownChannels, grpcService.ShutdownChannel())
	}

	// Wait for either OS signal or service-initiated shutdown
	select {
	case <-sigCh:
		logger.Info("\nReceived shutdown signal, shutting down gracefully...")
	case <-mergeShutdownChannels(shutdownChannels):
		logger.Info("Shutdown initiated by service...")
	}

	// Shutdown all services
	shutdownCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if unifiedServer != nil {
		logger.Info("[HTTP/WebSocket] Shutting down...")
		unifiedServer.app.ShutdownWithContext(shutdownCtx)
		unifiedServer.Close()
	}

	if grpcServerInstance != nil {
		logger.Info("[gRPC] Shutting down...")
		grpcServerInstance.GracefulStop()
		if grpcService != nil {
			grpcService.Close()
		}
	}

	// Close the multiplexer listener
	lis.Close()

	wg.Wait()
	logger.Info("All services stopped. Goodbye!")
}

// mergeShutdownChannels merges multiple shutdown channels into one
func mergeShutdownChannels(channels []<-chan struct{}) <-chan struct{} {
	merged := make(chan struct{})

	if len(channels) == 0 {
		// Return a channel that will never close if no channels provided
		return merged
	}

	var wg sync.WaitGroup
	for _, ch := range channels {
		if ch != nil {
			wg.Add(1)
			go func(c <-chan struct{}) {
				defer wg.Done()
				<-c
			}(ch)
		}
	}

	go func() {
		wg.Wait()
		close(merged)
	}()

	return merged
}
