package cmd

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRunRoot(t *testing.T) {
	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOutput(buf)
	Execute(rootCmd)
	assert.Contains(t, buf.String(), "and verifies their integrity")
}
