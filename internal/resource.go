// Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package internal

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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
}

func NewResourceFromUrl(urls []string, algo string, tags []string, filename string) (*Resource, error) {
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
	ok := false
	algo, err := getAlgoFromIntegrity(l.Integrity)
	if err != nil {
		return err
	}
	var downloadError error = nil
	for _, u := range l.Urls {
		// Download file in the target directory so that the call to
		// os.Rename is atomic.
		log.Debug().Str("URL", u).Msg("Downloading")

		localName := ""
		if l.Filename != "" {
			localName = l.Filename
		} else {
			localName = path.Base(u)
		}
		resPath := filepath.Join(dir, localName)

		// Check existing file first
		if _, err := os.Stat(resPath); err == nil {
			// File exists, validate its integrity
			if !ValidateLocalFile(resPath, l.Integrity) {
				return fmt.Errorf("integrity mismatch for '%s'", resPath)
			}
			// Set file permissions if needed
			if mode != NoFileMode {
				if err := os.Chmod(resPath, mode.Perm()); err != nil {
					return err
				}
			}
			ok = true
			continue
		} else if !os.IsNotExist(err) {
			// Handle other potential errors from os.Stat
			return fmt.Errorf("failed to stat file '%s': %v", resPath, err)
		}

		// Download new file
		lpath, err := GetUrlToDir(u, dir, ctx)
		if err != nil {
			downloadError = fmt.Errorf("failed to download '%s': %v", u, err)
			continue
		}

		// Validate downloaded file
		if err := checkIntegrityFromFile(lpath, algo, l.Integrity, u); err != nil {
			os.Remove(lpath)
			downloadError = err
			continue
		}

		// Move to final location
		if err := os.Rename(lpath, resPath); err != nil {
			os.Remove(lpath)
			downloadError = err
			continue
		}

		if mode != NoFileMode {
			if err := os.Chmod(resPath, mode.Perm()); err != nil {
				return err
			}
		}
		ok = true
		break
	}

	if !ok {
		return downloadError
	}
	return nil
}
func (l *Resource) Contains(url string) bool {
	for _, u := range l.Urls {
		if u == url {
			return true
		}
	}
	return false
}

func ValidateLocalFile(filePath string, expectedIntegrity string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}

	algo, err := getAlgoFromIntegrity(expectedIntegrity)
	if err != nil {
		return false
	}

	fileIntegrity, err := getIntegrityFromFile(filePath, algo)
	if err != nil {
		return false
	}

	return fileIntegrity == expectedIntegrity
}
