package cmd

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestRunRoot tests the execution of the root command to ensure it outputs the expected string related to integrity verification.

func TestRunRoot(t *testing.T) {
	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOutput(buf)
	Execute(rootCmd)
	assert.Contains(t, buf.String(), "and verifies their integrity")
}
