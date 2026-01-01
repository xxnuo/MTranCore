package main

import (
	"context"
	engine "github.com/xxnuo/MTranCore/engine"
	"github.com/xxnuo/MTranCore/internal/logger"
	pb "github.com/xxnuo/MTranCore/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"sync/atomic"
	"time"
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
func NewGRPCServerWithEngine(cfg *Config, em *EngineManager) *GRPCServer {
	return &GRPCServer{
		config:        cfg,
		shutdownCh:    make(chan struct{}),
		engineManager: em,
	}
}
func (g *GRPCServer) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	isReady := g.engineManager.IsReady()
	return &pb.HealthResponse{
		Code:    int32(CodeSuccess),
		Message: "OK",
		Health:  isReady,
	}, nil
}
func (g *GRPCServer) Exit(ctx context.Context, req *pb.ExitRequest) (*pb.ExitResponse, error) {
	if req.Time < 0 {
		return &pb.ExitResponse{
			Code:    int32(CodeExitInvalidParams),
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
	if req.Force {
		return &pb.ExitResponse{
			Code:    int32(CodeSuccess),
			Message: "Server is shutting down",
		}, nil
	}
	return &pb.ExitResponse{
		Code:    int32(CodeExitWaitingTask),
		Message: "Server is shutting down, waiting for streams to complete",
	}, nil
}
func (g *GRPCServer) Trans(ctx context.Context, req *pb.TransRequest) (*pb.TransResponse, error) {
	if g.isShuttingDown.Load() {
		return &pb.TransResponse{
			Code:    int32(CodeTransInternalError),
			Message: "Server is shutting down",
		}, nil
	}
	if req.Text == "" {
		return &pb.TransResponse{
			Code:    int32(CodeTransInvalidParams),
			Message: "text is required",
		}, nil
	}
	queue := g.engineManager.GetQueue()
	if queue == nil || !queue.IsReady() {
		logger.Error("Engine not health")
		return &pb.TransResponse{
			Code:    int32(CodeTransInvalidParams),
			Message: "Engine not health",
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
		logger.Error("Translation failed: %v", err)
		return &pb.TransResponse{
			Code:    int32(CodeTransFailure),
			Message: "Translation failed: " + err.Error(),
		}, nil
	}
	return &pb.TransResponse{
		Code:           int32(CodeSuccess),
		Message:        "OK",
		TranslatedText: translatedText,
	}, nil
}
func (g *GRPCServer) TransStream(stream pb.TranslatorService_TransStreamServer) error {
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
		if g.isShuttingDown.Load() {
			return stream.Send(&pb.TransResponse{
				Code:    int32(CodeTransInternalError),
				Message: "Server is shutting down",
			})
		}
		if req.Text == "" {
			if err := stream.Send(&pb.TransResponse{
				Code:           int32(CodeSuccess),
				Message:        "OK",
				TranslatedText: "",
			}); err != nil {
				return status.Errorf(codes.Internal, "failed to send response: %v", err)
			}
			continue
		}
		queue := g.engineManager.GetQueue()
		if queue == nil || !queue.IsReady() {
			logger.Error("Engine not health")
			return stream.Send(&pb.TransResponse{
				Code:    int32(CodeTransInvalidParams),
				Message: "Engine not health",
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
			logger.Error("Translation failed: %v", err)
			return stream.Send(&pb.TransResponse{
				Code:    int32(CodeTransFailure),
				Message: "Translation failed: " + err.Error(),
			})
		}
		if err := stream.Send(&pb.TransResponse{
			Code:           int32(CodeSuccess),
			Message:        "OK",
			TranslatedText: translatedText,
		}); err != nil {
			return status.Errorf(codes.Internal, "failed to send response: %v", err)
		}
	}
}
func (g *GRPCServer) ShutdownChannel() <-chan struct{} {
	return g.shutdownCh
}
func (g *GRPCServer) Close() error {
	return g.engineManager.Close()
}
