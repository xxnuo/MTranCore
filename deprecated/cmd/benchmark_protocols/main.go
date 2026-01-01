package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fasthttp/websocket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/xxnuo/MTranCore/proto"
)

type BenchmarkResult struct {
	Protocol        string
	TotalRequests   int
	SuccessCount    int64
	FailureCount    int64
	TotalDuration   time.Duration
	AvgLatency      time.Duration
	MinLatency      time.Duration
	MaxLatency      time.Duration
	P50Latency      time.Duration
	P95Latency      time.Duration
	P99Latency      time.Duration
	RequestsPerSec  float64
}

type Config struct {
	ServerAddr      string
	UnixSocket      string
	NumRequests     int
	Concurrency     int
	TestText        string
	TestHTML        bool
	WarmupRequests  int
}

var (
	serverAddr     = flag.String("server", "localhost:8080", "Server address")
	unixSocket     = flag.String("unix-socket", "", "Unix socket path for gRPC (optional)")
	numRequests    = flag.Int("requests", 10000, "Total number of requests")
	concurrency    = flag.Int("concurrent", 100, "Number of concurrent workers")
	testText       = flag.String("text", "Hello, world! This is a test message.", "Text to translate")
	testHTML       = flag.Bool("html", false, "Enable HTML mode")
	warmupRequests = flag.Int("warmup", 100, "Number of warmup requests")
	verbose        = flag.Bool("verbose", false, "Enable verbose error logging")
)

func main() {
	flag.Parse()

	cfg := Config{
		ServerAddr:     *serverAddr,
		UnixSocket:     *unixSocket,
		NumRequests:    *numRequests,
		Concurrency:    *concurrency,
		TestText:       *testText,
		TestHTML:       *testHTML,
		WarmupRequests: *warmupRequests,
	}

	fmt.Printf("=== Protocol Performance Benchmark ===\n")
	fmt.Printf("Server: %s\n", cfg.ServerAddr)
	fmt.Printf("Total Requests: %d\n", cfg.NumRequests)
	fmt.Printf("Concurrency: %d\n", cfg.Concurrency)
	fmt.Printf("Warmup Requests: %d\n", cfg.WarmupRequests)
	fmt.Printf("Test Text: %s\n", cfg.TestText)
	fmt.Printf("HTML Mode: %v\n\n", cfg.TestHTML)

	fmt.Printf("Checking server availability...\n")
	if !checkServerHealth(cfg.ServerAddr) {
		fmt.Printf("Error: Cannot connect to server at %s\n", cfg.ServerAddr)
		fmt.Printf("Please ensure the MTranCore worker is running with:\n")
		fmt.Printf("  go run ./cmd/worker -enable-http -enable-ws -enable-grpc\n")
		return
	}
	fmt.Printf("Server is available.\n\n")

	if cfg.WarmupRequests > 0 {
		fmt.Printf("Warming up...\n")
		warmupCfg := cfg
		warmupCfg.NumRequests = cfg.WarmupRequests
		warmupCfg.Concurrency = min(cfg.Concurrency, cfg.WarmupRequests)

		if *verbose {
			fmt.Printf("  Running %d warmup requests...\n", warmupCfg.NumRequests)
		}

		benchmarkHTTP(warmupCfg, true)
		benchmarkWebSocket(warmupCfg, true)
		benchmarkGRPC(warmupCfg, true)
		if cfg.UnixSocket != "" {
			benchmarkGRPCUnix(warmupCfg, true)
		}

		time.Sleep(2 * time.Second)
		fmt.Printf("\n")
	}

	results := make([]BenchmarkResult, 0)

	fmt.Printf("Testing HTTP...\n")
	if result := benchmarkHTTP(cfg, false); result != nil {
		results = append(results, *result)
		printResult(*result)
	}
	time.Sleep(1 * time.Second)

	fmt.Printf("\nTesting WebSocket...\n")
	if result := benchmarkWebSocket(cfg, false); result != nil {
		results = append(results, *result)
		printResult(*result)
	}
	time.Sleep(1 * time.Second)

	fmt.Printf("\nTesting gRPC...\n")
	if result := benchmarkGRPC(cfg, false); result != nil {
		results = append(results, *result)
		printResult(*result)
	}

	if cfg.UnixSocket != "" {
		time.Sleep(1 * time.Second)
		fmt.Printf("\nTesting gRPC (Unix Socket)...\n")
		if result := benchmarkGRPCUnix(cfg, false); result != nil {
			results = append(results, *result)
			printResult(*result)
		}
	}

	fmt.Printf("\n=== Summary ===\n")
	printComparison(results)
}

