package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fasthttp/websocket"
	pb "github.com/xxnuo/MTranCore/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TransRequest struct {
	Text string `json:"text"`
	HTML bool   `json:"html"`
}

type StressTestResult struct {
	Name           string
	Protocol       string
	TotalRequests  int64
	SuccessCount   int64
	FailureCount   int64
	Duration       time.Duration
	Throughput     float64
	FailureRate    float64
	ConcurrentLoad int
	Errors         []string
	Latencies      []time.Duration
	MinLatency     time.Duration
	MaxLatency     time.Duration
	AvgLatency     time.Duration
	P50Latency     time.Duration
	P95Latency     time.Duration
	P99Latency     time.Duration
}

var (
	serverURL   = flag.String("url", "http://localhost:8988", "Server URL")
	modelPath   = flag.String("model", "", "Model directory path (if empty, uses ./models/enzh)")
	testType    = flag.String("test", "all", "Test type: all, concurrency, sustained, memory, reload, mixed")
	concurrency = flag.Int("c", 50, "Number of concurrent workers for high-concurrency test")
	duration    = flag.Duration("d", 30*time.Second, "Duration for sustained load test")
	iterations  = flag.Int("n", 1000, "Number of iterations for memory stability test")
	reloads     = flag.Int("r", 5, "Number of reloads for rapid reload test")
	protocol    = flag.String("protocol", "all", "Protocol to use: all, http, grpc, ws")
)

// Realistic test dataset
var realisticTestData = []string{
	// Short sentences
	"Hello",
	"Good morning",
	"How are you?",
	"Thank you very much.",

	// Medium length sentences
	"The quick brown fox jumps over the lazy dog.",
	"Artificial intelligence is transforming the way we live and work.",
	"Machine translation has made significant progress in recent years.",
	"The weather is beautiful today, perfect for a walk in the park.",

	// Long sentences
	"In the field of natural language processing, neural machine translation has become the dominant approach, replacing traditional statistical methods.",
	"The development of large language models has revolutionized the field of artificial intelligence, enabling machines to understand and generate human-like text with unprecedented accuracy.",

	// Paragraphs
	"Climate change is one of the most pressing challenges facing our planet today. Rising temperatures, melting ice caps, and extreme weather events are just some of the consequences we are witnessing. It is crucial that we take immediate action to reduce greenhouse gas emissions and transition to renewable energy sources.",

	// HTML content
	"<p>This is a <strong>test</strong> with <em>HTML</em> tags.</p>",
	"<div>Welcome to our <a href='#'>website</a>. We offer the best services.</div>",

	// Technical text
	"Kubernetes is an open-source container orchestration platform that automates the deployment, scaling, and management of containerized applications.",
	"RESTful APIs use HTTP methods like GET, POST, PUT, and DELETE to perform CRUD operations on resources.",

	// Mixed complexity
	"Hello! Welcome to our platform.",
	"The meeting is scheduled for 3 PM tomorrow in Conference Room B.",
	"Please review the attached documents and provide your feedback by end of day Friday.",
}

