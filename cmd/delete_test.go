package cmd

import (
	"testing"

	"github.com/cisco-open/grabit/test"
	"github.com/stretchr/testify/assert"
)

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
