package cmd

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/cisco-open/grabit/internal"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/cisco-open/grabit/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getSha256Integrity(content string) string {
	hasher := sha256.New()
	hasher.Write([]byte(content))
	return fmt.Sprintf("sha256-%s", base64.StdEncoding.EncodeToString(hasher.Sum(nil)))
}

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

func TestOptimization(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("test content"))
		if err != nil {
			return
		}
	}))
	defer ts.Close()

	t.Run("Valid_File_Not_Redownloaded", func(t *testing.T) {
		tmpDir := test.TmpDir(t)
		testUrl := ts.URL + "/valid_test.txt"

		lockPath := test.TmpFile(t, "")
		lock, err := internal.NewLock(lockPath, true)
		require.NoError(t, err)

		err = lock.AddResource([]string{testUrl}, internal.RecommendedAlgo, nil, "valid_test.txt")
		require.NoError(t, err)

		// Update the Download call to match the new signature
		err = lock.Download(tmpDir, nil, nil, "", false) // Added 'false' for the new boolean argument
		require.NoError(t, err)
	})

	t.Run("Invalid_File_Redownloaded", func(t *testing.T) {
		tmpDir := test.TmpDir(t)
		testUrl := ts.URL + "/invalid_test.txt"

		lockPath := test.TmpFile(t, "")
		lock, err := internal.NewLock(lockPath, true)
		require.NoError(t, err)

		err = lock.AddResource([]string{testUrl}, internal.RecommendedAlgo, nil, "invalid_test.txt")
		require.NoError(t, err)

		err = lock.Save()
		require.NoError(t, err)

		invalidPath := filepath.Join(tmpDir, "invalid_test.txt")
		err = os.WriteFile(invalidPath, []byte("corrupted"), 0644)
		require.NoError(t, err)

		// Update the Download call to match the new signature
		err = lock.Download(tmpDir, nil, nil, "", false) // Added 'false' for the new boolean argument
		require.Error(t, err)
		assert.Contains(t, err.Error(), "integrity mismatch")
	})
}

func TestRunDownloadTriesAllUrls(t *testing.T) {
	content := `abcdef`
	contentIntegrity := getSha256Integrity(content)
	port := test.TestHttpHandler(content, t)
	testfilepath := test.TmpFile(t, fmt.Sprintf(`
	[[Resource]]
	Urls = ['http://cannot-be-resolved.no:12/test.html', 'http://localhost:%d/test.html']
	Integrity = '%s'
`, port, contentIntegrity))
	outputDir := test.TmpDir(t)
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"-f", testfilepath, "download", "--dir", outputDir})
	err := cmd.Execute()
	assert.Nil(t, err)
	for _, file := range []string{"test.html"} {
		test.AssertFileContains(t, fmt.Sprintf("%s/%s", outputDir, file), content)
	}
}
