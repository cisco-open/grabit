// this implementation includes a feature to skip unnecessary downloads if the file already exists.
// DownloadFile function checks for existing files and verifies their hash, and only downloads if necessary.
// This also includes error handling for hash mismatches ensuring data integrity improving efficiency
// and avoiding redundant downloads while maintaining file accuracy.

package downloader

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Downloader struct {
	Client *http.Client
}

func NewDownloader(timeout time.Duration) *Downloader {
	return &Downloader{
		Client: &http.Client{Timeout: timeout},
	}
}

func (d *Downloader) DownloadFile(url, targetDir, expectedHash string) error {
	fileName := filepath.Base(url)
	targetPath := filepath.Join(targetDir, fileName)

	if _, err := os.Stat(targetPath); err == nil {
		fileHash, err := calculateFileHash(targetPath)
		if err != nil {
			return err
		}

		if fileHash == expectedHash {
			fmt.Printf("File '%s' already exists and matches the expected hash. Skipping download.\n", fileName)
			return nil
		}
		fmt.Printf("File '%s' exists but hash mismatch. Downloading again.\n", fileName)
	}

	resp, err := d.Client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	// Verify the downloaded file's hash
	downloadedHash, err := calculateFileHash(targetPath)
	if err != nil {
		return err
	}
	if downloadedHash != expectedHash {
		return fmt.Errorf("hash mismatch: expected %s, got %s", expectedHash, downloadedHash)
	}

	fmt.Printf("Downloaded '%s' to '%s'.\n", fileName, targetPath)
	return nil
}

func calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
