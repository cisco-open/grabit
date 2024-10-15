package cmd

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunRoot(t *testing.T) {
	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOutput(buf)
	d := downloader.NewDownloader(10 * time.Second) // Create a new downloader with a 10-second timeout
	Execute(rootCmd)
	assert.Contains(t, buf.String(), "and verifies their integrity")
}