func benchmarkHTTP(cfg Config, silent bool) *BenchmarkResult {
	var successCount, failureCount int64
	latencies := make([]time.Duration, 0, cfg.NumRequests)
	var latenciesMu sync.Mutex
	var firstError error
	var firstErrorOnce sync.Once

	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        cfg.Concurrency * 2,
			MaxIdleConnsPerHost: cfg.Concurrency * 2,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	startTime := time.Now()

	var wg sync.WaitGroup
	requestChan := make(chan int, cfg.NumRequests)

	for i := 0; i < cfg.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for reqNum := range requestChan {
				reqStart := time.Now()

				reqBody := fmt.Sprintf(`{"text":"%s #%d","html":%v}`, cfg.TestText, reqNum, cfg.TestHTML)
				resp, err := client.Post(
					fmt.Sprintf("http://%s/trans", cfg.ServerAddr),
					"application/json",
					strings.NewReader(reqBody),
				)

				if err != nil {
					atomic.AddInt64(&failureCount, 1)
					firstErrorOnce.Do(func() {
						firstError = err
					})
					continue
				}

				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()

				latency := time.Since(reqStart)
				if resp.StatusCode == http.StatusOK {
					atomic.AddInt64(&successCount, 1)
					latenciesMu.Lock()
					latencies = append(latencies, latency)
					latenciesMu.Unlock()
				} else {
					atomic.AddInt64(&failureCount, 1)
					firstErrorOnce.Do(func() {
						firstError = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
					})
				}
			}
		}()
	}

	for i := 0; i < cfg.NumRequests; i++ {
		requestChan <- i
	}
	close(requestChan)

	wg.Wait()
	totalDuration := time.Since(startTime)

	if silent {
		return nil
	}

	if firstError != nil && *verbose {
		fmt.Printf("  [HTTP] First error: %v\n", firstError)
	}

	return calculateResult("HTTP", cfg.NumRequests, successCount, failureCount, totalDuration, latencies)
}

func benchmarkWebSocket(cfg Config, silent bool) *BenchmarkResult {
	var successCount, failureCount int64
	latencies := make([]time.Duration, 0, cfg.NumRequests)
	var latenciesMu sync.Mutex
	var firstError error
	var firstErrorOnce sync.Once

	dialer := &websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	startTime := time.Now()

	var wg sync.WaitGroup
	requestChan := make(chan int, cfg.NumRequests)

	for i := 0; i < cfg.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			conn, _, err := dialer.Dial(fmt.Sprintf("ws://%s/ws", cfg.ServerAddr), nil)
			if err != nil {
				atomic.AddInt64(&failureCount, int64(cfg.NumRequests/cfg.Concurrency))
				firstErrorOnce.Do(func() {
					firstError = err
				})
				return
			}
			defer conn.Close()

			for reqNum := range requestChan {
				reqStart := time.Now()

				req := map[string]interface{}{
					"type": "trans",
					"data": map[string]interface{}{
						"text": fmt.Sprintf("%s #%d", cfg.TestText, reqNum),
						"html": cfg.TestHTML,
					},
				}

				if err := conn.WriteJSON(req); err != nil {
					atomic.AddInt64(&failureCount, 1)
					continue
				}

				var resp map[string]interface{}
				if err := conn.ReadJSON(&resp); err != nil {
					atomic.AddInt64(&failureCount, 1)
					continue
				}

				latency := time.Since(reqStart)
				if code, ok := resp["code"].(float64); ok && int(code) == 200 {
					atomic.AddInt64(&successCount, 1)
					latenciesMu.Lock()
					latencies = append(latencies, latency)
					latenciesMu.Unlock()
				} else {
					atomic.AddInt64(&failureCount, 1)
					firstErrorOnce.Do(func() {
						msg := "unknown error"
						if msgStr, ok := resp["msg"].(string); ok {
							msg = msgStr
						}
						firstError = fmt.Errorf("WebSocket code=%v: %s", code, msg)
					})
				}
			}
		}()
	}

	for i := 0; i < cfg.NumRequests; i++ {
		requestChan <- i
	}
	close(requestChan)

	wg.Wait()
	totalDuration := time.Since(startTime)

	if silent {
		return nil
	}

	if firstError != nil && *verbose {
		fmt.Printf("  [WebSocket] First error: %v\n", firstError)
	}

	return calculateResult("WebSocket", cfg.NumRequests, successCount, failureCount, totalDuration, latencies)
}