// Client interface for different protocols
type Client interface {
	Trans(ctx context.Context, text string, html bool) (string, error)
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

func (c *HTTPClient) Trans(ctx context.Context, text string, html bool) (string, error) {
	reqBody := TransRequest{Text: text, HTML: html}
	body, _ := json.Marshal(reqBody)

	resp, err := c.client.Post(c.baseURL+"/trans", "application/json", bytes.NewReader(body))
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

func (c *GRPCClient) Trans(ctx context.Context, text string, html bool) (string, error) {
	resp, err := c.client.Trans(ctx, &pb.TransRequest{Text: text, Html: html})
	if err != nil {
		return "", err
	}
	if resp.Code != 200 {
		return "", fmt.Errorf("trans failed: %s", resp.Message)
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

func (c *WSClient) Trans(ctx context.Context, text string, html bool) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	msg := map[string]interface{}{
		"type": "trans",
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
		return "", fmt.Errorf("trans failed: %s", resp.Msg)
	}
	return resp.Data.TranslatedText, nil
}

func (c *WSClient) Close() error {
	return c.conn.Close()
}

var client Client

// Get realistic test text
func getTestText(index int) (string, bool) {
	text := realisticTestData[index%len(realisticTestData)]
	// Detect if it's HTML content
	isHTML := len(text) > 0 && text[0] == '<'
	return text, isHTML
}

// Calculate latency statistics
func calculateLatencyStats(latencies []time.Duration) (min, max, avg, p50, p95, p99 time.Duration) {
	if len(latencies) == 0 {
		return
	}

	sorted := make([]time.Duration, len(latencies))
	copy(sorted, latencies)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	min = sorted[0]
	max = sorted[len(sorted)-1]

	var sum time.Duration
	for _, l := range sorted {
		sum += l
	}
	avg = sum / time.Duration(len(sorted))

	p50 = sorted[int(float64(len(sorted))*0.50)]
	p95 = sorted[int(float64(len(sorted))*0.95)]
	p99 = sorted[int(math.Min(float64(len(sorted))*0.99, float64(len(sorted)-1)))]

	return
}

func main() {
	flag.Parse()

	// Determine model path
	modelDir := *modelPath
	if modelDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Printf("Failed to get current directory: %v\n", err)
			os.Exit(1)
		}
		modelDir = filepath.Join(cwd, "models", "enzh")
	}

	// Verify model exists
	if _, err := os.Stat(modelDir); os.IsNotExist(err) {
		fmt.Printf("Model directory not found: %s\n", modelDir)
		os.Exit(1)
	}

	fmt.Printf("╔════════════════════════════════════════════════════╗\n")
	fmt.Printf("║          Stress Test Configuration                ║\n")
	fmt.Printf("╠════════════════════════════════════════════════════╣\n")
	fmt.Printf("║ Server URL: %-39s ║\n", *serverURL)
	fmt.Printf("║ Protocol: %-41s ║\n", *protocol)
	fmt.Printf("║ Model Path: %-39s ║\n", truncateString(modelDir, 39))
	fmt.Printf("║ Test Type: %-40s ║\n", *testType)
	fmt.Printf("║ Test Dataset Size: %-31d ║\n", len(realisticTestData))
	fmt.Printf("╚════════════════════════════════════════════════════╝\n\n")

	// Determine list of protocols to test
	var protocols []string
	if *protocol == "all" {
		protocols = []string{"http", "ws", "grpc"}
	} else {
		protocols = []string{*protocol}
	}

	// Run tests for each protocol
	for _, proto := range protocols {
		fmt.Printf("\n╔════════════════════════════════════════════════════╗\n")
		fmt.Printf("║  Testing Protocol: %-32s ║\n", proto)
		fmt.Printf("╚════════════════════════════════════════════════════╝\n\n")

		// Initialize client based on protocol
		var err error
		switch proto {
		case "http":
			client = NewHTTPClient(*serverURL)
		case "grpc":
			// Extract host:port from URL
			u, err := url.Parse(*serverURL)
			if err != nil {
				fmt.Printf("Invalid server URL: %v\n", err)
				continue
			}
			address := u.Host
			if address == "" {
				address = *serverURL // Assume it's already host:port format
			}
			client, err = NewGRPCClient(address)
			if err != nil {
				fmt.Printf("Failed to create gRPC client: %v\n", err)
				continue
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
				continue
			}
		}

		// Run stress tests based on test type
		switch *testType {
		case "concurrency":
			runHighConcurrencyTest(modelDir, proto)
		case "sustained":
			runSustainedLoadTest(modelDir, proto)
		case "memory":
			runMemoryStabilityTest(modelDir, proto)
		case "reload":
			runRapidReloadTest(modelDir, proto)
		case "mixed":
			runMixedWorkloadTest(modelDir, proto)
		case "all":
			fmt.Printf("=== Running All Stress Tests ===\n\n")
			runHighConcurrencyTest(modelDir, proto)
			fmt.Println()
			runSustainedLoadTest(modelDir, proto)
			fmt.Println()
			runMemoryStabilityTest(modelDir, proto)
			fmt.Println()
			runRapidReloadTest(modelDir, proto)
			fmt.Println()
			runMixedWorkloadTest(modelDir, proto)
		default:
			fmt.Printf("Unknown test type: %s\n", *testType)
		}

		client.Close()

		// Add delay between protocols to ensure server state cleanup
		if len(protocols) > 1 {
			time.Sleep(2 * time.Second)
		}
	}

	fmt.Printf("\n╔════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  All Tests Completed                               ║\n")
	fmt.Printf("╚════════════════════════════════════════════════════╝\n")
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func runHighConcurrencyTest(modelDir string, proto string) {
	fmt.Printf("=== High Concurrency Test ===\n")

	requestsPerWorker := 10
	totalRequests := *concurrency * requestsPerWorker

	var wg sync.WaitGroup
	var successCount, failureCount int32
	var requestCounter int32
	errorChan := make(chan error, totalRequests)
	latencyChan := make(chan time.Duration, totalRequests)

	fmt.Printf("Starting high concurrency test: %d workers, %d requests each\n", *concurrency, requestsPerWorker)
	start := time.Now()

	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < requestsPerWorker; j++ {
				reqIdx := int(atomic.AddInt32(&requestCounter, 1))
				text, isHTML := getTestText(reqIdx)

				reqStart := time.Now()
				_, err := client.Trans(context.Background(), text, isHTML)
				latency := time.Since(reqStart)

				if err != nil {
					atomic.AddInt32(&failureCount, 1)
					errorChan <- fmt.Errorf("worker %d request %d failed: %w", workerID, j, err)
				} else {
					atomic.AddInt32(&successCount, 1)
					latencyChan <- latency
				}
			}
		}(i)
	}

	wg.Wait()
	close(errorChan)
	close(latencyChan)
	testDuration := time.Since(start)

	// Collect errors
	var errors []string
	for err := range errorChan {
		if len(errors) < 10 {
			errors = append(errors, err.Error())
		}
	}

	// Collect latencies
	var latencies []time.Duration
	for lat := range latencyChan {
		latencies = append(latencies, lat)
	}

	result := StressTestResult{
		Name:           "High Concurrency Test",
		Protocol:       proto,
		TotalRequests:  int64(totalRequests),
		SuccessCount:   int64(successCount),
		FailureCount:   int64(failureCount),
		Duration:       testDuration,
		Throughput:     float64(totalRequests) / testDuration.Seconds(),
		FailureRate:    float64(failureCount) / float64(totalRequests),
		ConcurrentLoad: *concurrency,
		Errors:         errors,
		Latencies:      latencies,
	}

	result.MinLatency, result.MaxLatency, result.AvgLatency, result.P50Latency, result.P95Latency, result.P99Latency = calculateLatencyStats(latencies)
	printResult(result)
}

