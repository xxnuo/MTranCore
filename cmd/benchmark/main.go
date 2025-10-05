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
	"strings"
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
	Protocol      string
	Name          string
	TotalRequests int64
	Duration      time.Duration
	Success       int64
	Failed        int64
	Errors        map[string]int64
	Latencies     []time.Duration
	AvgLatency    time.Duration
	MinLatency    time.Duration
	MaxLatency    time.Duration
	P50Latency    time.Duration
	P90Latency    time.Duration
	P95Latency    time.Duration
	P99Latency    time.Duration
	Throughput    float64
}

type TestCase struct {
	Name string
	Text string
	HTML bool
}

var (
	serverURL      = flag.String("url", "http://localhost:8988", "Server URL")
	modelPath      = flag.String("model", "", "Model directory path (if empty, uses ./models/enzh)")
	iterations     = flag.Int("n", 100, "Number of iterations per test")
	concurrency    = flag.Int("c", 1, "Number of concurrent workers")
	testType       = flag.String("test", "all", "Test type: all, compute, html, long, parallel")
	protocol       = flag.String("protocol", "all", "Protocol to use: all, http, grpc, grpc-unix, ws")
	warmup         = flag.Int("warmup", 10, "Number of warmup requests before benchmarking")
	grpcUnixSocket = flag.String("grpc-unix", "", "gRPC Unix socket path (enables gRPC Unix socket testing)")
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
	conn, err := grpc.NewClient(address, 
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		// Performance optimizations matching server settings
		grpc.WithInitialWindowSize(1 << 20),     // 1MB initial window size
		grpc.WithInitialConnWindowSize(1 << 20), // 1MB initial connection window size
		grpc.WithReadBufferSize(32 * 1024),      // 32KB read buffer
		grpc.WithWriteBufferSize(32 * 1024),     // 32KB write buffer
	)
	if err != nil {
		return nil, err
	}
	return &GRPCClient{
		conn:   conn,
		client: pb.NewTranslatorServiceClient(conn),
	}, nil
}