func benchmarkGRPC(cfg Config, silent bool) *BenchmarkResult {
	var successCount, failureCount int64
	latencies := make([]time.Duration, 0, cfg.NumRequests)
	var latenciesMu sync.Mutex
	var firstError error
	var firstErrorOnce sync.Once

	startTime := time.Now()

	var wg sync.WaitGroup
	requestChan := make(chan int, cfg.NumRequests)

	for i := 0; i < cfg.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			conn, err := grpc.NewClient(
				cfg.ServerAddr,
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			if err != nil {
				atomic.AddInt64(&failureCount, int64(cfg.NumRequests/cfg.Concurrency))
				firstErrorOnce.Do(func() {
					firstError = err
				})
				return
			}
			defer conn.Close()

			client := pb.NewTranslatorServiceClient(conn)

			for reqNum := range requestChan {
				reqStart := time.Now()

				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				resp, err := client.Trans(ctx, &pb.TransRequest{
					Text: fmt.Sprintf("%s #%d", cfg.TestText, reqNum),
					Html: cfg.TestHTML,
				})
				cancel()

				if err != nil {
					atomic.AddInt64(&failureCount, 1)
					firstErrorOnce.Do(func() {
						firstError = err
					})
					continue
				}

				latency := time.Since(reqStart)
				if resp.Code == 200 {
					atomic.AddInt64(&successCount, 1)
					latenciesMu.Lock()
					latencies = append(latencies, latency)
					latenciesMu.Unlock()
				} else {
					atomic.AddInt64(&failureCount, 1)
					firstErrorOnce.Do(func() {
						firstError = fmt.Errorf("gRPC code=%d: %s", resp.Code, resp.Message)
					})
				}
			}
		}()
	}

	for i := 0; i < cfg.NumRequests; i++ {
		requestChan <- i
	}
	close(requestChan)

	wg.Wait()
	totalDuration := time.Since(startTime)

	if silent {
		return nil
	}

	if firstError != nil && *verbose {
		fmt.Printf("  [gRPC] First error: %v\n", firstError)
	}

	return calculateResult("gRPC", cfg.NumRequests, successCount, failureCount, totalDuration, latencies)
}

func benchmarkGRPCUnix(cfg Config, silent bool) *BenchmarkResult {
	var successCount, failureCount int64
	latencies := make([]time.Duration, 0, cfg.NumRequests)
	var latenciesMu sync.Mutex
	var firstError error
	var firstErrorOnce sync.Once

	startTime := time.Now()

	var wg sync.WaitGroup
	requestChan := make(chan int, cfg.NumRequests)

	unixAddr := cfg.UnixSocket
	if !strings.HasPrefix(unixAddr, "/") {
		var err error
		unixAddr, err = filepath.Abs(unixAddr)
		if err != nil {
			return &BenchmarkResult{
				Protocol:      "gRPC (Unix)",
				TotalRequests: cfg.NumRequests,
				FailureCount:  int64(cfg.NumRequests),
				TotalDuration: time.Since(startTime),
			}
		}
	}

	for i := 0; i < cfg.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			conn, err := grpc.NewClient(
				"unix:"+unixAddr,
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			if err != nil {
				atomic.AddInt64(&failureCount, int64(cfg.NumRequests/cfg.Concurrency))
				firstErrorOnce.Do(func() {
					firstError = err
				})
				return
			}
			defer conn.Close()

			client := pb.NewTranslatorServiceClient(conn)

			for reqNum := range requestChan {
				reqStart := time.Now()

				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				resp, err := client.Trans(ctx, &pb.TransRequest{
					Text: fmt.Sprintf("%s #%d", cfg.TestText, reqNum),
					Html: cfg.TestHTML,
				})
				cancel()

				if err != nil {
					atomic.AddInt64(&failureCount, 1)
					firstErrorOnce.Do(func() {
						firstError = err
					})
					continue
				}

				latency := time.Since(reqStart)
				if resp.Code == 200 {
					atomic.AddInt64(&successCount, 1)
					latenciesMu.Lock()
					latencies = append(latencies, latency)
					latenciesMu.Unlock()
				} else {
					atomic.AddInt64(&failureCount, 1)
					firstErrorOnce.Do(func() {
						firstError = fmt.Errorf("gRPC code=%d: %s", resp.Code, resp.Message)
					})
				}
			}
		}()
	}

	for i := 0; i < cfg.NumRequests; i++ {
		requestChan <- i
	}
	close(requestChan)

	wg.Wait()
	totalDuration := time.Since(startTime)

	if silent {
		return nil
	}

	if firstError != nil && *verbose {
		fmt.Printf("  [gRPC Unix] First error: %v\n", firstError)
	}

	return calculateResult("gRPC (Unix)", cfg.NumRequests, successCount, failureCount, totalDuration, latencies)
}

