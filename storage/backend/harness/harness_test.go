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

func (m *MockClient) GetListURL(ctx context.Context, key, continuationToken string) (string, error) {
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

func TestList(t *testing.T) {
	logger := log.NewNopLogger()

	// Mock XML response
	xmlResponse := `
	<ListBucketResult>
		<Contents>
			<Key>file1.txt</Key>
			<LastModified>2024-05-01T12:00:00Z</LastModified>
			<Size>1024</Size>
		</Contents>
		<Contents>
			<Key>file2.txt</Key>
			<LastModified>2024-05-02T12:00:00Z</LastModified>
			<Size>2048</Size>
		</Contents>
		<IsTruncated>false</IsTruncated>
	</ListBucketResult>
	`

	// Create a mock HTTP server
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, xmlResponse)
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	// Create a mock client
	mockClient := &MockClient{
		URL: server.URL,
	}

	backend := &Backend{
		logger: logger,
		client: mockClient,
	}

	// Execute List method
	entries, err := backend.List(context.Background(), "test-key")

	// Check for errors
	if err != nil {
		t.Errorf("List method returned an unexpected error: %v", err)
	}

	// Check the number of entries
	if len(entries) != 2 {
		t.Errorf("List method returned unexpected number of entries: got %d, want %d", len(entries), 2)
	}

	// Check the content of the entries
	expectedEntries := []common.FileEntry{
		{Path: "file1.txt", Size: 1024, LastModified: time.Date(2024, 5, 1, 12, 0, 0, 0, time.UTC)},
		{Path: "file2.txt", Size: 2048, LastModified: time.Date(2024, 5, 2, 12, 0, 0, 0, time.UTC)},
	}
	for i, entry := range entries {
		if entry != expectedEntries[i] {
			t.Errorf("List method returned unexpected entry at index %d: got %+v, want %+v", i, entry, expectedEntries[i])
		}
	}
}