// NewGRPCUnixClient creates a gRPC client using Unix domain socket
func NewGRPCUnixClient(socketPath string) (*GRPCClient, error) {
	conn, err := grpc.NewClient("unix://"+socketPath,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		// Performance optimizations matching server settings
		grpc.WithInitialWindowSize(1 << 20),     // 1MB initial window size
		grpc.WithInitialConnWindowSize(1 << 20), // 1MB initial connection window size
		grpc.WithReadBufferSize(32 * 1024),      // 32KB read buffer
		grpc.WithWriteBufferSize(32 * 1024),     // 32KB write buffer
	)
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

// WSClientPool manages multiple WebSocket connections for concurrent testing
type WSClientPool struct {
	clients []*WSClient
	next    uint32
}

func NewWSClientPool(wsURL string, size int) (*WSClientPool, error) {
	pool := &WSClientPool{
		clients: make([]*WSClient, size),
	}
	for i := 0; i < size; i++ {
		client, err := NewWSClient(wsURL)
		if err != nil {
			// Close already created clients
			for j := 0; j < i; j++ {
				pool.clients[j].Close()
			}
			return nil, err
		}
		pool.clients[i] = client
	}
	return pool, nil
}

func (p *WSClientPool) Get() *WSClient {
	idx := atomic.AddUint32(&p.next, 1) % uint32(len(p.clients))
	return p.clients[idx]
}

func (p *WSClientPool) Close() error {
	for _, client := range p.clients {
		client.Close()
	}
	return nil
}

// Real-world test cases with diverse scenarios
var testCases = []TestCase{
	{
		Name: "Short Greeting",
		Text: "Hello, how are you today?",
		HTML: false,
	},
	{
		Name: "News Headline",
		Text: "Breaking: Scientists discover new approach to renewable energy that could revolutionize power generation worldwide.",
		HTML: false,
	},
	{
		Name: "Product Description",
		Text: "This premium wireless headphone features active noise cancellation, 30-hour battery life, and crystal-clear audio quality. Perfect for travel, work, and entertainment. Includes carrying case and charging cable.",
		HTML: false,
	},
	{
		Name: "Email Message",
		Text: "Dear Team, I hope this message finds you well. I wanted to follow up on our discussion from yesterday's meeting regarding the project timeline. Based on the feedback received, I've updated the schedule and attached the revised document for your review. Please let me know if you have any questions or concerns. Best regards.",
		HTML: false,
	},
	{
		Name: "Technical Article",
		Text: "Machine learning models require large amounts of training data to achieve optimal performance. The process involves data collection, preprocessing, feature engineering, model selection, training, validation, and deployment. Modern frameworks like TensorFlow and PyTorch have simplified this workflow significantly. However, challenges remain in areas such as data quality, model interpretability, and computational efficiency. Researchers continue to develop new algorithms and techniques to address these limitations.",
		HTML: false,
	},
	{
		Name: "Legal Notice",
		Text: "By accessing this website, you agree to be bound by these terms and conditions. The company reserves the right to modify these terms at any time without prior notice. Users are responsible for reviewing the terms periodically. Continued use of the service constitutes acceptance of any changes. All intellectual property rights remain with the company.",
		HTML: false,
	},
	{
		Name: "HTML Article",
		Text: "<article><h1>Welcome to Modern Web Development</h1><p>Learn the latest technologies and best practices for building <strong>amazing web applications</strong>.</p><section><h2>Key Features</h2><ul><li>Responsive design</li><li>Performance optimization</li><li>Security best practices</li><li>Accessibility standards</li></ul></section><p>Join thousands of developers worldwide!</p></article>",
		HTML: true,
	},
	{
		Name: "Medical Information",
		Text: "Patient care requires comprehensive assessment and individualized treatment planning. Healthcare providers must consider medical history, current symptoms, diagnostic test results, and potential contraindications when prescribing medications. Regular monitoring and follow-up appointments are essential to ensure treatment effectiveness and patient safety.",
		HTML: false,
	},
	{
		Name: "Customer Support",
		Text: "Thank you for contacting our support team. We understand you're experiencing issues with your account login. To resolve this problem, please try resetting your password using the 'Forgot Password' link on the login page. If the issue persists, our technical support team is available 24/7 to assist you. You can reach us via phone, email, or live chat.",
		HTML: false,
	},
	{
		Name: "Long Document",
		Text: "The rapid advancement of technology has fundamentally transformed how businesses operate in the modern economy. Cloud computing has enabled organizations to scale their infrastructure dynamically, reducing capital expenditure while improving operational flexibility. Artificial intelligence and machine learning applications are automating routine tasks, enhancing decision-making processes, and creating new opportunities for innovation. However, these technological changes also present challenges. Cybersecurity threats have become more sophisticated, requiring continuous investment in security measures and employee training. Data privacy regulations such as GDPR and CCPA impose strict requirements on how companies collect, store, and process customer information. Furthermore, the increasing reliance on digital systems creates vulnerabilities to system failures and service disruptions. Organizations must develop comprehensive risk management strategies that address these concerns while still embracing technological innovation. The key to success lies in maintaining a balance between leveraging new technologies for competitive advantage and managing the associated risks effectively.",
		HTML: false,
	},
}

func main() {
	flag.Parse()

	// Determine model path
	modelDir := *modelPath
	if modelDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Printf("❌ Error getting current directory: %v\n", err)
			os.Exit(1)
		}
		modelDir = filepath.Join(cwd, "models", "enzh")
	}

	// Verify model exists
	if _, err := os.Stat(modelDir); os.IsNotExist(err) {
		fmt.Printf("❌ Model directory not found: %s\n", modelDir)
		os.Exit(1)
	}

	printHeader()
	fmt.Printf("Server URL:    %s\n", *serverURL)
	fmt.Printf("Protocol(s):   %s\n", *protocol)
	fmt.Printf("Model Path:    %s\n", modelDir)
	fmt.Printf("Iterations:    %d per test\n", *iterations)
	fmt.Printf("Concurrency:   %d workers\n", *concurrency)
	fmt.Printf("Warmup:        %d requests\n", *warmup)
	fmt.Printf("Test Type:     %s\n", *testType)
	fmt.Printf("═══════════════════════════════════════════════════════════\n\n")

	// Determine which protocols to test
	var protocols []string
	if *protocol == "all" {
		protocols = []string{"http", "grpc", "ws"}
		// Add grpc-unix if socket path is provided
		if *grpcUnixSocket != "" {
			protocols = append(protocols, "grpc-unix")
		}
	} else {
		protocols = []string{*protocol}
	}

	// Run benchmarks for each protocol
	allResults := make(map[string][]BenchmarkResult)

	for _, proto := range protocols {
		fmt.Printf("\n╔═══════════════════════════════════════════════════════════╗\n")
		fmt.Printf("║  Testing Protocol: %-40s ║\n", strings.ToUpper(proto))
		fmt.Printf("╚═══════════════════════════════════════════════════════════╝\n\n")

		results := runProtocolBenchmark(proto, modelDir)
		allResults[proto] = results

		// Print results for this protocol
		for _, result := range results {
			printDetailedResult(result)
			fmt.Println()
		}
	}

	// Print comparison if multiple protocols tested
	if len(protocols) > 1 {
		printComparison(allResults)
	}

	fmt.Printf("\n✅ Benchmark completed successfully!\n")
}

