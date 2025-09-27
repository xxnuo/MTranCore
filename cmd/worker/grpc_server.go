package main

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	engine "github.com/xxnuo/MTranCore/engine"
	pb "github.com/xxnuo/MTranCore/proto"
)

// GRPCServer implements the TranslatorService gRPC interface
type GRPCServer struct {
	pb.UnimplementedTranslatorServiceServer
	translator     *engine.Translator
	loadedFiles    *LoadedFiles
	mu             sync.RWMutex
	config         *Config
	activeStreams  int32 // atomic counter for active streams
	isShuttingDown atomic.Bool
	shutdownCh     chan struct{}
}

// NewGRPCServer creates a new gRPC server instance
func NewGRPCServer(cfg *Config) *GRPCServer {
	return &GRPCServer{
		config:     cfg,
		shutdownCh: make(chan struct{}),
	}
}

// Health checks server health
func (g *GRPCServer) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{
		Code:    int32(CodeSuccess),
		Message: "OK",
	}, nil
}

// Poweron loads the translation engine with model files
func (g *GRPCServer) Poweron(ctx context.Context, req *pb.PoweronRequest) (*pb.PoweronResponse, error) {
	if req.Path == "" {
		return &pb.PoweronResponse{
			Code:    int32(CodePoweronInvalidParams),
			Message: "path is required",
		}, nil
	}

	// Resolve path: if absolute, use as-is; otherwise join with work directory
	var fullPath string
	if filepath.IsAbs(req.Path) {
		fullPath = req.Path
	} else {
		fullPath = filepath.Join(g.config.WorkDir, req.Path)
	}

	// Check if path exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return &pb.PoweronResponse{
			Code:    int32(CodePoweronPathNotExists),
			Message: "path does not exist: " + fullPath,
		}, nil
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	// Unload existing engine if any
	if g.translator != nil {
		if err := g.translator.Close(context.Background()); err != nil {
			Warn("Failed to close existing translator: %v", err)
		}
		g.translator = nil
	}
	if g.loadedFiles != nil {
		g.loadedFiles.Close()
		g.loadedFiles = nil
	}

	// Create translator using model directory
	config := EngineConfig{ModelDir: fullPath}
	translator, loadedFiles, err := CreateTranslator(ctx, config)
	if err != nil {
		// Determine error code based on error message
		errMsg := err.Error()
		if containsAny(errMsg, "not found", "missing") {
			return &pb.PoweronResponse{
				Code:    int32(CodePoweronIncompleteFiles),
				Message: err.Error(),
			}, nil
		}
		return &pb.PoweronResponse{
			Code:    int32(CodePoweronInternalError),
			Message: err.Error(),
		}, nil
	}

	g.translator = translator
	g.loadedFiles = loadedFiles

	return &pb.PoweronResponse{
		Code:    int32(CodeSuccess),
		Message: "Engine loaded successfully",
	}, nil
}

// Poweroff shuts down the server
func (g *GRPCServer) Poweroff(ctx context.Context, req *pb.PoweroffRequest) (*pb.PoweroffResponse, error) {
	// Validate parameters
	if req.Time < 0 {
		return &pb.PoweroffResponse{
			Code:    int32(CodePoweroffInvalidParams),
			Message: "time must be non-negative",
		}, nil
	}

	g.isShuttingDown.Store(true)

	// Handle shutdown in goroutine
	go func() {
		// Wait for specified time
		if req.Time > 0 {
			time.Sleep(time.Duration(req.Time) * time.Second)
		}

		// If not force shutdown, wait for active streams
		if !req.Force {
			// Wait for active streams to complete (with timeout)
			timeout := time.After(30 * time.Second)
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()

			for {
				select {
				case <-timeout:
					Warn("Shutdown timeout reached, forcing shutdown")
					goto shutdown
				case <-ticker.C:
					if atomic.LoadInt32(&g.activeStreams) == 0 {
						goto shutdown
					}
				}
			}
		}

	shutdown:
		close(g.shutdownCh)
	}()

	if req.Force {
		return &pb.PoweroffResponse{
			Code:    int32(CodeSuccess),
			Message: "Server is shutting down",
		}, nil
	} else {
		return &pb.PoweroffResponse{
			Code:    int32(CodePoweroffWaitingTask),
			Message: "Server is shutting down, waiting for streams to complete",
		}, nil
	}
}

