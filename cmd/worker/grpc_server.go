package main

import (
	"context"
	"fmt"
	"io"
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
	engineManager  *EngineManager
	config         *Config
	activeStreams  int32 // atomic counter for active streams
	isShuttingDown atomic.Bool
	shutdownCh     chan struct{}
}

// NewGRPCServer creates a new gRPC server instance
func NewGRPCServer(cfg *Config) *GRPCServer {
	return &GRPCServer{
		config:        cfg,
		shutdownCh:    make(chan struct{}),
		engineManager: NewEngineManager(cfg),
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
	result := g.engineManager.Poweron(ctx, req.Path)

	if !result.Success {
		return &pb.PoweronResponse{
			Code:    int32(result.ErrorCode),
			Message: result.ErrorMessage,
		}, nil
	}

	message := "Engine loaded successfully"
	if result.AlreadyLoaded {
		message = "Engine already loaded"
	}

	return &pb.PoweronResponse{
		Code:    int32(CodeSuccess),
		Message: message,
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
			g.engineManager.WaitForIdle(&g.activeStreams, 30*time.Second)
		}

		close(g.shutdownCh)
	}()

	if req.Force {
		return &pb.PoweroffResponse{
			Code:    int32(CodeSuccess),
			Message: "Server is shutting down",
		}, nil
	}
	return &pb.PoweroffResponse{
		Code:    int32(CodePoweroffWaitingTask),
		Message: "Server is shutting down, waiting for streams to complete",
	}, nil
}

// Reboot reloads the translation engine
func (g *GRPCServer) Reboot(ctx context.Context, req *pb.RebootRequest) (*pb.RebootResponse, error) {
	// Validate parameters
	if req.Time < 0 {
		return &pb.RebootResponse{
			Code:    int32(CodeRebootInvalidParams),
			Message: "time must be non-negative",
		}, nil
	}

	// Handle reboot in goroutine if time is specified
	if req.Time > 0 {
		g.engineManager.RebootAsync(int(req.Time), req.Force, &g.activeStreams, nil)
		return &pb.RebootResponse{
			Code:    int32(CodeSuccess),
			Message: "Engine will reboot in " + fmt.Sprintf("%d", req.Time) + " seconds",
		}, nil
	}

	// Immediate reboot
	result := g.engineManager.Reboot(ctx, req.Force, &g.activeStreams)

	if !result.Success {
		return &pb.RebootResponse{
			Code:    int32(result.ErrorCode),
			Message: result.ErrorMessage,
		}, nil
	}

	message := "Engine rebooted successfully"
	if req.Force {
		message = "Engine rebooted (forced)"
	}

	return &pb.RebootResponse{
		Code:    int32(CodeSuccess),
		Message: message,
	}, nil
}

// Ready gets the current engine status
func (g *GRPCServer) Ready(ctx context.Context, req *pb.ReadyRequest) (*pb.ReadyResponse, error) {
	isReady := g.engineManager.IsReady()

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

	// Use the translation queue to avoid concurrent WASM execution
	queue := g.engineManager.GetQueue()
	if queue == nil || !queue.IsReady() {
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

	translatedText, err := queue.Translate(ctx, translationReq)
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

		// Use the translation queue to avoid concurrent WASM execution
		queue := g.engineManager.GetQueue()
		if queue == nil || !queue.IsReady() {
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
		translatedText, err := queue.Translate(ctx, translationReq)
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
	return g.engineManager.Close()
}
