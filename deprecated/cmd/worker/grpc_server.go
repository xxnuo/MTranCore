package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	engine "github.com/xxnuo/MTranCore/engine"
	"github.com/xxnuo/MTranCore/internal/logger"
	pb "github.com/xxnuo/MTranCore/proto"
)

type GRPCServer struct {
	pb.UnimplementedTranslatorServiceServer
	engineManager  *EngineManager
	config         *Config
	activeStreams  int32
	isShuttingDown atomic.Bool
	shutdownCh     chan struct{}
}

func NewGRPCServer(cfg *Config) *GRPCServer {
	return &GRPCServer{
		config:        cfg,
		shutdownCh:    make(chan struct{}),
		engineManager: NewEngineManager(cfg),
	}
}

func (g *GRPCServer) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	ready := g.engineManager.IsReady()
	return &pb.HealthResponse{
		Code:    int32(CodeSuccess),
		Message: "OK",
		Ready:   ready,
	}, nil
}

func (g *GRPCServer) Trans(ctx context.Context, req *pb.TransRequest) (*pb.TransResponse, error) {
	if g.isShuttingDown.Load() {
		return nil, status.Error(codes.Unavailable, "server is shutting down")
	}

	atomic.AddInt32(&g.activeStreams, 1)
	defer atomic.AddInt32(&g.activeStreams, -1)

	queue := g.engineManager.GetQueue()
	if queue == nil {
		return &pb.TransResponse{
			Code:    int32(CodeNotReady),
			Message: "Translation engine not ready",
		}, nil
	}

	transReq := engine.TranslationRequest{
		Text: req.Text,
		Options: engine.TranslationOptions{
			HTML: req.Html,
		},
	}

	result, err := queue.Translate(ctx, transReq)
	if err != nil {
		if isFatalWASMError(err) {
			logger.Error("Fatal WASM error detected, exiting process: %v", err)
			go func() {
				time.Sleep(100 * time.Millisecond)
				os.Exit(1)
			}()
		}
		return &pb.TransResponse{
			Code:    int32(CodeTransFailure),
			Message: fmt.Sprintf("Translation failed: %v", err),
		}, nil
	}

	return &pb.TransResponse{
		Code:           int32(CodeSuccess),
		Message:        "OK",
		TranslatedText: result,
	}, nil
}

func (g *GRPCServer) TransStream(stream pb.TranslatorService_TransStreamServer) error {
	if g.isShuttingDown.Load() {
		return status.Error(codes.Unavailable, "server is shutting down")
	}

	atomic.AddInt32(&g.activeStreams, 1)
	defer atomic.AddInt32(&g.activeStreams, -1)

	queue := g.engineManager.GetQueue()
	if queue == nil {
		return status.Error(codes.FailedPrecondition, "Translation engine not ready")
	}

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return status.Errorf(codes.Internal, "Failed to receive request: %v", err)
		}

		transReq := engine.TranslationRequest{
			Text: req.Text,
			Options: engine.TranslationOptions{
				HTML: req.Html,
			},
		}

		result, err := queue.Translate(stream.Context(), transReq)
		if err != nil {
			if isFatalWASMError(err) {
				logger.Error("Fatal WASM error detected, exiting process: %v", err)
				go func() {
					time.Sleep(100 * time.Millisecond)
					os.Exit(1)
				}()
			}
			if err := stream.Send(&pb.TransResponse{
				Code:    int32(CodeTransFailure),
				Message: fmt.Sprintf("Translation failed: %v", err),
			}); err != nil {
				return status.Errorf(codes.Internal, "Failed to send response: %v", err)
			}
			continue
		}

		if err := stream.Send(&pb.TransResponse{
			Code:           int32(CodeSuccess),
			Message:        "OK",
			TranslatedText: result,
		}); err != nil {
			return status.Errorf(codes.Internal, "Failed to send response: %v", err)
		}
	}
}

func (g *GRPCServer) Exit(ctx context.Context, req *pb.ExitRequest) (*pb.ExitResponse, error) {
	if req.Time < 0 {
		return &pb.ExitResponse{
			Code:    int32(CodeInvalidParams),
			Message: "time must be non-negative",
		}, nil
	}

	g.isShuttingDown.Store(true)

	go func() {
		if req.Time > 0 {
			time.Sleep(time.Duration(req.Time) * time.Second)
		}

		if !req.Force {
			g.engineManager.WaitForIdle(&g.activeStreams, 30*time.Second)
		}

		close(g.shutdownCh)
	}()

	return &pb.ExitResponse{
		Code:    int32(CodeSuccess),
		Message: "Shutdown initiated",
	}, nil
}

func (g *GRPCServer) ShutdownChannel() <-chan struct{} {
	return g.shutdownCh
}

func (g *GRPCServer) Close() error {
	if g.engineManager != nil {
		return g.engineManager.Close()
	}
	return nil
}

func isFatalWASMError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return containsAny(errMsg, "module closed", "exit_code")
}
