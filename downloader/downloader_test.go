package downloader

import (
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDownloader(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "grabit_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test content"))
	}))
	defer server.Close()

	hash := sha256.Sum256([]byte("test content"))
	expectedHash := hex.EncodeToString(hash[:])

	downloader := NewDownloader(5 * time.Second)

	tests := []struct {
		name          string
		url           string
		expectedHash  string
		expectedError bool
	}{
		{"Valid download", server.URL, expectedHash, false},
		{"Invalid hash", server.URL, "invalid_hash", true},
		{"Invalid URL", "http://invalid.url", expectedHash, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := downloader.DownloadFile(tt.url, tempDir, tt.expectedHash)

			if tt.expectedError && err == nil {
				t.Errorf("Expected an error, but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectedError {
				fileName := filepath.Base(tt.url)
				filePath := filepath.Join(tempDir, fileName)
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					t.Errorf("Expected file %s to exist, but it doesn't", filePath)
				}

				content, err := ioutil.ReadFile(filePath)
				if err != nil {
					t.Errorf("Failed to read file: %v", err)
				}
				if string(content) != "test content" {
					t.Errorf("File content mismatch. Expected 'test content', got '%s'", string(content))
				}
			}
		})
	}
}