func runProtocolBenchmark(proto, modelDir string) []BenchmarkResult {
	fmt.Printf("🔌 Establishing %s connection...\n", strings.ToUpper(proto))

	client, wsPool, err := createClient(proto)
	if err != nil {
		fmt.Printf("❌ Failed to create %s client: %v\n", proto, err)
		return nil
	}

	fmt.Printf("✅ Connection established successfully!\n")

	if wsPool != nil {
		defer wsPool.Close()
		fmt.Printf("📡 WebSocket pool size: %d connections\n", *concurrency)
	} else if client != nil {
		defer client.Close()
	}

	// Load engine
	fmt.Printf("\n📦 Loading translation engine...\n")
	fmt.Printf("   Model path: %s\n", modelDir)
	fmt.Printf("   Protocol: %s\n", strings.ToUpper(proto))

	loadStart := time.Now()
	ctx := context.Background()

	var poweronClient Client
	if wsPool != nil {
		poweronClient = wsPool.Get()
	} else {
		poweronClient = client
	}

	if err := poweronClient.Poweron(ctx, modelDir); err != nil {
		fmt.Printf("❌ Failed to load engine: %v\n", err)
		return nil
	}
	loadDuration := time.Since(loadStart)
	fmt.Printf("✅ Engine loaded successfully in %s!\n\n", formatDuration(loadDuration))

	// Warmup
	if *warmup > 0 {
		fmt.Printf("🔥 Warming up with %d requests...\n", *warmup)
		warmupStart := time.Now()
		successCount := 0

		for i := 0; i < *warmup; i++ {
			var warmupClient Client
			if wsPool != nil {
				warmupClient = wsPool.Get()
			} else {
				warmupClient = client
			}
			_, err := warmupClient.Compute(ctx, "Hello, world!", false)
			if err == nil {
				successCount++
			}

			if (i+1)%5 == 0 || i+1 == *warmup {
				fmt.Printf("\r   Progress: %d/%d", i+1, *warmup)
			}
		}

		warmupDuration := time.Since(warmupStart)
		fmt.Printf("\r   Progress: %d/%d - ✅ Completed in %s (Success: %d/%d)\n\n",
			*warmup, *warmup, formatDuration(warmupDuration), successCount, *warmup)
	}

	// Run benchmarks based on test type
	var results []BenchmarkResult

	fmt.Printf("🚀 Starting benchmark tests...\n")
	fmt.Printf("   Test type: %s\n", *testType)
	fmt.Printf("   Iterations per test: %d\n", *iterations)
	fmt.Printf("   Concurrency: %d\n\n", *concurrency)

	switch *testType {
	case "all":
		fmt.Printf("📋 Running all %d test cases:\n\n", len(testCases))
		for i, tc := range testCases {
			fmt.Printf("═══════════════════ Test %d/%d ═══════════════════\n", i+1, len(testCases))
			result := benchmarkTest(proto, tc, client, wsPool)
			results = append(results, result)
		}
		// Add parallel test if concurrency > 1
		if *concurrency > 1 {
			fmt.Printf("═══════════════════ Parallel Test ═══════════════════\n")
			result := benchmarkParallel(proto, client, wsPool)
			results = append(results, result)
		}
	case "parallel":
		result := benchmarkParallel(proto, client, wsPool)
		results = append(results, result)
	default:
		// Run specific test case
		for _, tc := range testCases {
			if strings.EqualFold(tc.Name, *testType) ||
				(*testType == "compute" && tc.Name == "Short Greeting") ||
				(*testType == "html" && tc.HTML) {
				result := benchmarkTest(proto, tc, client, wsPool)
				results = append(results, result)
				break
			}
		}
	}

	fmt.Printf("\n✅ All tests completed for %s protocol!\n", strings.ToUpper(proto))

	return results
}

