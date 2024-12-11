// Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package test

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func GetSha256Integrity(content string) string {
	hasher := sha256.New()
	hasher.Write([]byte(content))
	return fmt.Sprintf("sha256-%s", base64.StdEncoding.EncodeToString(hasher.Sum(nil)))
}

func TmpFile(t *testing.T, content string) string {
	f, err := os.CreateTemp(t.TempDir(), "test")
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.WriteString(content)
	if err != nil {
		t.Fatal(err)
	}
	name := f.Name()
	t.Cleanup(func() { os.RemoveAll(name) })
	return name
}

func TmpDir(t *testing.T) string {
	dir, err := os.MkdirTemp(t.TempDir(), "test")
	if err != nil {
		log.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func HttpHandler(handler http.HandlerFunc) (int, *httptest.Server) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	s := httptest.NewUnstartedServer(http.HandlerFunc(handler))
	s.Listener.Close()
	s.Listener = l
	s.Start()
	return l.Addr().(*net.TCPAddr).Port, s
}

type RecordedRequest struct {
	Method  string
	Url     string
	Body    []byte
	Headers map[string][]string
}

func NewRecordedRequest(r *http.Request) *RecordedRequest {
	var bodyBytes []byte
	if r.Body != nil {
		bodyBytes, _ = io.ReadAll(r.Body)
	}
	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	return &RecordedRequest{
		Method:  r.Method,
		Url:     r.URL.String(),
		Body:    bodyBytes,
		Headers: r.Header}
}

type RecorderHttpServer struct {
	*httptest.Server
	Requests *[]RecordedRequest
}

func NewRecorderHttpServer(handler http.HandlerFunc, t *testing.T) (*RecorderHttpServer, int) {
	requests := make([]RecordedRequest, 0)

	outerHandler := func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, *NewRecordedRequest(r))
		handler(w, r)
	}
	port, server := HttpHandler(outerHandler)
	t.Cleanup(func() { server.Close() })
	return &RecorderHttpServer{Server: server, Requests: &requests}, port
}

// TestHttpHandler creates a new HTTP server and returns the port and serves
// the given content. Its lifetime is tied to the given testing.T object.
func TestHttpHandler(content string, t *testing.T) int {
	port, _ := TestHttpHandlerWithServer(content, t)
	return port
}

func TestHttpHandlerWithServer(content string, t *testing.T) (int, *httptest.Server) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(content))
		if err != nil {
			t.Fatal(err)
		}
	}
	port, server := HttpHandler(handler)
	t.Cleanup(func() { server.Close() })
	return port, server
}

// AssertFileContains asserts that the file at the given path exists and
// contains the given content.
func AssertFileContains(t *testing.T, path, content string) {
	assert.FileExists(t, path)
	fileContent, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, content, string(fileContent))
}
