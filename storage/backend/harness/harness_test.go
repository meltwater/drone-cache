//go:build integration
// +build integration

package harness

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/meltwater/drone-cache/storage/common"
)

// MockClient is a mock implementation of the Client interface for testing purposes.
type MockClient struct {
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

func (m *MockClient) GetEntriesList(ctx context.Context, key string) ([]common.FileEntry, error) {
	mockEntries := []common.FileEntry{
		{Path: "file1.txt", Size: 1024, LastModified: time.Date(2024, 5, 1, 12, 0, 0, 0, time.UTC)},
		{Path: "file2.txt", Size: 2048, LastModified: time.Date(2024, 5, 2, 12, 0, 0, 0, time.UTC)},
	}

	return mockEntries, nil
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

func TestList(t *testing.T) {
	logger := log.NewNopLogger()

	mockEntries := []common.FileEntry{
		{Path: "file1.txt", Size: 1024, LastModified: time.Date(2024, 5, 1, 12, 0, 0, 0, time.UTC)},
		{Path: "file2.txt", Size: 2048, LastModified: time.Date(2024, 5, 2, 12, 0, 0, 0, time.UTC)},
	}
	mockResponseBody, _ := json.Marshal(mockEntries)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(mockResponseBody)
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	mockClient := &MockClient{
		URL: server.URL,
	}

	backend := &Backend{
		logger: logger,
		client: mockClient,
	}

	entries, err := backend.List(context.Background(), "test-prefix")

	if err != nil {
		t.Errorf("List method returned an unexpected error: %v", err)
	}

	if len(entries) != len(mockEntries) {
		t.Errorf("List method returned unexpected number of entries: got %d, want %d", len(entries), len(mockEntries))
	}
	for i, entry := range entries {
		if entry.Path != mockEntries[i].Path || entry.Size != mockEntries[i].Size || !entry.LastModified.Equal(mockEntries[i].LastModified) {
			t.Errorf("List method returned unexpected entry at index %d: got %+v, want %+v", i, entry, mockEntries[i])
		}
	}
}
