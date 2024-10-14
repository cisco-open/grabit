// Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package internal

import (
	"context"
	"crypto/sha256"
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

// getUrl downloads the given resource and returns the path to it.
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

// GetUrlToDir downloads the given resource to the given directory and returns the path to it.
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

func (l *Resource) Contains(url string) bool {
	for _, u := range l.Urls {
		if u == url {
			return true
		}
	}
	return false
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

func (l *Resource) DownloadFile(url, targetDir string) error {
	fileName := filepath.Base(url)
	targetPath := filepath.Join(targetDir, fileName)

	if _, err := os.Stat(targetPath); err == nil {
		fileHash, err := calculateFileHash(targetPath)
		if err == nil && fileHash == l.Integrity {
			log.Debug().Str("File", fileName).Msg("File already exists with correct hash. Skipping download.")
			return nil
		}
	}

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

	return nil
}
