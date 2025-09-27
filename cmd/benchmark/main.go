package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fasthttp/websocket"
	pb "github.com/xxnuo/MTranCore/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ComputeRequest struct {
	Text string `json:"text"`
	HTML bool   `json:"html"`
}

type PoweronRequest struct {
	Path string `json:"path"`
}

type BenchmarkResult struct {
	Name          string
	TotalRequests int64
	Duration      time.Duration
	Success       int64
	Failed        int64
	AvgLatency    time.Duration
	MinLatency    time.Duration
	MaxLatency    time.Duration
	Throughput    float64
}

var (
	serverURL   = flag.String("url", "http://localhost:8080", "Server URL")
	modelPath   = flag.String("model", "", "Model directory path (if empty, uses ./models/enzh)")
	iterations  = flag.Int("n", 100, "Number of iterations")
	concurrency = flag.Int("c", 1, "Number of concurrent workers")
	testType    = flag.String("test", "all", "Test type: all, compute, html, long, parallel")
	protocol    = flag.String("protocol", "http", "Protocol to use: http, grpc, ws")
)

// Client interface for different protocols
type Client interface {
	Poweron(ctx context.Context, modelPath string) error
	Compute(ctx context.Context, text string, html bool) (string, error)
	Close() error
}

// HTTPClient implements Client for HTTP protocol
type HTTPClient struct {
	baseURL string
	client  *http.Client
}

func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *HTTPClient) Poweron(ctx context.Context, modelPath string) error {
	reqBody := PoweronRequest{Path: modelPath}
	body, _ := json.Marshal(reqBody)

	resp, err := c.client.Post(c.baseURL+"/poweron", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}
	return nil
}

func (c *HTTPClient) Compute(ctx context.Context, text string, html bool) (string, error) {
	reqBody := ComputeRequest{Text: text, HTML: html}
	body, _ := json.Marshal(reqBody)

	resp, err := c.client.Post(c.baseURL+"/compute", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	// Server now returns plain text on success
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (c *HTTPClient) Close() error {
	return nil
}

// GRPCClient implements Client for gRPC protocol
type GRPCClient struct {
	conn   *grpc.ClientConn
	client pb.TranslatorServiceClient
}

func NewGRPCClient(address string) (*GRPCClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &GRPCClient{
		conn:   conn,
		client: pb.NewTranslatorServiceClient(conn),
	}, nil
}

func (c *GRPCClient) Poweron(ctx context.Context, modelPath string) error {
	resp, err := c.client.Poweron(ctx, &pb.PoweronRequest{Path: modelPath})
	if err != nil {
		return err
	}
	if resp.Code != 200 {
		return fmt.Errorf("poweron failed: %s", resp.Message)
	}
	return nil
}

func (c *GRPCClient) Compute(ctx context.Context, text string, html bool) (string, error) {
	resp, err := c.client.Compute(ctx, &pb.ComputeRequest{Text: text, Html: html})
	if err != nil {
		return "", err
	}
	if resp.Code != 200 {
		return "", fmt.Errorf("compute failed: %s", resp.Message)
	}
	return resp.TranslatedText, nil
}

func (c *GRPCClient) Close() error {
	return c.conn.Close()
}

// WSClient implements Client for WebSocket protocol
type WSClient struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func NewWSClient(wsURL string) (*WSClient, error) {
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return nil, err
	}
	return &WSClient{conn: conn}, nil
}

func (c *WSClient) Poweron(ctx context.Context, modelPath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	msg := map[string]interface{}{
		"type": "poweron",
		"data": map[string]string{"path": modelPath},
	}

	if err := c.conn.WriteJSON(msg); err != nil {
		return err
	}

	var resp struct {
		Type string `json:"type"`
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := c.conn.ReadJSON(&resp); err != nil {
		return err
	}
	if resp.Code != 200 {
		return fmt.Errorf("poweron failed: %s", resp.Msg)
	}
	return nil
}

func (c *WSClient) Compute(ctx context.Context, text string, html bool) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	msg := map[string]interface{}{
		"type": "compute",
		"data": map[string]interface{}{"text": text, "html": html},
	}

	if err := c.conn.WriteJSON(msg); err != nil {
		return "", err
	}

	var resp struct {
		Type string `json:"type"`
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			TranslatedText string `json:"translated_text"`
		} `json:"data"`
	}
	if err := c.conn.ReadJSON(&resp); err != nil {
		return "", err
	}
	if resp.Code != 200 {
		return "", fmt.Errorf("compute failed: %s", resp.Msg)
	}
	return resp.Data.TranslatedText, nil
}

