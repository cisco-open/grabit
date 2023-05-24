// Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package internal

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLockInvalid(t *testing.T) {
	_, err := NewLock("/u/d/x/invalid", false)
	assert.NotNil(t, err)
}

func TestNewLockValid(t *testing.T) {
	path := tmpFile(t, `
	[[Resource]]
	Urls = ['http://localhost:123456/test.html']
	Integrity = 'sha256-asdasdasd'
	Tags = ['tag1', 'tag2']
`)
	lock, err := NewLock(path, false)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(lock.conf.Resource))
	statement := lock.conf.Resource[0]
	assert.Equal(t, "sha256-asdasdasd", statement.Integrity)
	assert.Equal(t, []string{"tag1", "tag2"}, statement.Tags)
}

func TestLockManipulations(t *testing.T) {
	path := tmpFile(t, `
	[[Resource]]
	Urls = ['http://localhost:123456/test.html']
	Integrity = 'sha256-asdasdasd'
  `)
	lock, err := NewLock(path, false)
	assert.Nil(t, err)
	handler := func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`abcdef`))
		if err != nil {
			t.Fatal(err)
		}
	}
	port, server := httpHandler(handler)
	defer server.Close()
	resource := fmt.Sprintf("http://localhost:%d/test2.html", port)
	err = lock.AddResource([]string{resource}, "sha512", []string{}, "")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(lock.conf.Resource))
	err = lock.Save()
	assert.Nil(t, err)
	lock.DeleteResource(resource)
	assert.Equal(t, 1, len(lock.conf.Resource))
}

func TestDuplicateResource(t *testing.T) {
	url := "http://localhost:123456/test.html"
	path := tmpFile(t, fmt.Sprintf(`
		[[Resource]]
		Urls = ['%s']
		Integrity = 'sha256-asdasdasd'`, url))
	lock, err := NewLock(path, false)
	assert.Nil(t, err)
	err = lock.AddResource([]string{url}, "sha512", []string{}, "")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "already present")
}

func TestDownload(t *testing.T) {
	httpContent := []byte(`abcdef`)
	handler := func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write(httpContent)
		if err != nil {
			t.Fatal(err)
		}
	}
	port, server := httpHandler(handler)
	defer server.Close()
	path := tmpFile(t, fmt.Sprintf(`
		[[Resource]]
		Urls = ['http://localhost:%d/test.html']
		Integrity = 'sha256-vvV+x/U6bUC+tkCngKY5yDvCmsipgW8fxsXG3Nk8RyE='`, port))
	lock, err := NewLock(path, false)
	assert.Nil(t, err)
	dir := tmpDir(t)
	err = lock.Download(dir, []string{}, []string{})
	if err != nil {
		t.Fatal(err)
	}
	resFile := filepath.Join(dir, "test.html")
	assert.FileExists(t, resFile)
	content, err := os.ReadFile(resFile)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, httpContent, content)
}
