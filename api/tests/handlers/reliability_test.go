package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
	"github.com/gin-gonic/gin"
)

// Endpoint definition
type Endpoint struct {
	Name   string
	Method string
	URL    string
	Body   string
}

// Test endpoints
var endpoints = []Endpoint{
	{"CreateCase", "POST", "/api/v1/cases", `{"title":"Stress Test Case"}`},
	{"GetCase", "GET", "/api/v1/cases/1", ""},
	{"UploadEvidence", "POST", "/api/v1/evidence", `{"filename":"test.log"}`},
	{"GetEvidence", "GET", "/api/v1/evidence-metadata/1", ""},
	{"SendChat", "POST", "/api/v1/chat/send", `{"message":"test"}`},
	{"GenerateReport", "POST", "/api/v1/reports", `{"title":"Stress Report"}`},
}

// Stress test result metrics
type Result struct {
	Name         string
	Total        int
	Success      int
	Failure      int
	ErrorRate    float64
	Duration     time.Duration
}

// Run stress test for a single endpoint
func runStressTest(t *testing.T, ep Endpoint, concurrentUsers, totalRequests int, server *httptest.Server) Result {
	var wg sync.WaitGroup
	requestsPerWorker := totalRequests / concurrentUsers

	var mu sync.Mutex
	var success, failure int
	start := time.Now()

	for i := 0; i < concurrentUsers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client := &http.Client{Timeout: 30 * time.Second}
			
			for j := 0; j < requestsPerWorker; j++ {
				req, err := http.NewRequest(ep.Method, server.URL+ep.URL, bytes.NewBufferString(ep.Body))
				if err != nil {
					mu.Lock()
					failure++
					mu.Unlock()
					continue
				}
				
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("userID", "550e8400-e29b-41d4-a716-446655440000")
				req.Header.Set("tenantID", "550e8400-e29b-41d4-a716-446655440001")
				req.Header.Set("teamID", "550e8400-e29b-41d4-a716-446655440002")
				
				resp, err := client.Do(req)
				mu.Lock()
				if err != nil || resp.StatusCode >= 500 {
					failure++
				} else {
					success++
				}
				mu.Unlock()
				
				if resp != nil {
					resp.Body.Close()
				}
			}
		}()
	}

	wg.Wait()
	duration := time.Since(start)
	total := success + failure
	errorRate := float64(failure) / float64(total) * 100

	return Result{
		Name:      ep.Name,
		Total:     total,
		Success:   success,
		Failure:   failure,
		ErrorRate: errorRate,
		Duration:  duration,
	}
}

// The actual test
func TestStressEndpoints(t *testing.T) {
	// Set up the test server
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Register your endpoints here
	// registerEvidenceTestEndpoints(router)
	// registerOtherEndpoints(router)
	
	server := httptest.NewServer(router)
	defer server.Close()

	concurrentUsers := 10
	totalRequests := 50

	for _, ep := range endpoints {
		t.Run(ep.Name, func(t *testing.T) {
			result := runStressTest(t, ep, concurrentUsers, totalRequests, server)
			t.Logf("[%s] Requests: %d | Errors: %d | ErrorRate: %.2f%% | Duration: %s",
				result.Name, result.Total, result.Failure, result.ErrorRate, result.Duration.String())

			// Assert low error rate
			if result.ErrorRate > 5 {
				t.Errorf("High error rate: %.2f%% for endpoint %s", result.ErrorRate, result.Name)
			}
		})
	}
}