// Ready gets the current engine status
func (g *GRPCServer) Ready(ctx context.Context, req *pb.ReadyRequest) (*pb.ReadyResponse, error) {
	g.mu.RLock()
	isReady := g.translator != nil
	g.mu.RUnlock()

	return &pb.ReadyResponse{
		Code:    int32(CodeSuccess),
		Message: "OK",
		Ready:   isReady,
	}, nil
}

// Compute translates a single text
func (g *GRPCServer) Compute(ctx context.Context, req *pb.ComputeRequest) (*pb.ComputeResponse, error) {
	// Check if shutting down
	if g.isShuttingDown.Load() {
		return &pb.ComputeResponse{
			Code:    int32(CodeComputeInternalError),
			Message: "Server is shutting down",
		}, nil
	}

	if req.Text == "" {
		return &pb.ComputeResponse{
			Code:    int32(CodeComputeInvalidParams),
			Message: "text is required",
		}, nil
	}

	g.mu.RLock()
	translator := g.translator
	g.mu.RUnlock()

	if translator == nil {
		return &pb.ComputeResponse{
			Code:    int32(CodeComputeInvalidParams),
			Message: "Engine is not ready. Please call poweron first",
		}, nil
	}

	translationReq := engine.TranslationRequest{
		Text: req.Text,
		Options: engine.TranslationOptions{
			HTML: req.Html,
		},
	}

	translatedText, err := translator.Translate(ctx, translationReq)
	if err != nil {
		return &pb.ComputeResponse{
			Code:    int32(CodeComputeFailure),
			Message: "Translation failed: " + err.Error(),
		}, nil
	}

	return &pb.ComputeResponse{
		Code:           int32(CodeSuccess),
		Message:        "OK",
		TranslatedText: translatedText,
	}, nil
}

// ComputeStream translates multiple texts in streaming fashion
func (g *GRPCServer) ComputeStream(stream pb.TranslatorService_ComputeStreamServer) error {
	atomic.AddInt32(&g.activeStreams, 1)
	defer atomic.AddInt32(&g.activeStreams, -1)

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return status.Errorf(codes.Internal, "failed to receive request: %v", err)
		}

		// Check if shutting down
		if g.isShuttingDown.Load() {
			return stream.Send(&pb.ComputeResponse{
				Code:    int32(CodeComputeInternalError),
				Message: "Server is shutting down",
			})
		}

		if req.Text == "" {
			if err := stream.Send(&pb.ComputeResponse{
				Code:           int32(CodeSuccess),
				Message:        "OK",
				TranslatedText: "",
			}); err != nil {
				return status.Errorf(codes.Internal, "failed to send response: %v", err)
			}
			continue
		}

		g.mu.RLock()
		translator := g.translator
		g.mu.RUnlock()

		if translator == nil {
			return stream.Send(&pb.ComputeResponse{
				Code:    int32(CodeComputeInvalidParams),
				Message: "Engine is not ready. Please call poweron first",
			})
		}

		translationReq := engine.TranslationRequest{
			Text: req.Text,
			Options: engine.TranslationOptions{
				HTML: req.Html,
			},
		}

		ctx := stream.Context()
		translatedText, err := translator.Translate(ctx, translationReq)
		if err != nil {
			return stream.Send(&pb.ComputeResponse{
				Code:    int32(CodeComputeFailure),
				Message: "Translation failed: " + err.Error(),
			})
		}

		if err := stream.Send(&pb.ComputeResponse{
			Code:           int32(CodeSuccess),
			Message:        "OK",
			TranslatedText: translatedText,
		}); err != nil {
			return status.Errorf(codes.Internal, "failed to send response: %v", err)
		}
	}
}

// ShutdownChannel returns the shutdown channel
func (g *GRPCServer) ShutdownChannel() <-chan struct{} {
	return g.shutdownCh
}

// Close closes the gRPC server and releases resources
func (g *GRPCServer) Close() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.translator != nil {
		if err := g.translator.Close(context.Background()); err != nil {
			return err
		}
		g.translator = nil
	}

	if g.loadedFiles != nil {
		g.loadedFiles.Close()
		g.loadedFiles = nil
	}

	return nil
}
