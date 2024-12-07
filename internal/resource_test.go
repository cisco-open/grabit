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
			res:   Resource{Urls: []string{fmt.Sprintf("http://localhost:%d/test.html", port)}, Integrity: fmt.Sprintf("%s-vvV+x/U6bUC+tkCngKY5yDvCmsipgW8fxsXG3Nk8RyE=", algo), Tags: []string{}, Filename: ""},
		},
		{
			urls:          []string{"invalid url"},
			valid:         false,
			errorContains: "failed to download",
		},
	}

	for _, data := range tests {
		resource, err := NewResourceFromUrl(data.urls, algo, []string{}, "")
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
