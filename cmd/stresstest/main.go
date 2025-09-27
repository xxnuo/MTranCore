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

type StressTestResult struct {
	Name           string
	TotalRequests  int64
	SuccessCount   int64
	FailureCount   int64
	Duration       time.Duration
	Throughput     float64
	FailureRate    float64
	ConcurrentLoad int
	Errors         []string
}

var (
	serverURL   = flag.String("url", "http://localhost:8988", "Server URL")
	modelPath   = flag.String("model", "", "Model directory path (if empty, uses ./models/enzh)")
	testType    = flag.String("test", "all", "Test type: all, concurrency, sustained, memory, reload, mixed")
	concurrency = flag.Int("c", 50, "Number of concurrent workers for high-concurrency test")
	duration    = flag.Duration("d", 30*time.Second, "Duration for sustained load test")
	iterations  = flag.Int("n", 1000, "Number of iterations for memory stability test")
	reloads     = flag.Int("r", 5, "Number of reloads for rapid reload test")
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

	fmt.Printf("=== Stress Test Configuration ===\n")
	fmt.Printf("Server URL: %s\n", *serverURL)
	fmt.Printf("Protocol: %s\n", *protocol)
	fmt.Printf("Model Path: %s\n", modelDir)
	fmt.Printf("Test Type: %s\n", *testType)
	fmt.Printf("=================================\n\n")

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

	// Run stress tests based on test type
	switch *testType {
	case "concurrency":
		runHighConcurrencyTest(modelDir)
	case "sustained":
		runSustainedLoadTest(modelDir)
	case "memory":
		runMemoryStabilityTest(modelDir)
	case "reload":
		runRapidReloadTest(modelDir)
	case "mixed":
		runMixedWorkloadTest(modelDir)
	case "all":
		fmt.Printf("=== Running All Stress Tests ===\n\n")
		runHighConcurrencyTest(modelDir)
		fmt.Println()
		runSustainedLoadTest(modelDir)
		fmt.Println()
		runMemoryStabilityTest(modelDir)
		fmt.Println()
		runRapidReloadTest(modelDir)
		fmt.Println()
		runMixedWorkloadTest(modelDir)
	default:
		fmt.Printf("Unknown test type: %s\n", *testType)
		os.Exit(1)
	}
}

func runHighConcurrencyTest(modelDir string) {
	fmt.Printf("=== High Concurrency Test ===\n")
	fmt.Printf("Loading engine...\n")
	ctx := context.Background()
	if err := client.Poweron(ctx, modelDir); err != nil {
		fmt.Printf("Failed to load engine: %v\n", err)
		return
	}
	fmt.Printf("Engine loaded successfully!\n\n")

	requestsPerWorker := 10
	totalRequests := *concurrency * requestsPerWorker

	var wg sync.WaitGroup
	var successCount, failureCount int32
	errorChan := make(chan error, totalRequests)

	fmt.Printf("Starting high concurrency test: %d workers, %d requests each\n", *concurrency, requestsPerWorker)
	start := time.Now()

	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < requestsPerWorker; j++ {
				text := fmt.Sprintf("Hello from worker %d, request %d!", workerID, j)
				_, err := client.Compute(context.Background(), text, false)
				if err != nil {
					atomic.AddInt32(&failureCount, 1)
					errorChan <- fmt.Errorf("worker %d request %d failed: %w", workerID, j, err)
				} else {
					atomic.AddInt32(&successCount, 1)
				}
			}
		}(i)
	}

	wg.Wait()
	close(errorChan)
	testDuration := time.Since(start)

	// Collect errors
	var errors []string
	for err := range errorChan {
		if len(errors) < 10 {
			errors = append(errors, err.Error())
		}
	}

	result := StressTestResult{
		Name:           "High Concurrency",
		TotalRequests:  int64(totalRequests),
		SuccessCount:   int64(successCount),
		FailureCount:   int64(failureCount),
		Duration:       testDuration,
		Throughput:     float64(totalRequests) / testDuration.Seconds(),
		FailureRate:    float64(failureCount) / float64(totalRequests),
		ConcurrentLoad: *concurrency,
		Errors:         errors,
	}

	printResult(result)
}