func runSustainedLoadTest(modelDir string, proto string) {
	fmt.Printf("=== Sustained Load Test ===\n")

	workers := 10
	deadline := time.Now().Add(*duration)

	var wg sync.WaitGroup
	var requestCount, successCount, failureCount int32
	stopChan := make(chan struct{})
	latencyChan := make(chan time.Duration, 10000)

	fmt.Printf("Starting sustained load test: %d workers, duration %v\n", workers, *duration)
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

					text, isHTML := getTestText(reqNum)
					reqStart := time.Now()
					_, err := client.Trans(context.Background(), text, isHTML)
					latency := time.Since(reqStart)

					atomic.AddInt32(&requestCount, 1)

					if err != nil {
						atomic.AddInt32(&failureCount, 1)
					} else {
						atomic.AddInt32(&successCount, 1)
						select {
						case latencyChan <- latency:
						default:
							// Channel full, skip this latency
						}
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
				currentSuccess := atomic.LoadInt32(&successCount)
				currentFailure := atomic.LoadInt32(&failureCount)
				fmt.Printf("Progress: %.0fs / %.0fs - Requests: %d (Success: %d, Failed: %d, %.2f req/s)\n",
					elapsed.Seconds(), duration.Seconds(), currentRequests, currentSuccess, currentFailure, float64(currentRequests)/elapsed.Seconds())
			case <-stopChan:
				return
			}
		}
	}()

	wg.Wait()
	close(stopChan)
	close(latencyChan)
	testDuration := time.Since(start)

	// Collect latencies
	var latencies []time.Duration
	for lat := range latencyChan {
		latencies = append(latencies, lat)
	}

	result := StressTestResult{
		Name:           "Sustained Load Test",
		Protocol:       proto,
		TotalRequests:  int64(requestCount),
		SuccessCount:   int64(successCount),
		FailureCount:   int64(failureCount),
		Duration:       testDuration,
		Throughput:     float64(requestCount) / testDuration.Seconds(),
		FailureRate:    float64(failureCount) / float64(requestCount),
		ConcurrentLoad: workers,
		Latencies:      latencies,
	}

	result.MinLatency, result.MaxLatency, result.AvgLatency, result.P50Latency, result.P95Latency, result.P99Latency = calculateLatencyStats(latencies)
	printResult(result)
}