func createClient(proto string) (Client, *WSClientPool, error) {
	switch proto {
	case "http":
		return NewHTTPClient(*serverURL), nil, nil

	case "grpc":
		u, err := url.Parse(*serverURL)
		if err != nil {
			return nil, nil, err
		}
		address := u.Host
		if address == "" {
			address = *serverURL
		}
		client, err := NewGRPCClient(address)
		return client, nil, err

	case "grpc-unix":
		if *grpcUnixSocket == "" {
			return nil, nil, fmt.Errorf("grpc-unix requires --grpc-unix flag with socket path")
		}
		client, err := NewGRPCUnixClient(*grpcUnixSocket)
		return client, nil, err

	case "ws":
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

		// Create connection pool for concurrent tests
		poolSize := *concurrency
		if poolSize < 1 {
			poolSize = 1
		}
		pool, err := NewWSClientPool(wsURL, poolSize)
		return nil, pool, err

	default:
		return nil, nil, fmt.Errorf("unknown protocol: %s", proto)
	}
}

func benchmarkTest(proto string, tc TestCase, client Client, wsPool *WSClientPool) BenchmarkResult {
	result := BenchmarkResult{
		Protocol:   proto,
		Name:       tc.Name,
		MinLatency: time.Hour,
		Errors:     make(map[string]int64),
		Latencies:  make([]time.Duration, 0, *iterations),
	}

	fmt.Printf("📊 Running test: %s\n", tc.Name)
	fmt.Printf("   Text length: %d chars, HTML: %v\n", len(tc.Text), tc.HTML)

	// Show text preview
	preview := tc.Text
	if len(preview) > 80 {
		preview = preview[:77] + "..."
	}
	fmt.Printf("   Preview: %s\n", preview)

	start := time.Now()
	ctx := context.Background()
	var latencyMutex sync.Mutex
	var sampleTranslation string

	for i := 0; i < *iterations; i++ {
		var testClient Client
		if wsPool != nil {
			testClient = wsPool.Get()
		} else {
			testClient = client
		}

		reqStart := time.Now()
		translated, err := testClient.Compute(ctx, tc.Text, tc.HTML)
		latency := time.Since(reqStart)

		latencyMutex.Lock()
		result.Latencies = append(result.Latencies, latency)
		latencyMutex.Unlock()

		if err != nil {
			atomic.AddInt64(&result.Failed, 1)
			errMsg := err.Error()
			if len(errMsg) > 50 {
				errMsg = errMsg[:50] + "..."
			}
			latencyMutex.Lock()
			result.Errors[errMsg]++
			latencyMutex.Unlock()
		} else {
			atomic.AddInt64(&result.Success, 1)
			// Save first successful translation as sample
			if i == 0 && translated != "" {
				sampleTranslation = translated
			}
		}

		if latency < result.MinLatency {
			result.MinLatency = latency
		}
		if latency > result.MaxLatency {
			result.MaxLatency = latency
		}

		// More frequent progress updates with statistics
		if (i+1)%10 == 0 || i+1 == *iterations {
			currentSuccessRate := float64(result.Success) / float64(i+1) * 100
			avgLatencySoFar := time.Duration(0)
			if len(result.Latencies) > 0 {
				var total time.Duration
				for _, l := range result.Latencies {
					total += l
				}
				avgLatencySoFar = total / time.Duration(len(result.Latencies))
			}
			fmt.Printf("\r   Progress: %d/%d | Success: %.1f%% | Avg Latency: %s   ",
				i+1, *iterations, currentSuccessRate, formatDuration(avgLatencySoFar))
		}
	}

	result.Duration = time.Since(start)
	result.TotalRequests = int64(*iterations)
	calculateStats(&result)

	fmt.Printf("\n")

	// Show sample translation result
	if sampleTranslation != "" {
		fmt.Printf("   Sample translation: ")
		if len(sampleTranslation) > 80 {
			fmt.Printf("%s...\n", sampleTranslation[:77])
		} else {
			fmt.Printf("%s\n", sampleTranslation)
		}
	}

	fmt.Printf("\n")
	return result
}