func runSustainedLoadTest(modelDir string) {
	fmt.Printf("=== Sustained Load Test ===\n")
	fmt.Printf("Loading engine...\n")
	ctx := context.Background()
	if err := client.Poweron(ctx, modelDir); err != nil {
		fmt.Printf("Failed to load engine: %v\n", err)
		return
	}
	fmt.Printf("Engine loaded successfully!\n\n")

	workers := 10
	deadline := time.Now().Add(*duration)

	var wg sync.WaitGroup
	var requestCount, successCount, failureCount int32
	stopChan := make(chan struct{})

	fmt.Printf("Starting sustained load test: %d workers for %v\n", workers, *duration)
	start := time.Now()

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			reqNum := 0

			for {
				select {
				case <-stopChan:
					return
				default:
					if time.Now().After(deadline) {
						return
					}

					text := fmt.Sprintf("Sustained load worker %d request %d", workerID, reqNum)
					_, err := client.Compute(context.Background(), text, false)
					atomic.AddInt32(&requestCount, 1)

					if err != nil {
						atomic.AddInt32(&failureCount, 1)
					} else {
						atomic.AddInt32(&successCount, 1)
					}

					reqNum++
				}
			}
		}(i)
	}

	// Progress reporter
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				elapsed := time.Since(start)
				currentRequests := atomic.LoadInt32(&requestCount)
				fmt.Printf("Progress: %.0fs / %.0fs - Requests: %d (%.2f req/s)\n",
					elapsed.Seconds(), duration.Seconds(), currentRequests, float64(currentRequests)/elapsed.Seconds())
			case <-stopChan:
				return
			}
		}
	}()

	wg.Wait()
	close(stopChan)
	testDuration := time.Since(start)

	result := StressTestResult{
		Name:           "Sustained Load",
		TotalRequests:  int64(requestCount),
		SuccessCount:   int64(successCount),
		FailureCount:   int64(failureCount),
		Duration:       testDuration,
		Throughput:     float64(requestCount) / testDuration.Seconds(),
		FailureRate:    float64(failureCount) / float64(requestCount),
		ConcurrentLoad: workers,
	}

	printResult(result)
}

func runMemoryStabilityTest(modelDir string) {
	fmt.Printf("=== Memory Stability Test ===\n")
	fmt.Printf("Loading engine...\n")
	ctx := context.Background()
	if err := client.Poweron(ctx, modelDir); err != nil {
		fmt.Printf("Failed to load engine: %v\n", err)
		return
	}
	fmt.Printf("Engine loaded successfully!\n\n")

	var successCount int32

	fmt.Printf("Starting memory stability test with %d iterations\n", *iterations)
	start := time.Now()

	for i := 0; i < *iterations; i++ {
		text := "This is a memory stability test iteration. We are checking for memory leaks."
		_, err := client.Compute(ctx, text, false)
		if err == nil {
			atomic.AddInt32(&successCount, 1)
		}

		// Progress indicator
		if (i+1)%100 == 0 || i+1 == *iterations {
			fmt.Printf("Progress: %d/%d (%.1f%%)\n", i+1, *iterations, float64(i+1)/float64(*iterations)*100)
		}
	}

	testDuration := time.Since(start)
	failureCount := int32(*iterations) - successCount

	result := StressTestResult{
		Name:          "Memory Stability",
		TotalRequests: int64(*iterations),
		SuccessCount:  int64(successCount),
		FailureCount:  int64(failureCount),
		Duration:      testDuration,
		Throughput:    float64(*iterations) / testDuration.Seconds(),
		FailureRate:   float64(failureCount) / float64(*iterations),
	}

	printResult(result)
}

func runRapidReloadTest(modelDir string) {
	fmt.Printf("=== Rapid Engine Reload Test ===\n")
	fmt.Printf("Testing %d rapid engine reloads\n\n", *reloads)

	var successCount, failureCount int32
	start := time.Now()

	ctx := context.Background()
	for i := 0; i < *reloads; i++ {
		fmt.Printf("Reload %d/%d\n", i+1, *reloads)

		// Load engine
		if err := client.Poweron(ctx, modelDir); err != nil {
			fmt.Printf("  Failed: %v\n", err)
			atomic.AddInt32(&failureCount, 1)
			continue
		}

		// Verify engine is working with a translation
		text := fmt.Sprintf("Test translation after reload %d", i+1)
		_, err := client.Compute(ctx, text, false)
		if err != nil {
			fmt.Printf("  Translation failed: %v\n", err)
			atomic.AddInt32(&failureCount, 1)
			continue
		}

		atomic.AddInt32(&successCount, 1)
		fmt.Printf("  Success\n")

		// Small delay between reloads
		time.Sleep(100 * time.Millisecond)
	}

	testDuration := time.Since(start)

	result := StressTestResult{
		Name:          "Rapid Reload",
		TotalRequests: int64(*reloads),
		SuccessCount:  int64(successCount),
		FailureCount:  int64(failureCount),
		Duration:      testDuration,
		FailureRate:   float64(failureCount) / float64(*reloads),
	}

	printResult(result)
}

