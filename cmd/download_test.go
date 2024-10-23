package cmd

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/cisco-open/grabit/test"
	"github.com/stretchr/testify/assert"
)

// The following function computes the SHA-256 hash of a given string content
// and returns it as a base64-encoded string prefixed with "sha256-".
// This can be useful for integrity checks of data.
func getSha256Integrity(content string) string {
	hasher := sha256.New()
	hasher.Write([]byte(content))
	return fmt.Sprintf("sha256-%s", base64.StdEncoding.EncodeToString(hasher.Sum(nil)))
}

// TestRunDownload tests the functionality of downloading files from a specified URL with integrity checks.
// It sets up a mock HTTP server, generates a temporary configuration file with resource URLs and their integrity,
// and then executes a command to download the files into a specified output directory.
// Finally, it asserts that the downloaded files contain the expected content.

func TestRunDownload(t *testing.T) {
	content := `abcdef`
	contentIntegrity := getSha256Integrity(content)
	port := test.TestHttpHandler(content, t)
	testfilepath := test.TmpFile(t, fmt.Sprintf(`
	[[Resource]]
	Urls = ['http://localhost:%d/test.html']
	Integrity = '%s'

	[[Resource]]
	Urls = ['http://localhost:%d/test3.html']
	Integrity = '%s'
`, port, contentIntegrity, port, contentIntegrity))
	outputDir := test.TmpDir(t)
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"-f", testfilepath, "download", "--dir", outputDir})
	err := cmd.Execute()
	assert.Nil(t, err)
	for _, file := range []string{"test.html", "test3.html"} {
		test.AssertFileContains(t, fmt.Sprintf("%s/%s", outputDir, file), content)
	}
}

// TestRunDownloadWithTags tests the download functionality of resources specified in a configuration file.
// It verifies that the correct files are downloaded based on the specified tags and integrity checks.

func TestRunDownloadWithTags(t *testing.T) {
	content := `abcdef`
	contentIntegrity := getSha256Integrity(content)
	port := test.TestHttpHandler(content, t)
	testfilepath := test.TmpFile(t, fmt.Sprintf(`
	[[Resource]]
	Urls = ['http://localhost:%d/test.html']
	Integrity = '%s'
	Tags = ['tag']

	[[Resource]]
	Urls = ['http://localhost:%d/test2.html']
	Integrity = '%s'
	Tags = ['tag1', 'tag2']
`, port, contentIntegrity, port, contentIntegrity))
	outputDir := test.TmpDir(t)
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"-f", testfilepath, "download", "--tag", "tag", "--dir", outputDir})
	err := cmd.Execute()
	assert.Nil(t, err)
	for _, file := range []string{"test.html"} {
		test.AssertFileContains(t, fmt.Sprintf("%s/%s", outputDir, file), content)
	}
}

// TestRunDownloadWithoutTags verifies the functionality of downloading resources
// without specific tags. It sets up a temporary HTTP server and a configuration
// file with resources, then executes the download command while excluding a
// specified tag. Finally, it asserts that the expected file is downloaded correctly.
func TestRunDownloadWithoutTags(t *testing.T) {
	content := `abcdef`
	contentIntegrity := getSha256Integrity(content)
	port := test.TestHttpHandler(content, t)
	testfilepath := test.TmpFile(t, fmt.Sprintf(`
	[[Resource]]
	Urls = ['http://localhost:%d/test.html']
	Integrity = '%s'
	Tags = ['tag']

	[[Resource]]
	Urls = ['http://localhost:%d/test2.html']
	Integrity = '%s'
	Tags = ['tag1', 'tag2']
`, port, contentIntegrity, port, contentIntegrity))
	outputDir := test.TmpDir(t)
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"-f", testfilepath, "download", "--notag", "tag", "--dir", outputDir})
	err := cmd.Execute()
	assert.Nil(t, err)
	for _, file := range []string{"test2.html"} {
		test.AssertFileContains(t, fmt.Sprintf("%s/%s", outputDir, file), content)
	}
}

// This code contains two test functions for a download command in a Go application.
// The first test, TestRunDownloadMultipleErrors, checks that the command correctly handles
// multiple download errors, including connection issues and host resolution failures.

func TestRunDownloadMultipleErrors(t *testing.T) {
	testfilepath := test.TmpFile(t, `
	[[Resource]]
	Urls = ['http://localhost:1234/test.html']
	Integrity = 'sha256-unused'

	[[Resource]]
	Urls = ['http://cannot-be-resolved.no:12/test.html']
	Integrity = 'sha256-unused'
`)
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"-f", testfilepath, "download"})
	err := cmd.Execute()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to download")
	assert.Contains(t, err.Error(), "connection refused")
	assert.Contains(t, err.Error(), "no such host")
}

// TestRunDownloadFailsIntegrityTest, verifies that the command fails
// when the downloaded content's integrity does not match the expected SHA256 hash.
// Both tests utilize temporary files and directories for testing without affecting
// the actual filesystem.
func TestRunDownloadFailsIntegrityTest(t *testing.T) {
	content := `abcdef`
	port := test.TestHttpHandler(content, t)
	testfilepath := test.TmpFile(t, fmt.Sprintf(`
	[[Resource]]
	Urls = ['http://localhost:%d/test.html']
	Integrity = 'sha256-bogus'
`, port))
	outputDir := test.TmpDir(t)
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"-f", testfilepath, "download", "--dir", outputDir})
	err := cmd.Execute()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "integrity mismatch")
}