func runMemoryStabilityTest(modelDir string, proto string) {
	fmt.Printf("=== Memory Stability Test ===\n")

	var successCount int32
	latencies := make([]time.Duration, 0, *iterations)
	var mu sync.Mutex

	fmt.Printf("Starting memory stability test with %d iterations\n", *iterations)
	start := time.Now()
	ctx := context.Background()

	for i := 0; i < *iterations; i++ {
		text, isHTML := getTestText(i)
		reqStart := time.Now()
		_, err := client.Trans(ctx, text, isHTML)
		latency := time.Since(reqStart)

		if err == nil {
			atomic.AddInt32(&successCount, 1)
			mu.Lock()
			latencies = append(latencies, latency)
			mu.Unlock()
		}

		// Progress indicator
		if (i+1)%100 == 0 || i+1 == *iterations {
			elapsed := time.Since(start)
			fmt.Printf("Progress: %d/%d (%.1f%%) - Average rate: %.2f req/s\n",
				i+1, *iterations, float64(i+1)/float64(*iterations)*100, float64(i+1)/elapsed.Seconds())
		}
	}

	testDuration := time.Since(start)
	failureCount := int32(*iterations) - successCount

	result := StressTestResult{
		Name:          "Memory Stability Test",
		Protocol:      proto,
		TotalRequests: int64(*iterations),
		SuccessCount:  int64(successCount),
		FailureCount:  int64(failureCount),
		Duration:      testDuration,
		Throughput:    float64(*iterations) / testDuration.Seconds(),
		FailureRate:   float64(failureCount) / float64(*iterations),
		Latencies:     latencies,
	}

	result.MinLatency, result.MaxLatency, result.AvgLatency, result.P50Latency, result.P95Latency, result.P99Latency = calculateLatencyStats(latencies)
	printResult(result)
}

func runRapidReloadTest(modelDir string, proto string) {
	fmt.Printf("=== Rapid Reload Test ===\n")
	fmt.Printf("Testing %d rapid engine reloads\n\n", *reloads)

	var successCount, failureCount int32
	latencies := make([]time.Duration, 0, *reloads)
	var mu sync.Mutex
	start := time.Now()

	ctx := context.Background()
	for i := 0; i < *reloads; i++ {
		fmt.Printf("Reload %d/%d\n", i+1, *reloads)

		// Verify engine is working with a translation
		reloadStart := time.Now()
		text, isHTML := getTestText(i)
		_, err := client.Trans(ctx, text, isHTML)
		latency := time.Since(reloadStart)

		if err != nil {
			fmt.Printf("  Translation failed: %v\n", err)
			atomic.AddInt32(&failureCount, 1)
			continue
		}

		atomic.AddInt32(&successCount, 1)
		mu.Lock()
		latencies = append(latencies, latency)
		mu.Unlock()
		fmt.Printf("  Success (duration: %v)\n", latency)

		// Small delay between reloads
		time.Sleep(100 * time.Millisecond)
	}

	testDuration := time.Since(start)

	result := StressTestResult{
		Name:          "Rapid Reload Test",
		Protocol:      proto,
		TotalRequests: int64(*reloads),
		SuccessCount:  int64(successCount),
		FailureCount:  int64(failureCount),
		Duration:      testDuration,
		FailureRate:   float64(failureCount) / float64(*reloads),
		Latencies:     latencies,
	}

	result.MinLatency, result.MaxLatency, result.AvgLatency, result.P50Latency, result.P95Latency, result.P99Latency = calculateLatencyStats(latencies)
	printResult(result)
}