func runMixedWorkloadTest(modelDir string) {
	fmt.Printf("=== Mixed Workload Test ===\n")
	fmt.Printf("Loading engine...\n")
	ctx := context.Background()
	if err := client.Poweron(ctx, modelDir); err != nil {
		fmt.Printf("Failed to load engine: %v\n", err)
		return
	}
	fmt.Printf("Engine loaded successfully!\n\n")

	testDuration := 20 * time.Second
	deadline := time.Now().Add(testDuration)

	var wg sync.WaitGroup
	var translateCount, healthCount, readyCount int32
	stopChan := make(chan struct{})

	fmt.Printf("Starting mixed workload test for %v\n", testDuration)
	start := time.Now()

	// Translation workers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				select {
				case <-stopChan:
					return
				default:
					if time.Now().After(deadline) {
						return
					}

					_, err := client.Compute(context.Background(), "Mixed workload translation test", false)
					if err == nil {
						atomic.AddInt32(&translateCount, 1)
					}
					time.Sleep(100 * time.Millisecond)
				}
			}
		}(i)
	}

	// Health check workers
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stopChan:
					return
				default:
					if time.Now().After(deadline) {
						return
					}

					resp, err := http.Get(*serverURL + "/health")
					if err == nil {
						resp.Body.Close()
						atomic.AddInt32(&healthCount, 1)
					}
					time.Sleep(50 * time.Millisecond)
				}
			}
		}()
	}

	// Ready check workers
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stopChan:
					return
				default:
					if time.Now().After(deadline) {
						return
					}

					resp, err := http.Get(*serverURL + "/ready")
					if err == nil {
						resp.Body.Close()
						atomic.AddInt32(&readyCount, 1)
					}
					time.Sleep(50 * time.Millisecond)
				}
			}
		}()
	}

	// Progress reporter
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				elapsed := time.Since(start)
				currentTotal := atomic.LoadInt32(&translateCount) + atomic.LoadInt32(&healthCount) + atomic.LoadInt32(&readyCount)
				fmt.Printf("Progress: %.0fs / %.0fs - Total requests: %d (%.2f req/s)\n",
					elapsed.Seconds(), testDuration.Seconds(), currentTotal, float64(currentTotal)/elapsed.Seconds())
			case <-stopChan:
				return
			}
		}
	}()

	wg.Wait()
	close(stopChan)
	actualDuration := time.Since(start)

	totalRequests := translateCount + healthCount + readyCount

	fmt.Printf("\nMixed workload test completed:\n")
	fmt.Printf("  Duration: %v\n", actualDuration)
	fmt.Printf("  Total requests: %d\n", totalRequests)
	fmt.Printf("  Translation requests: %d\n", translateCount)
	fmt.Printf("  Health check requests: %d\n", healthCount)
	fmt.Printf("  Ready check requests: %d\n", readyCount)
	fmt.Printf("  Overall throughput: %.2f req/s\n", float64(totalRequests)/actualDuration.Seconds())

	if totalRequests == 0 {
		fmt.Printf("  WARNING: No successful requests!\n")
	}
}

func printResult(result StressTestResult) {
	fmt.Printf("\n--- %s Result ---\n", result.Name)
	fmt.Printf("Total Requests:  %d\n", result.TotalRequests)
	fmt.Printf("Successful:      %d\n", result.SuccessCount)
	fmt.Printf("Failed:          %d\n", result.FailureCount)
	fmt.Printf("Duration:        %v\n", result.Duration)
	if result.Throughput > 0 {
		fmt.Printf("Throughput:      %.2f req/s\n", result.Throughput)
	}
	fmt.Printf("Failure Rate:    %.2f%%\n", result.FailureRate*100)
	if result.ConcurrentLoad > 0 {
		fmt.Printf("Concurrent Load: %d workers\n", result.ConcurrentLoad)
	}

	if len(result.Errors) > 0 {
		fmt.Printf("\nFirst %d errors:\n", len(result.Errors))
		for i, err := range result.Errors {
			fmt.Printf("  %d. %s\n", i+1, err)
		}
	}

	// Assessment
	if result.FailureRate > 0.1 {
		fmt.Printf("\n⚠️  WARNING: High failure rate (>10%%)\n")
	} else if result.FailureRate > 0 {
		fmt.Printf("\n✓ Acceptable failure rate (<10%%)\n")
	} else {
		fmt.Printf("\n✓ All requests successful!\n")
	}
}
