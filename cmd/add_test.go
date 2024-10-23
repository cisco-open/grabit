package cmd

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/cisco-open/grabit/test"
	"github.com/stretchr/testify/assert"
)

// TestRunAdd tests the functionality of adding a resource by simulating an HTTP server response.
// It sets up a handler that responds with "abcdef", starts the server, and executes a command
// to add a resource from the specified URL, asserting that no error occurs during execution.

func TestRunAdd(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`abcdef`))
		if err != nil {
			t.Fatal(err)
		}
	}
	port, server := test.HttpHandler(handler)
	defer server.Close()
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"-f", test.TmpFile(t, ""), "add", fmt.Sprintf("http://localhost:%d/test.html", port)})
	err := cmd.Execute()
	assert.Nil(t, err)
}
