package cmd

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/cisco-open/grabit/test"
	"github.com/stretchr/testify/assert"
)

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