func runMixedWorkloadTest(modelDir string, proto string) {
	fmt.Printf("=== Mixed Workload Test ===\n")

	testDuration := 20 * time.Second
	deadline := time.Now().Add(testDuration)

	var wg sync.WaitGroup
	         var translateCount, healthCount int32
	         stopChan := make(chan struct{})
	latencyChan := make(chan time.Duration, 1000)

	fmt.Printf("Starting mixed workload test, duration %v\n", testDuration)
	start := time.Now()

	// Translation workers
	for i := 0; i < 5; i++ {
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

					text, isHTML := getTestText(reqNum)
					reqStart := time.Now()
					_, err := client.Trans(context.Background(), text, isHTML)
					latency := time.Since(reqStart)

					if err == nil {
						atomic.AddInt32(&translateCount, 1)
						select {
						case latencyChan <- latency:
						default:
						}
					}
					reqNum++
					time.Sleep(100 * time.Millisecond)
				}
			}
		}(i)
	}

	        // Health check workers (HTTP protocol only)
	        if proto == "http" {
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
	        }
	
	        // Progress reporter
	        go func() {
	                ticker := time.NewTicker(5 * time.Second)
	                defer ticker.Stop()
	                for {
	                        select {
	                        case <-ticker.C:
	                                elapsed := time.Since(start)
	                                currentTotal := atomic.LoadInt32(&translateCount) + atomic.LoadInt32(&healthCount)
	                                fmt.Printf("Progress: %.0fs / %.0fs - Total requests: %d (Translate: %d, Health: %d, %.2f req/s)\n",
	                                        elapsed.Seconds(), testDuration.Seconds(), currentTotal,
	                                        atomic.LoadInt32(&translateCount), atomic.LoadInt32(&healthCount),
	                                        float64(currentTotal)/elapsed.Seconds())
	                        case <-stopChan:
	                                return
	                        }
	                }
	        }()
	
	        wg.Wait()
	        close(stopChan)
	        close(latencyChan)
	        actualDuration := time.Since(start)
	
	        // Collect latencies
	        var latencies []time.Duration
	        for lat := range latencyChan {
	                latencies = append(latencies, lat)
	        }
	
	        totalRequests := translateCount + healthCount
	
	        fmt.Printf("\nMixed workload test completed:\n")
	        fmt.Printf("  Duration: %v\n", actualDuration)
	        fmt.Printf("  Total requests: %d\n", totalRequests)
	        fmt.Printf("  Translation requests: %d\n", translateCount)
	        fmt.Printf("  Health check requests: %d\n", healthCount)
	        fmt.Printf("  Overall throughput: %.2f req/s\n", float64(totalRequests)/actualDuration.Seconds())
	if len(latencies) > 0 {
		min, max, avg, p50, p95, p99 := calculateLatencyStats(latencies)
		fmt.Printf("\nTranslation request latency statistics:\n")
		fmt.Printf("  Min: %v\n", min)
		fmt.Printf("  Max: %v\n", max)
		fmt.Printf("  Average: %v\n", avg)
		fmt.Printf("  P50: %v\n", p50)
		fmt.Printf("  P95: %v\n", p95)
		fmt.Printf("  P99: %v\n", p99)
	}

	if totalRequests == 0 {
		fmt.Printf("\n  Warning: No successful requests!\n")
	}
}

