package internal

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/cisco-open/grabit/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAddWithArtifactoryCache verifies adding a resource with caching enabled.
func TestAddWithArtifactoryCacheTokenNotSet(t *testing.T) {
	// Setup a simple HTTP handler that always returns "test content".
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`test content`))
	}
	// Start the HTTP server and get the port it runs on.
	port, server := test.HttpHandler(handler)
	defer server.Close() // Ensure the server is stopped after the test.

	// Create a temporary lock file for testing.
	path := test.TmpFile(t, "")
	lock, err := NewLock(path, true)
	require.NoError(t, err) // Fail the test if the lock cannot be created.

	// Set up URLs for the source file and cache.
	sourceURL := fmt.Sprintf("http://localhost:%d/test.txt", port)
	cacheURL := fmt.Sprintf("http://localhost:%d", port)

	// Attempt to add a resource to the lock file.
	err = lock.AddResource([]string{sourceURL}, "sha256", []string{}, "", cacheURL)
	// Verify that the error message indicates the token is not set.
	assert.Contains(t, err.Error(), "GRABIT_ARTIFACTORY_TOKEN environment variable is not set")
}

// TestDeleteWithArtifactoryCache verifies deleting a resource with caching enabled.
func TestDeleteWithArtifactoryCache(t *testing.T) {
	// Set the GRABIT_ARTIFACTORY_TOKEN environment variable.
	t.Setenv("GRABIT_ARTIFACTORY_TOKEN", "test-token")

	// Setup an HTTP handler to handle DELETE requests.
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" { // Respond with OK for DELETE requests.
			w.WriteHeader(http.StatusOK)
		}
	}
	// Start the HTTP server and get the port it runs on.
	port, server := test.HttpHandler(handler)
	defer server.Close() // Ensure the server is stopped after the test.

	// Set up URLs for the source file and cache.
	sourceURL := fmt.Sprintf("http://localhost:%d/test.txt", port)
	cacheURL := fmt.Sprintf("http://localhost:%d", port)

	// Create a lock file with the resource and cache information.
	lockContent := fmt.Sprintf(`[[Resource]]
        Urls = ['%s']
        Integrity = 'sha256-test'
        CacheUri = '%s'`, sourceURL, cacheURL)

	lockPath := test.TmpFile(t, lockContent)
	lock, err := NewLock(lockPath, false)
	require.NoError(t, err) // Fail the test if the lock cannot be created.

	// Save the lock file before modifying it.
	err = lock.Save()
	require.NoError(t, err) // Fail the test if saving fails.

	// Delete the resource from the lock file.
	lock.DeleteResource(sourceURL)

	// Checks that http.StatusOK was returned meaning the delete was successful
	resp, err := http.Get(server.URL)
	if err != nil {
		require.NoError(t, err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 but got %d", resp.StatusCode)
	}

	// Save the lock file again after deletion.
	err = lock.Save()
	require.NoError(t, err) // Fail the test if saving fails.

	// Reload the lock file and verify the resource is gone.
	newLock, err := NewLock(lockPath, false)
	require.NoError(t, err)
	assert.Equal(t, 0, len(newLock.conf.Resource)) // Ensure no resources remain.
}
