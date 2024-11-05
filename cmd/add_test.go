package cmd

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/cisco-open/grabit/test"
	"github.com/stretchr/testify/assert"
)

func TestRunAdd(t *testing.T) {
	// Set the GRABIT_ARTIFACTORY_TOKEN environment variable.
	t.Setenv("GRABIT_ARTIFACTORY_TOKEN", "test-token")

	// Setup HTTP handler for the resource
	handler := func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`abcdef`))
		if err != nil {
			t.Fatal(err)
		}
	}
	port, server := test.HttpHandler(handler)
	defer server.Close()

	// Setup dummy cache server
	cacheHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" {
			w.WriteHeader(http.StatusCreated)
		}
	}
	cachePort, cacheServer := test.HttpHandler(cacheHandler)
	defer cacheServer.Close()

	// Create empty lockfile
	lockFile := test.TmpFile(t, "")

	cmd := NewRootCmd()
	// Add cache URL to the command
	cacheURL := fmt.Sprintf("http://localhost:%d", cachePort)
	cmd.SetArgs([]string{
		"-f", lockFile,
		"add",
		fmt.Sprintf("http://localhost:%d/test.html", port),
		"--cache", cacheURL,
	})

	err := cmd.Execute()
	assert.Nil(t, err)
}
