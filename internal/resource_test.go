// Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package internal

import (
	"context"
	"fmt"
	"github.com/cisco-open/grabit/test"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

//This is a test function for creating a new Resource from a given URL.
//It sets up a mock HTTP server to respond with a specific byte sequence.
//The test cases check the validity of URLs, ensuring that valid URLs return the expected Resource and invalid URLs return appropriate error messages.

func TestNewResourceFromUrl(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`abcdef`))
		if err != nil {
			t.Fatal(err)
		}
	}
	port, server := test.HttpHandler(handler)
	defer server.Close()
	algo := "sha256"
	tests := []struct {
		urls          []string
		valid         bool
		errorContains string
		res           Resource
	}{
		{
			urls:  []string{fmt.Sprintf("http://localhost:%d/test.html", port)},
			valid: true,
			res:   Resource{Urls: []string{fmt.Sprintf("http://localhost:%d/test.html", port)}, Integrity: fmt.Sprintf("%s-vvV+x/U6bUC+tkCngKY5yDvCmsipgW8fxsXG3Nk8RyE=", algo), Tags: []string{}, Filename: ""},
		},
		{
			urls:          []string{"invalid url"},
			valid:         false,
			errorContains: "failed to download",
		},
	}

	for _, data := range tests {
		resource, err := NewResourceFromUrl(data.urls, "sha256", []string{}, "", false)
		assert.Equal(t, data.valid, err == nil)
		if err != nil {
			assert.Contains(t, err.Error(), data.errorContains)
		} else {
			assert.Equal(t, data.res, *resource)
		}
	}
}

// TestDynamicResourceDownload tests the downloading of a dynamic resource.
// It sets up a temporary HTTP server that responds with the current time,
// then attempts to download the resource twice to ensure that the download
// functionality works correctly even with changing content.

func TestDynamicResourceDownload(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(time.Now().String()))
	}
	port, server := test.HttpHandler(handler)
	defer server.Close()

	url := fmt.Sprintf("http://localhost:%d/dynamic", port)
	resource := &Resource{
		Urls:    []string{url},
		Dynamic: true,
	}

	dir := t.TempDir()
	err := resource.Download(dir, 0644, context.Background())
	assert.NoError(t, err)

	// Download again to ensure it doesn't fail due to content change
	err = resource.Download(dir, 0644, context.Background())
	assert.NoError(t, err)
}
