// Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package internal

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/cisco-open/grabit/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
			res:   Resource{Urls: []string{fmt.Sprintf("http://localhost:%d/test.html", port)}, Integrity: fmt.Sprintf("%s-vvV+x/U6bUC+tkCngKY5yDvCmsipgW8fxsXG3Nk8RyE=", algo), Tags: []string{}, Filename: "", ArtifactoryCacheURL: ""},
		},
		{
			urls:          []string{"invalid url"},
			valid:         false,
			errorContains: "failed to download",
		},
	}
	for _, data := range tests {
		resource, err := NewResourceFromUrl(data.urls, algo, []string{}, "", "")
		assert.Equal(t, data.valid, err == nil)
		if err != nil {
			assert.Contains(t, err.Error(), data.errorContains)
		} else {
			assert.Equal(t, data.res, *resource)
		}
	}
}

func TestResourceDownloadWithValidFileAlreadyPresent(t *testing.T) {
	content := `abcdef`
	contentIntegrity := test.GetSha256Integrity(content)
	port := 33 // unused because the file is already present.
	testFileName := "test.html"
	resource := Resource{Urls: []string{fmt.Sprintf("http://localhost:%d/%s", port, testFileName)}, Integrity: contentIntegrity, Tags: []string{}, Filename: ""}
	outputDir := test.TmpDir(t)
	err := os.WriteFile(filepath.Join(outputDir, testFileName), []byte(content), 0644)
	assert.Nil(t, err)
	err = resource.Download(outputDir, 0644, context.Background())
	assert.Nil(t, err)
	for _, file := range []string{testFileName} {
		test.AssertFileContains(t, fmt.Sprintf("%s/%s", outputDir, file), content)
	}
}

func TestResourceDownloadWithInValidFileAlreadyPresent(t *testing.T) {
	content := `abcdef`
	contentIntegrity := test.GetSha256Integrity(content)
	port := 33 // unused because the file, although invalid, is already present.
	testFileName := "test.html"
	resource := Resource{Urls: []string{fmt.Sprintf("http://localhost:%d/%s", port, testFileName)}, Integrity: contentIntegrity, Tags: []string{}, Filename: ""}
	outputDir := test.TmpDir(t)
	err := os.WriteFile(filepath.Join(outputDir, testFileName), []byte("invalid"), 0644)
	assert.Nil(t, err)
	err = resource.Download(outputDir, 0644, context.Background())
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "integrity mismatch")
	assert.Contains(t, err.Error(), "existing file")
}

func TestUseResourceWithCache(t *testing.T) {
	content := `abcdef`
	token := "test-token"
	port, server := test.TestHttpHandlerWithServer(content, t)
	fileName := "test.txt"
	sourceURL := fmt.Sprintf("http://localhost:%d", port)

	artServer, artPort := test.NewRecorderHttpServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			_, err := w.Write([]byte(content))
			if err != nil {
				t.Fatal(err)
			}
		}
	}, t)
	baseCacheURL := fmt.Sprintf("http://localhost:%d/", artPort)

	// Create resource, download it, upload it to cache.
	t.Setenv(GRABIT_ARTIFACTORY_TOKEN_ENV_VAR, token)
	resource, err := NewResourceFromUrl([]string{sourceURL}, "sha256", []string{}, fileName, baseCacheURL)
	require.Nil(t, err)
	server.Close() // Close origin server: file will be served from cache.
	outputDir := test.TmpDir(t)

	// Download resource from cache.
	err = resource.Download(outputDir, 0644, context.Background())
	require.Nil(t, err)
	for _, file := range []string{"test.txt"} {
		test.AssertFileContains(t, fmt.Sprintf("%s/%s", outputDir, file), content)
	}
	assert.Equal(t, 2, len(*artServer.Requests))
	assert.Equal(t, "PUT", (*artServer.Requests)[0].Method)
	assert.Equal(t, []byte(content), (*artServer.Requests)[0].Body)
	assert.Equal(t, []string([]string{fmt.Sprintf("Bearer %s", token)}), (*artServer.Requests)[0].Headers["Authorization"])
	assert.Equal(t, "GET", (*artServer.Requests)[1].Method)

	// Delete resource, deleting it from cache.
	err = resource.Delete()
	require.Nil(t, err)
	assert.Equal(t, 3, len(*artServer.Requests))
	assert.Equal(t, "DELETE", (*artServer.Requests)[2].Method)
	assert.Equal(t, []string([]string{fmt.Sprintf("Bearer %s", token)}), (*artServer.Requests)[2].Headers["Authorization"])
}

func TestResourceWithCacheCorruptedCache(t *testing.T) {
	content := `abcdef`
	token := "test-token"
	port, server := test.TestHttpHandlerWithServer(content, t)
	fileName := "test.txt"
	sourceURL := fmt.Sprintf("http://localhost:%d", port)

	_, artPort := test.NewRecorderHttpServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			_, err := w.Write([]byte("invalid-content"))
			if err != nil {
				t.Fatal(err)
			}
		}
	}, t)
	baseCacheURL := fmt.Sprintf("http://localhost:%d/", artPort)

	// Create resource, download it, upload it to cache.
	t.Setenv(GRABIT_ARTIFACTORY_TOKEN_ENV_VAR, token)
	resource, err := NewResourceFromUrl([]string{sourceURL}, "sha256", []string{}, fileName, baseCacheURL)
	require.Nil(t, err)
	server.Close() // Close origin server: file will be served from cache.
	outputDir := test.TmpDir(t)

	// Download resource from cache.
	err = resource.Download(outputDir, 0644, context.Background())
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "cache file at")
	assert.Contains(t, err.Error(), "with incorrect integrity")
}

func TestResourceWithCacheNoToken(t *testing.T) {
	t.Setenv(GRABIT_ARTIFACTORY_TOKEN_ENV_VAR, "")
	fileName := "test.txt"
	port := 33
	sourceURL := fmt.Sprintf("http://localhost:%d", port)
	_, err := NewResourceFromUrl([]string{sourceURL}, "sha256", []string{}, fileName, "http://localhost:8080/")
	assert.NotNil(t, err)
}
