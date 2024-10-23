package cmd

import (
	"testing"

	"github.com/cisco-open/grabit/test"
	"github.com/stretchr/testify/assert"
)

// TestRunDelete tests the delete command of the CLI application.
// It creates a temporary file with a resource configuration, sets up the command with the appropriate arguments, and executes it.
// The test asserts that no error occurred during the execution of the command.
func TestRunDelete(t *testing.T) {
	testfilepath := test.TmpFile(t, `
	[[Resource]]
	Urls = ['http://localhost:123456/test.html']
	Integrity = 'sha256-asdasdasd'
	Tags = ['tag1', 'tag2']
`)
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"-f", testfilepath, "delete", "http://localhost:123456/test.html"})
	err := cmd.Execute()
	assert.Nil(t, err)
}