func printResult(result StressTestResult) {
	fmt.Printf("\n╔════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  Test Result: %-37s  ║\n", result.Name)
	fmt.Printf("╠════════════════════════════════════════════════════╣\n")
	fmt.Printf("║ Protocol: %-41s  ║\n", result.Protocol)
	fmt.Printf("╠════════════════════════════════════════════════════╣\n")
	fmt.Printf("║ Basic Statistics                                   ║\n")
	fmt.Printf("╟────────────────────────────────────────────────────╢\n")
	fmt.Printf("║   Total Requests:  %-32d ║\n", result.TotalRequests)
	fmt.Printf("║   Success:         %-32d ║\n", result.SuccessCount)
	fmt.Printf("║   Failed:          %-32d ║\n", result.FailureCount)
	fmt.Printf("║   Duration:        %-32v ║\n", result.Duration)
	if result.Throughput > 0 {
		fmt.Printf("║   Throughput:      %-26.2f req/s ║\n", result.Throughput)
	}
	fmt.Printf("║   Failure Rate:    %-29.2f%% ║\n", result.FailureRate*100)
	if result.ConcurrentLoad > 0 {
		fmt.Printf("║   Concurrent Load: %-26d workers ║\n", result.ConcurrentLoad)
	}

	// Latency statistics
	if len(result.Latencies) > 0 {
		fmt.Printf("╠════════════════════════════════════════════════════╣\n")
		fmt.Printf("║ Latency Statistics                                 ║\n")
		fmt.Printf("╟────────────────────────────────────────────────────╢\n")
		fmt.Printf("║   Min:             %-32v ║\n", result.MinLatency)
		fmt.Printf("║   Max:             %-32v ║\n", result.MaxLatency)
		fmt.Printf("║   Average:         %-32v ║\n", result.AvgLatency)
		fmt.Printf("║   P50 (Median):    %-32v ║\n", result.P50Latency)
		fmt.Printf("║   P95:             %-32v ║\n", result.P95Latency)
		fmt.Printf("║   P99:             %-32v ║\n", result.P99Latency)
		fmt.Printf("║   Samples:         %-32d ║\n", len(result.Latencies))
	}

	if len(result.Errors) > 0 {
		fmt.Printf("╠════════════════════════════════════════════════════╣\n")
		fmt.Printf("║ Error Details (first %d)                           ║\n", len(result.Errors))
		fmt.Printf("╟────────────────────────────────────────────────────╢\n")
		for i, err := range result.Errors {
			errStr := truncateString(err, 48)
			fmt.Printf("║ %d. %-47s║\n", i+1, errStr)
		}
	}

	// Assessment
	fmt.Printf("╠════════════════════════════════════════════════════╣\n")
	fmt.Printf("║ Assessment                                         ║\n")
	fmt.Printf("╟────────────────────────────────────────────────────╢\n")
	if result.FailureRate > 0.1 {
		fmt.Printf("║   Status: Warning - High failure rate (>10%%)      ║\n")
	} else if result.FailureRate > 0 {
		fmt.Printf("║   Status: Good - Acceptable failure rate (<10%%)   ║\n")
	} else {
		fmt.Printf("║   Status: Excellent - All requests succeeded!      ║\n")
	}

	// Performance assessment based on latency
	if len(result.Latencies) > 0 {
		if result.P95Latency < 100*time.Millisecond {
			fmt.Printf("║   Performance: Excellent - P95 latency < 100ms     ║\n")
		} else if result.P95Latency < 500*time.Millisecond {
			fmt.Printf("║   Performance: Good - P95 latency < 500ms          ║\n")
		} else if result.P95Latency < 1*time.Second {
			fmt.Printf("║   Performance: Fair - P95 latency < 1s             ║\n")
		} else {
			fmt.Printf("║   Performance: Poor - P95 latency >= 1s            ║\n")
		}
	}

	fmt.Printf("╚════════════════════════════════════════════════════╝\n")
}