func (c *WSClient) Close() error {
	return c.conn.Close()
}

var client Client

func main() {
	flag.Parse()

	// Determine model path
	modelDir := *modelPath
	if modelDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current directory: %v\n", err)
			os.Exit(1)
		}
		modelDir = filepath.Join(cwd, "models", "enzh")
	}

	// Verify model exists
	if _, err := os.Stat(modelDir); os.IsNotExist(err) {
		fmt.Printf("Model directory not found: %s\n", modelDir)
		os.Exit(1)
	}

	fmt.Printf("=== Benchmark Configuration ===\n")
	fmt.Printf("Server URL: %s\n", *serverURL)
	fmt.Printf("Protocol: %s\n", *protocol)
	fmt.Printf("Model Path: %s\n", modelDir)
	fmt.Printf("Iterations: %d\n", *iterations)
	fmt.Printf("Concurrency: %d\n", *concurrency)
	fmt.Printf("Test Type: %s\n", *testType)
	fmt.Printf("==============================\n\n")

	// Initialize client based on protocol
	var err error
	switch *protocol {
	case "http":
		client = NewHTTPClient(*serverURL)
	case "grpc":
		// Extract host:port from URL
		u, err := url.Parse(*serverURL)
		if err != nil {
			fmt.Printf("Invalid server URL: %v\n", err)
			os.Exit(1)
		}
		address := u.Host
		if address == "" {
			address = *serverURL // Assume it's already host:port format
		}
		client, err = NewGRPCClient(address)
		if err != nil {
			fmt.Printf("Failed to create gRPC client: %v\n", err)
			os.Exit(1)
		}
	case "ws":
		// Convert http(s) URL to ws(s)
		wsURL := *serverURL
		if u, err := url.Parse(wsURL); err == nil {
			if u.Scheme == "http" {
				u.Scheme = "ws"
			} else if u.Scheme == "https" {
				u.Scheme = "wss"
			}
			u.Path = "/ws"
			wsURL = u.String()
		}
		client, err = NewWSClient(wsURL)
		if err != nil {
			fmt.Printf("Failed to create WebSocket client: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown protocol: %s\n", *protocol)
		os.Exit(1)
	}
	defer client.Close()

	// Load engine
	fmt.Printf("Loading engine...\n")
	ctx := context.Background()
	if err := client.Poweron(ctx, modelDir); err != nil {
		fmt.Printf("Failed to load engine: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Engine loaded successfully!\n\n")

	// Run benchmarks based on test type
	switch *testType {
	case "compute":
		result := benchmarkCompute("Simple Text Translation", "Hello, world!", false)
		printResult(result)
	case "html":
		result := benchmarkCompute("HTML Translation", "<p>Hello, world!</p><div>This is a test.</div>", true)
		printResult(result)
	case "long":
		longText := "The quick brown fox jumps over the lazy dog. " +
			"This sentence contains every letter of the alphabet. " +
			"Machine translation has made significant progress in recent years. " +
			"Neural networks and deep learning have revolutionized the field. " +
			"Modern translation systems can handle complex linguistic structures."
		result := benchmarkCompute("Long Text Translation", longText, false)
		printResult(result)
	case "parallel":
		result := benchmarkParallel()
		printResult(result)
	case "all":
		results := []BenchmarkResult{
			benchmarkCompute("Simple Text Translation", "Hello, world!", false),
			benchmarkCompute("HTML Translation", "<p>Hello, world!</p><div>This is a test.</div>", true),
			benchmarkCompute("Long Text Translation", getLongText(), false),
		}
		if *concurrency > 1 {
			results = append(results, benchmarkParallel())
		}
		fmt.Printf("\n=== Benchmark Summary ===\n\n")
		for _, result := range results {
			printResult(result)
			fmt.Println()
		}
	default:
		fmt.Printf("Unknown test type: %s\n", *testType)
		os.Exit(1)
	}
}

func benchmarkCompute(name, text string, html bool) BenchmarkResult {
	result := BenchmarkResult{
		Name:       name,
		MinLatency: time.Hour, // Start with a large value
	}

	fmt.Printf("Running benchmark: %s\n", name)
	start := time.Now()
	ctx := context.Background()

	for i := 0; i < *iterations; i++ {
		reqStart := time.Now()
		_, err := client.Compute(ctx, text, html)
		latency := time.Since(reqStart)

		if err != nil {
			atomic.AddInt64(&result.Failed, 1)
		} else {
			atomic.AddInt64(&result.Success, 1)
		}

		// Update latency stats
		if latency < result.MinLatency {
			result.MinLatency = latency
		}
		if latency > result.MaxLatency {
			result.MaxLatency = latency
		}

		// Progress indicator
		if (i+1)%10 == 0 || i+1 == *iterations {
			fmt.Printf("\rProgress: %d/%d", i+1, *iterations)
		}
	}

	result.Duration = time.Since(start)
	result.TotalRequests = int64(*iterations)
	result.AvgLatency = result.Duration / time.Duration(*iterations)
	result.Throughput = float64(result.TotalRequests) / result.Duration.Seconds()

	fmt.Printf("\n")
	return result
}

func benchmarkParallel() BenchmarkResult {
	result := BenchmarkResult{
		Name:       "Parallel Translation",
		MinLatency: time.Hour,
	}

	var wg sync.WaitGroup
	var latencyMutex sync.Mutex
	totalPerWorker := *iterations / *concurrency

	fmt.Printf("Running parallel benchmark with %d workers\n", *concurrency)
	start := time.Now()
	ctx := context.Background()

	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < totalPerWorker; j++ {
				reqStart := time.Now()
				_, err := client.Compute(ctx, "Hello, world!", false)
				latency := time.Since(reqStart)

				if err != nil {
					atomic.AddInt64(&result.Failed, 1)
				} else {
					atomic.AddInt64(&result.Success, 1)
				}

				// Update latency stats (with mutex)
				latencyMutex.Lock()
				if latency < result.MinLatency {
					result.MinLatency = latency
				}
				if latency > result.MaxLatency {
					result.MaxLatency = latency
				}
				latencyMutex.Unlock()
			}
		}(i)
	}

	wg.Wait()
	result.Duration = time.Since(start)
	result.TotalRequests = int64(totalPerWorker * *concurrency)
	result.AvgLatency = result.Duration / time.Duration(result.TotalRequests)
	result.Throughput = float64(result.TotalRequests) / result.Duration.Seconds()

	return result
}

func printResult(result BenchmarkResult) {
	fmt.Printf("--- %s ---\n", result.Name)
	fmt.Printf("Total Requests:  %d\n", result.TotalRequests)
	fmt.Printf("Successful:      %d\n", result.Success)
	fmt.Printf("Failed:          %d\n", result.Failed)
	fmt.Printf("Duration:        %v\n", result.Duration)
	fmt.Printf("Avg Latency:     %v\n", result.AvgLatency)
	fmt.Printf("Min Latency:     %v\n", result.MinLatency)
	fmt.Printf("Max Latency:     %v\n", result.MaxLatency)
	fmt.Printf("Throughput:      %.2f req/s\n", result.Throughput)
}

func getLongText() string {
	return "The quick brown fox jumps over the lazy dog. " +
		"This sentence contains every letter of the alphabet. " +
		"Machine translation has made significant progress in recent years. " +
		"Neural networks and deep learning have revolutionized the field. " +
		"Modern translation systems can handle complex linguistic structures."
}
