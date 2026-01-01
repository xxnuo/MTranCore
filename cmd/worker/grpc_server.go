package main

import (
	"context"
	"io"
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

func NewGRPCServerWithEngine(cfg *Config, em *EngineManager) *GRPCServer {
	return &GRPCServer{
		config:        cfg,
		shutdownCh:    make(chan struct{}),
		engineManager: em,
	}
}

func (g *GRPCServer) Poweroff(ctx context.Context, req *pb.PoweroffRequest) (*pb.PoweroffResponse, error) {
	if req.Time < 0 {
		return &pb.PoweroffResponse{
			Code:    int32(CodePoweroffInvalidParams),
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

func (g *GRPCServer) Compute(ctx context.Context, req *pb.ComputeRequest) (*pb.ComputeResponse, error) {
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

	queue := g.engineManager.GetQueue()
	if queue == nil || !queue.IsReady() {
		logger.Fatal("Engine not ready. Worker should auto-load on startup")
	}

	translationReq := engine.TranslationRequest{
		Text: req.Text,
		Options: engine.TranslationOptions{
			HTML: req.Html,
		},
	}

	translatedText, err := queue.Translate(ctx, translationReq)
	if err != nil {
		logger.Fatal("Translation failed: %v", err)
	}

	return &pb.ComputeResponse{
		Code:           int32(CodeSuccess),
		Message:        "OK",
		TranslatedText: translatedText,
	}, nil
}

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

		queue := g.engineManager.GetQueue()
		if queue == nil || !queue.IsReady() {
			logger.Fatal("Engine not ready. Worker should auto-load on startup")
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
			logger.Fatal("Translation failed: %v", err)
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

func (g *GRPCServer) ShutdownChannel() <-chan struct{} {
	return g.shutdownCh
}

func (g *GRPCServer) Close() error {
	return g.engineManager.Close()
}