func benchmarkParallel(proto string, client Client, wsPool *WSClientPool) BenchmarkResult {
	result := BenchmarkResult{
		Protocol:   proto,
		Name:       "Parallel Concurrent Load",
		MinLatency: time.Hour,
		Errors:     make(map[string]int64),
		Latencies:  make([]time.Duration, 0, *iterations),
	}

	var wg sync.WaitGroup
	var latencyMutex sync.Mutex
	var completed atomic.Int64
	totalPerWorker := *iterations / *concurrency

	fmt.Printf("📊 Running parallel benchmark with %d concurrent workers\n", *concurrency)
	fmt.Printf("   %d requests per worker, %d total requests\n", totalPerWorker, totalPerWorker**concurrency)
	fmt.Printf("   Test text: \"Hello, world! This is a parallel test.\"\n")

	start := time.Now()
	ctx := context.Background()

	// Progress reporter
	done := make(chan bool)
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				current := completed.Load()
				total := int64(totalPerWorker * *concurrency)
				progress := float64(current) / float64(total) * 100
				successRate := float64(result.Success) / float64(current+1) * 100
				if current > 0 {
					fmt.Printf("\r   Progress: %d/%d (%.1f%%) | Success: %.1f%%   ",
						current, total, progress, successRate)
				}
			}
		}
	}()

	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < totalPerWorker; j++ {
				var testClient Client
				if wsPool != nil {
					testClient = wsPool.Get()
				} else {
					testClient = client
				}

				reqStart := time.Now()
				_, err := testClient.Compute(ctx, "Hello, world! This is a parallel test.", false)
				latency := time.Since(reqStart)

				latencyMutex.Lock()
				result.Latencies = append(result.Latencies, latency)
				if latency < result.MinLatency {
					result.MinLatency = latency
				}
				if latency > result.MaxLatency {
					result.MaxLatency = latency
				}
				latencyMutex.Unlock()

				if err != nil {
					atomic.AddInt64(&result.Failed, 1)
					errMsg := err.Error()
					if len(errMsg) > 50 {
						errMsg = errMsg[:50] + "..."
					}
					latencyMutex.Lock()
					result.Errors[errMsg]++
					latencyMutex.Unlock()
				} else {
					atomic.AddInt64(&result.Success, 1)
				}

				completed.Add(1)
			}
		}(i)
	}

	wg.Wait()
	done <- true
	close(done)

	result.Duration = time.Since(start)
	result.TotalRequests = int64(totalPerWorker * *concurrency)
	calculateStats(&result)

	fmt.Printf("\r   Progress: %d/%d (100.0%%) | Success: %.1f%%   \n",
		result.TotalRequests, result.TotalRequests,
		float64(result.Success)/float64(result.TotalRequests)*100)
	fmt.Printf("   ✅ Completed in %s!\n\n", formatDuration(result.Duration))
	return result
}

