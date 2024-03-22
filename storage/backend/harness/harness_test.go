//go:build integration
// +build integration

package harness

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-kit/kit/log"
)

// MockClient is a mock implementation of the Client interface for testing purposes.
type MockClient struct{
    URL string
}

func (m *MockClient) GetUploadURL(ctx context.Context, key string) (string, error) {
	return m.URL, nil
}

func (m *MockClient) GetDownloadURL(ctx context.Context, key string) (string, error) {
	return m.URL, nil
}

func (m *MockClient) GetExistsURL(ctx context.Context, key string) (string, error) {
	return m.URL, nil
}

func TestGet(t *testing.T) {
	logger := log.NewNopLogger()
    // Create a mock HTTP server
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "test data")
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	backend := &Backend{
		logger: logger,
		client: &MockClient{
            URL: server.URL,
        },
	}
	// Execute Get method
	var buf bytes.Buffer
	err := backend.Get(context.Background(), "test-key", &buf)

	// Check for errors
	if err != nil {
		t.Errorf("Get method returned an unexpected error: %v", err)
	}

	// Check the content of the buffer
	expected := "test data"
	if buf.String() != expected {
		t.Errorf("Get method returned unexpected data: got %s, want %s", buf.String(), expected)
	}
}

func TestPut(t *testing.T) {
	logger := log.NewNopLogger()
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	backend := &Backend{
		logger: logger,
		client: &MockClient{
            URL: server.URL,
        },
	}

	// Execute Put method
	err := backend.Put(context.Background(), "test-key", bytes.NewBuffer([]byte("test data")))

	// Check for errors
	if err != nil {
		t.Errorf("Put method returned an unexpected error: %v", err)
	}
}

func TestExists(t *testing.T) {
	logger := log.NewNopLogger()
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Add("ETag", "test")
		w.WriteHeader(http.StatusOK)
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	backend := &Backend{
		logger: logger,
		client: &MockClient{
            URL: server.URL,
        },
	}

	// Execute Exists method
	exists, err := backend.Exists(context.Background(), "test-key")

	// Check for errors
	if err != nil {
		t.Errorf("Exists method returned an unexpected error: %v", err)
	}

	// Check the existence flag
	if !exists {
		t.Error("Exists method returned false, expected true")
	}
}

func TestNotExists(t *testing.T) {
	logger := log.NewNopLogger()
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	backend := &Backend{
		logger: logger,
		client: &MockClient{
            URL: server.URL,
        },
	}

	// Execute Exists method
	exists, err := backend.Exists(context.Background(), "test-key")

	// Check for errors
	if err != nil {
		t.Errorf("Exists method returned an unexpected error: %v", err)
	}

	// Check the existence flag
	if exists {
		t.Error("Exists method returned true, expected false")
	}
}

func TestNotExistsWithout404(t *testing.T) {
	logger := log.NewNopLogger()
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	backend := &Backend{
		logger: logger,
		client: &MockClient{
            URL: server.URL,
        },
	}

	// Execute Exists method
	exists, err := backend.Exists(context.Background(), "test-key")

	// Check for errors
	if err == nil {
		t.Error("Exists method did not return error")
	}

	// Check the existence flag
	if exists {
		t.Error("Exists method returned true, expected false")
	}
}