func calculateResult(protocol string, totalReqs int, success, failure int64, duration time.Duration, latencies []time.Duration) *BenchmarkResult {
	if len(latencies) == 0 {
		return &BenchmarkResult{
			Protocol:      protocol,
			TotalRequests: totalReqs,
			SuccessCount:  success,
			FailureCount:  failure,
			TotalDuration: duration,
		}
	}

	sortDurations(latencies)

	var totalLatency time.Duration
	for _, lat := range latencies {
		totalLatency += lat
	}

	return &BenchmarkResult{
		Protocol:       protocol,
		TotalRequests:  totalReqs,
		SuccessCount:   success,
		FailureCount:   failure,
		TotalDuration:  duration,
		AvgLatency:     totalLatency / time.Duration(len(latencies)),
		MinLatency:     latencies[0],
		MaxLatency:     latencies[len(latencies)-1],
		P50Latency:     percentile(latencies, 0.50),
		P95Latency:     percentile(latencies, 0.95),
		P99Latency:     percentile(latencies, 0.99),
		RequestsPerSec: float64(success) / duration.Seconds(),
	}
}

func sortDurations(durations []time.Duration) {
	for i := 0; i < len(durations)-1; i++ {
		for j := i + 1; j < len(durations); j++ {
			if durations[i] > durations[j] {
				durations[i], durations[j] = durations[j], durations[i]
			}
		}
	}
}

func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	index := int(float64(len(sorted)) * p)
	if index >= len(sorted) {
		index = len(sorted) - 1
	}
	return sorted[index]
}

func printResult(result BenchmarkResult) {
	fmt.Printf("\nResults for %s:\n", result.Protocol)
	fmt.Printf("  Total Requests:    %d\n", result.TotalRequests)
	fmt.Printf("  Success:           %d\n", result.SuccessCount)
	fmt.Printf("  Failure:           %d\n", result.FailureCount)
	fmt.Printf("  Total Duration:    %v\n", result.TotalDuration)
	fmt.Printf("  Requests/sec:      %.2f\n", result.RequestsPerSec)
	fmt.Printf("  Avg Latency:       %v\n", result.AvgLatency)
	fmt.Printf("  Min Latency:       %v\n", result.MinLatency)
	fmt.Printf("  Max Latency:       %v\n", result.MaxLatency)
	fmt.Printf("  P50 Latency:       %v\n", result.P50Latency)
	fmt.Printf("  P95 Latency:       %v\n", result.P95Latency)
	fmt.Printf("  P99 Latency:       %v\n", result.P99Latency)
}

func printComparison(results []BenchmarkResult) {
	if len(results) == 0 {
		fmt.Println("No results to compare")
		return
	}

	fastest := results[0]
	for _, r := range results[1:] {
		if r.AvgLatency < fastest.AvgLatency {
			fastest = r
		}
	}

	fmt.Printf("\nFastest Protocol: %s (Avg Latency: %v, RPS: %.2f)\n\n",
		fastest.Protocol, fastest.AvgLatency, fastest.RequestsPerSec)

	fmt.Printf("%-12s | %12s | %12s | %12s | %12s | %12s\n",
		"Protocol", "Avg Latency", "P95 Latency", "P99 Latency", "RPS", "Success Rate")
	fmt.Println(strings.Repeat("-", 90))

	for _, r := range results {
		successRate := float64(r.SuccessCount) / float64(r.TotalRequests) * 100
		fmt.Printf("%-12s | %12v | %12v | %12v | %12.2f | %11.2f%%\n",
			r.Protocol, r.AvgLatency, r.P95Latency, r.P99Latency, r.RequestsPerSec, successRate)
	}

	fmt.Printf("\nLatency Comparison (relative to fastest):\n")
	for _, r := range results {
		if r.Protocol == fastest.Protocol {
			fmt.Printf("  %s: baseline\n", r.Protocol)
		} else {
			ratio := float64(r.AvgLatency) / float64(fastest.AvgLatency)
			fmt.Printf("  %s: %.2fx slower\n", r.Protocol, ratio)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func checkServerHealth(addr string) bool {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(fmt.Sprintf("http://%s/health", addr))
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
