// Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package internal

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/rs/zerolog/log"

	"github.com/carlmjohnson/requests"
)

// Resource represents an external resource to be downloaded.
type Resource struct {
	Urls      []string
	Integrity string
	Tags      []string `toml:",omitempty"`
	Filename  string   `toml:",omitempty"`
	Dynamic   bool     `toml:",omitempty"`
}

// this function creates a new resource object from a list of URls, a hash algorithm, tags, a filename and a dynamic flag
// it fetches the first URl and downloads it to a temporary file, calculates its integrity hash and constructs a resource object
func NewResourceFromUrl(urls []string, algo string, tags []string, filename string, dynamic bool) (*Resource, error) {
	if len(urls) < 1 {
		return nil, fmt.Errorf("empty url list")
	}

	url := urls[0]
	ctx := context.Background()
	path, err := GetUrltoTempFile(url, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get url: %s", err)
	}
	defer os.Remove(path)
	integrity, err := getIntegrityFromFile(path, algo)
	if err != nil {
		return nil, fmt.Errorf("failed to compute ressource integrity: %s", err)
	}
	return &Resource{Urls: urls, Integrity: integrity, Tags: tags, Filename: filename}, nil
}

// getUrl downloads the given resource from URL and saves it to the specified filename returning to the path.
// using the requests package to handle HTTP GET request and logs the download process
func getUrl(u string, fileName string, ctx context.Context) (string, error) {
	_, err := url.Parse(u)
	if err != nil {
		return "", fmt.Errorf("invalid url '%s': %s", u, err)
	}
	log.Debug().Str("URL", u).Msg("Downloading")
	err = requests.
		URL(u).
		Header("Accept", "*/*").
		ToFile(fileName).
		Fetch(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to download '%s': %s", u, err)
	}
	log.Debug().Str("URL", u).Msg("Downloaded")
	return fileName, nil
}

// this function checks the integrity of a resource downloaded from a given URL against an expected integrity hash
// it downloads the resource to a temporary file, calculates its integrity hash and compares it to the expected integrity value or hash
func checkIntegrityFromUrl(url string, expectedIntegrity string) error {
	tempFile, err := GetUrltoTempFile(url, context.Background())
	if err != nil {
		return err
	}
	defer os.Remove(tempFile)

	algo, err := getAlgoFromIntegrity(expectedIntegrity)
	if err != nil {
		return err
	}

	return checkIntegrityFromFile(tempFile, algo, expectedIntegrity, url)
}

// GetUrlToDir downloads the given resource to the given directory and returns the path to it. generating a temporary filename based on the SHA-256 hash of the URL and calls
func GetUrlToDir(u string, targetDir string, ctx context.Context) (string, error) {
	// create temporary name in the target directory.
	h := sha256.New()
	h.Write([]byte(u))
	fileName := filepath.Join(targetDir, fmt.Sprintf(".%s", hex.EncodeToString(h.Sum(nil))))
	return getUrl(u, fileName, ctx)
}

// GetUrlWithDir downloads the given resource to a temporary file and returns the path to it.
func GetUrltoTempFile(u string, ctx context.Context) (string, error) {
	file, err := os.CreateTemp("", "prefix")
	if err != nil {
		log.Fatal().Err(err)
	}
	fileName := file.Name()
	return getUrl(u, fileName, ctx)
}

// this function downloads the resource from each URL in the URLs list,
// if a download is successful, it sets the file's permission and returns, if all downloads fail, it returns an error.
func (l *Resource) Download(dir string, mode os.FileMode, ctx context.Context) error {
	for _, u := range l.Urls {
		err := l.DownloadFile(u, dir)
		if err != nil {
			continue
		}

		localName := l.Filename
		if localName == "" {
			localName = path.Base(u)
		}
		resPath := filepath.Join(dir, localName)

		if mode != NoFileMode {
			err = os.Chmod(resPath, mode.Perm())
			if err != nil {
				return err
			}
		}
		return nil
	}
	return fmt.Errorf("failed to download resource from any URL")
}

// this method checks if a given URL is present in the URL list of the resource
func (l *Resource) Contains(url string) bool {
	for _, u := range l.Urls {
		if u == url {
			return true
		}
	}
	return false
}

// this function calculates the SHA-256 hash of a file specified by its path and returns the hash as a hexadecimal string
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

	return base64.StdEncoding.EncodeToString(hash.Sum(nil)), nil
}

// this method downloads a file from a specified URL to a target directory,
// it checks if a file with the same name already exists and verifies its integrity using the provided hash
// if the file does not exist or the hash does not match it downloads the file, calculates its hash and checks for integrity
func (l *Resource) DownloadFile(url, targetDir string) error {
	fileName := filepath.Base(url)
	targetPath := filepath.Join(targetDir, fileName)
	duplicateCount := 0

	// Check for existing file and hash before download
	if _, err := os.Stat(targetPath); err == nil {
		fileHash, err := calculateFileHash(targetPath)
		if err == nil && fileHash == l.Integrity {
			duplicateCount++
			log.Info().Str("File", fileName).Msg("Existing file verified with correct hash")
			return nil
		}
	}

	// Download and verify new file
	resp, err := http.Get(url)
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

	downloadedHash, err := calculateFileHash(targetPath)
	if err != nil {
		return err
	}
	if downloadedHash != l.Integrity {
		return fmt.Errorf("hash mismatch: expected %s, got %s", l.Integrity, downloadedHash)
	}

	if duplicateCount > 0 {
		log.Info().Int("duplicates", duplicateCount).Msg("Duplicate files found during download")
	}

	return nil
}
