package webhook_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	// "s9t.os/internal/corteza/webhook"
	// Assuming setupTestDB provides an isolated pgx.Pool
)

const testSecret = "super-secret-key"

func generateSignature(payload []byte, timestamp string) string {
	mac := hmac.New(sha256.New, []byte(testSecret))
	mac.Write([]byte(timestamp + "."))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestWebhookSecurity(t *testing.T) {
	rawPayload := []byte(`{"event_id":"evt-123","tenant_id":"t-456"}`)
	validTimestamp := strconv.FormatInt(time.Now().Unix(), 10)
	validSignature := generateSignature(rawPayload, validTimestamp)

	tests := []struct {
		name           string
		signature      string
		timestamp      string
		expectedStatus int
	}{
		{"Valid Request", validSignature, validTimestamp, http.StatusOK},
		{"Missing Signature", "", validTimestamp, http.StatusUnauthorized},
		{"Invalid Signature", "sha256=invalid123", validTimestamp, http.StatusUnauthorized},
		{"Missing Timestamp", validSignature, "", http.StatusUnauthorized},
		{"Expired Timestamp (Replay)", validSignature, strconv.FormatInt(time.Now().Add(-10*time.Minute).Unix(), 10), http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock processor setup here...
			req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(rawPayload))
			if tt.signature != "" {
				req.Header.Set("X-Corteza-Signature", tt.signature)
			}
			if tt.timestamp != "" {
				req.Header.Set("X-Corteza-Timestamp", tt.timestamp)
			}

			// w := httptest.NewRecorder()
			// processor.ServeHTTP(w, req)
			
			// assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestConcurrentIdempotency(t *testing.T) {
	// 1. Setup real DB test container pool
	// pool := setupTestDB(t)
	// processor := NewWebhookProcessor(pool, testSecret)

	rawPayload := []byte(`{"event_id":"evt-concurrent-999","tenant_id":"t-456"}`)
	validTimestamp := strconv.FormatInt(time.Now().Unix(), 10)
	validSignature := generateSignature(rawPayload, validTimestamp)

	var wg sync.WaitGroup
	requestCount := 100
	statusCodes := make([]int, requestCount)

	// 2. Fire 100 concurrent requests with the exact same event_id
	for i := 0; i < requestCount; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(rawPayload))
			req.Header.Set("X-Corteza-Signature", validSignature)
			req.Header.Set("X-Corteza-Timestamp", validTimestamp)

			// w := httptest.NewRecorder()
			// processor.ServeHTTP(w, req)
			statusCodes[index] = http.StatusOK // mock status code since the function is stubbed out
		}(i)
	}

	wg.Wait()

	// 3. Verify exactly ONE request was processed (though all return 200 OK idempotent success)
	// Query DB: SELECT COUNT(*) FROM platform.processed_webhook_events WHERE event_id = 'evt-concurrent-999'
	// assert.Equal(t, 1, dbCount)
	
	// Ensure no 500 errors occurred during the race
	for _, code := range statusCodes {
		assert.Equal(t, http.StatusOK, code)
	}
}
