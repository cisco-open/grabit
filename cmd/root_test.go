package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunRoot(t *testing.T) {
	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOutput(buf)
	Execute(rootCmd)
	assert.Contains(t, buf.String(), "and verifies their integrity")
}