func calculateStats(result *BenchmarkResult) {
	if len(result.Latencies) == 0 {
		return
	}

	// Sort latencies for percentile calculation
	sorted := make([]time.Duration, len(result.Latencies))
	copy(sorted, result.Latencies)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	// Calculate average
	var total time.Duration
	for _, lat := range sorted {
		total += lat
	}
	result.AvgLatency = total / time.Duration(len(sorted))

	// Calculate percentiles
	result.P50Latency = sorted[int(float64(len(sorted))*0.50)]
	result.P90Latency = sorted[int(float64(len(sorted))*0.90)]
	result.P95Latency = sorted[int(float64(len(sorted))*0.95)]
	result.P99Latency = sorted[int(math.Min(float64(len(sorted))*0.99, float64(len(sorted)-1)))]

	// Calculate throughput
	result.Throughput = float64(result.TotalRequests) / result.Duration.Seconds()
}

func printHeader() {
	fmt.Printf("\n")
	fmt.Printf("╔═══════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║       🚀 Translation Service Benchmark Tool 🚀           ║\n")
	fmt.Printf("╚═══════════════════════════════════════════════════════════╝\n")
	fmt.Printf("\n═══════════════════ Configuration ═════════════════════════\n")
}

func printDetailedResult(result BenchmarkResult) {
	fmt.Printf("┌───────────────────────────────────────────────────────────┐\n")
	fmt.Printf("│ Test: %-52s│\n", result.Name)
	fmt.Printf("│ Protocol: %-48s│\n", strings.ToUpper(result.Protocol))
	fmt.Printf("├───────────────────────────────────────────────────────────┤\n")

	// Request statistics
	successRate := float64(result.Success) / float64(result.TotalRequests) * 100
	failureRate := float64(result.Failed) / float64(result.TotalRequests) * 100
	fmt.Printf("│ 📊 Request Statistics:                                   │\n")
	fmt.Printf("│   Total Requests:    %-33d│\n", result.TotalRequests)
	fmt.Printf("│   ✅ Successful:     %-33d│\n", result.Success)
	fmt.Printf("│   ❌ Failed:         %-33d│\n", result.Failed)
	fmt.Printf("│   Success Rate:      %-30.2f%% │\n", successRate)
	if result.Failed > 0 {
		fmt.Printf("│   Failure Rate:      %-30.2f%% │\n", failureRate)
	}
	fmt.Printf("│                                                           │\n")

	// Timing statistics
	fmt.Printf("│ ⏱️  Timing Statistics:                                    │\n")
	fmt.Printf("│   Total Duration:    %-33s│\n", formatDuration(result.Duration))
	fmt.Printf("│   Throughput:        %-28.2f req/s │\n", result.Throughput)
	avgTimePerReq := result.Duration / time.Duration(result.TotalRequests)
	fmt.Printf("│   Avg Time/Request:  %-33s│\n", formatDuration(avgTimePerReq))
	fmt.Printf("│                                                           │\n")

	// Latency statistics with more detail
	fmt.Printf("│ 📈 Latency Distribution:                                 │\n")
	fmt.Printf("│   Min:               %-33s│\n", formatDuration(result.MinLatency))
	fmt.Printf("│   Average:           %-33s│\n", formatDuration(result.AvgLatency))
	fmt.Printf("│   Median (P50):      %-33s│\n", formatDuration(result.P50Latency))
	fmt.Printf("│   P90:               %-33s│\n", formatDuration(result.P90Latency))
	fmt.Printf("│   P95:               %-33s│\n", formatDuration(result.P95Latency))
	fmt.Printf("│   P99:               %-33s│\n", formatDuration(result.P99Latency))
	fmt.Printf("│   Max:               %-33s│\n", formatDuration(result.MaxLatency))

	// Calculate latency spread
	latencySpread := result.MaxLatency - result.MinLatency
	fmt.Printf("│   Spread (Max-Min):  %-33s│\n", formatDuration(latencySpread))

	// Performance indicators
	fmt.Printf("│                                                           │\n")
	fmt.Printf("│ 🎯 Performance Indicators:                               │\n")

	// Consistency score (based on P95/P50 ratio)
	if result.P50Latency > 0 {
		consistencyRatio := float64(result.P95Latency) / float64(result.P50Latency)
		consistency := "Excellent"
		if consistencyRatio > 2.0 {
			consistency = "Poor"
		} else if consistencyRatio > 1.5 {
			consistency = "Fair"
		} else if consistencyRatio > 1.2 {
			consistency = "Good"
		}
		fmt.Printf("│   Consistency:       %-24s (%.2fx) │\n", consistency, consistencyRatio)
	}

	// Throughput category
	throughputCategory := "Low"
	if result.Throughput > 100 {
		throughputCategory = "Excellent"
	} else if result.Throughput > 50 {
		throughputCategory = "High"
	} else if result.Throughput > 20 {
		throughputCategory = "Good"
	} else if result.Throughput > 10 {
		throughputCategory = "Moderate"
	}
	fmt.Printf("│   Throughput Class:  %-33s│\n", throughputCategory)

	// Error details if any
	if len(result.Errors) > 0 {
		fmt.Printf("│                                                           │\n")
		fmt.Printf("│ ❌ Error Details:                                         │\n")
		for errMsg, count := range result.Errors {
			if len(errMsg) > 45 {
				errMsg = errMsg[:42] + "..."
			}
			percentage := float64(count) / float64(result.TotalRequests) * 100
			fmt.Printf("│   %-40s: %4d (%.1f%%) │\n", errMsg, count, percentage)
		}
	}

	fmt.Printf("└───────────────────────────────────────────────────────────┘\n")
}

