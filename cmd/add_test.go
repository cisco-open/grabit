package cmd

import (
	"fmt"
	"testing"

	"github.com/cisco-open/grabit/internal"
	"github.com/cisco-open/grabit/test"
	"github.com/stretchr/testify/assert"
)

func TestRunAdd(t *testing.T) {
	port := test.TestHttpHandler("abcdef", t)
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"-f", test.TmpFile(t, ""), "add", fmt.Sprintf("http://localhost:%d/test.html", port)})
	err := cmd.Execute()
	assert.Nil(t, err)
}

func TestRunAddWithArtifactoryCache(t *testing.T) {
	t.Setenv(internal.GRABIT_ARTIFACTORY_TOKEN_ENV_VAR, "artifactory-token")
	port := test.TestHttpHandler("abcdef", t)
	artPort := test.TestHttpHandler("abcdef", t)
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"-f", test.TmpFile(t, ""), "add", fmt.Sprintf("http://localhost:%d/test.html", port), "--artifactory-cache-url", fmt.Sprintf("http://localhost:%d/artifactory", artPort)})
	err := cmd.Execute()
	assert.Nil(t, err)
}