func printComparison(allResults map[string][]BenchmarkResult) {
	fmt.Printf("\n\n")
	fmt.Printf("╔═══════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║              📊 Protocol Comparison Summary               ║\n")
	fmt.Printf("╚═══════════════════════════════════════════════════════════╝\n\n")

	// Group results by test name
	testGroups := make(map[string]map[string]BenchmarkResult)
	for proto, results := range allResults {
		for _, result := range results {
			if testGroups[result.Name] == nil {
				testGroups[result.Name] = make(map[string]BenchmarkResult)
			}
			testGroups[result.Name][proto] = result
		}
	}

	// Print comparison for each test
	testCount := 0
	for testName, protoResults := range testGroups {
		testCount++
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("Test %d: %s\n", testCount, testName)
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

		fmt.Printf("%-12s │ %8s │ %12s │ %10s │ %10s │ %10s │ %10s\n",
			"Protocol", "Success", "Throughput", "Avg", "P50", "P95", "P99")
		fmt.Printf("─────────────┼──────────┼──────────────┼────────────┼────────────┼────────────┼──────────\n")

		// Calculate best values for highlighting
		bestThroughput := 0.0
		bestAvgLatency := time.Hour
		bestP95Latency := time.Hour
		for _, result := range protoResults {
			if result.Throughput > bestThroughput {
				bestThroughput = result.Throughput
			}
			if result.AvgLatency < bestAvgLatency && result.AvgLatency > 0 {
				bestAvgLatency = result.AvgLatency
			}
			if result.P95Latency < bestP95Latency && result.P95Latency > 0 {
				bestP95Latency = result.P95Latency
			}
		}

		// Collect and sort protocol names
		protocols := make([]string, 0, len(protoResults))
		for proto := range protoResults {
			protocols = append(protocols, proto)
		}
		sort.Strings(protocols)

		for _, proto := range protocols {
			result := protoResults[proto]
			marker := ""
			// Highlight the best performer
			if result.Throughput == bestThroughput && result.AvgLatency == bestAvgLatency {
				marker = " 🏆"
			} else if result.Throughput == bestThroughput {
				marker = " ⚡"
			} else if result.AvgLatency == bestAvgLatency {
				marker = " 🎯"
			}

			successRate := float64(result.Success) / float64(result.TotalRequests) * 100
			fmt.Printf("%-12s │ %6.1f%% │ %10.2f/s │ %10s │ %10s │ %10s │ %10s%s\n",
				strings.ToUpper(proto),
				successRate,
				result.Throughput,
				formatDuration(result.AvgLatency),
				formatDuration(result.P50Latency),
				formatDuration(result.P95Latency),
				formatDuration(result.P99Latency),
				marker,
			)
		}

		// Add performance difference analysis
		fmt.Printf("\n")
		if len(protoResults) > 1 {
			fmt.Printf("Performance Analysis:\n")

			// Find best and worst
			var bestProto, worstProto string
			bestThroughputVal := 0.0
			worstThroughputVal := 1e9

			for proto, result := range protoResults {
				if result.Throughput > bestThroughputVal {
					bestThroughputVal = result.Throughput
					bestProto = proto
				}
				if result.Throughput < worstThroughputVal {
					worstThroughputVal = result.Throughput
					worstProto = proto
				}
			}

			if bestProto != worstProto && worstThroughputVal > 0 {
				improvement := ((bestThroughputVal - worstThroughputVal) / worstThroughputVal) * 100
				fmt.Printf("  • %s is %.1f%% faster than %s in throughput\n",
					strings.ToUpper(bestProto), improvement, strings.ToUpper(worstProto))
			}

			// Latency comparison
			bestLatProto := ""
			bestLatVal := time.Hour
			for proto, result := range protoResults {
				if result.P95Latency < bestLatVal && result.P95Latency > 0 {
					bestLatVal = result.P95Latency
					bestLatProto = proto
				}
			}
			if bestLatProto != "" {
				fmt.Printf("  • %s has the best P95 latency: %s\n",
					strings.ToUpper(bestLatProto), formatDuration(bestLatVal))
			}
		}

		fmt.Printf("\nLegend: 🏆 Best Overall | ⚡ Highest Throughput | 🎯 Lowest Latency\n")
		fmt.Printf("\n")
	}

	// Overall summary
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("Overall Protocol Performance Summary\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	protocolStats := make(map[string]struct {
		avgThroughput float64
		avgLatency    time.Duration
		wins          int
		tests         int
	})

	for _, protoResults := range testGroups {
		// Find winner for each test
		bestTP := 0.0
		winnerProto := ""
		for proto, result := range protoResults {
			if result.Throughput > bestTP {
				bestTP = result.Throughput
				winnerProto = proto
			}
		}

		// Accumulate stats
		for proto, result := range protoResults {
			stats := protocolStats[proto]
			stats.avgThroughput += result.Throughput
			stats.avgLatency += result.AvgLatency
			stats.tests++
			if proto == winnerProto {
				stats.wins++
			}
			protocolStats[proto] = stats
		}
	}

	fmt.Printf("%-12s │ %12s │ %12s │ %10s\n", "Protocol", "Avg Throughput", "Avg Latency", "Wins")
	fmt.Printf("─────────────┼──────────────┼──────────────┼──────────\n")

	// Collect and sort protocol names
	sortedProtocols := make([]string, 0, len(protocolStats))
	for proto := range protocolStats {
		sortedProtocols = append(sortedProtocols, proto)
	}
	sort.Strings(sortedProtocols)

	for _, proto := range sortedProtocols {
		stats := protocolStats[proto]
		if stats.tests > 0 {
			avgTP := stats.avgThroughput / float64(stats.tests)
			avgLat := stats.avgLatency / time.Duration(stats.tests)
			fmt.Printf("%-12s │ %10.2f/s │ %12s │ %d/%d\n",
				strings.ToUpper(proto),
				avgTP,
				formatDuration(avgLat),
				stats.wins,
				stats.tests,
			)
		}
	}
	fmt.Printf("\n")
}

func formatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%dns", d.Nanoseconds())
	} else if d < time.Millisecond {
		return fmt.Sprintf("%.2fµs", float64(d.Nanoseconds())/1000.0)
	} else if d < time.Second {
		return fmt.Sprintf("%.2fms", float64(d.Microseconds())/1000.0)
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}
