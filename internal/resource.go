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
	CacheUri  string   `toml:",omitempty"`
}

func NewResourceFromUrl(urls []string, algo string, tags []string, filename string, cacheURL string) (*Resource, error) {
	if len(urls) < 1 {
		return nil, fmt.Errorf("empty url list")
	}
	url := urls[0]
	ctx := context.Background()
	path, err := GetUrltoTempFile(url, "", ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get url: %s", err)
	}
	defer os.Remove(path)
	integrity, err := getIntegrityFromFile(path, algo)
	if err != nil {
		return nil, fmt.Errorf("failed to compute ressource integrity: %s", err)
	}

	return &Resource{Urls: urls, Integrity: integrity, Tags: tags, Filename: filename, CacheUri: cacheURL}, nil
}

// getUrl downloads the given resource and returns the path to it.
func getUrl(u string, fileName string, bearer string, ctx context.Context) (string, error) {
	_, err := url.Parse(u)
	if err != nil {
		return "", fmt.Errorf("invalid url '%s': %s", u, err)
	}
	log.Debug().Str("URL", u).Msg("Downloading")

	req := requests.
		URL(u).
		Header("Accept", "*/*").
		ToFile(fileName)

	if bearer != "" {
		req.Header("Authorization", fmt.Sprintf("Bearer %s", bearer))
	}

	err = req.Fetch(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to download '%s': %s", u, err)
	}

	return fileName, nil
}

// GetUrlToDir downloads the given resource to the given directory and returns the path to it.
func GetUrlToDir(u string, targetDir string, bearer string, ctx context.Context) (string, error) {
	// create temporary name in the target directory.
	h := sha256.New()
	h.Write([]byte(u))
	fileName := filepath.Join(targetDir, fmt.Sprintf(".%s", hex.EncodeToString(h.Sum(nil))))
	return getUrl(u, fileName, bearer, ctx)
}

// GetUrlWithDir downloads the given resource to a temporary file and returns the path to it.
func GetUrltoTempFile(u string, bearer string, ctx context.Context) (string, error) {
	file, err := os.CreateTemp("", "prefix")
	if err != nil {
		log.Fatal().Err(err)
	}
	fileName := file.Name()
	return getUrl(u, fileName, "", ctx)
}

func (l *Resource) Download(dir string, mode os.FileMode, ctx context.Context) error {
	// Check if a cache URL exists to use Artifactory first
	if l.CacheUri != "" {
		token := os.Getenv("GRABIT_ARTIFACTORY_TOKEN")
		if token != "" {
			artifactoryURL := fmt.Sprintf("%s/%s", l.CacheUri, l.Integrity)
			localName := l.Filename
			if localName == "" {
				localName = path.Base(l.Urls[0])
			}
			resPath := filepath.Join(dir, localName)

			// Use getUrl directly with bearer token for Artifactory
			tmpPath, err := getUrl(artifactoryURL, resPath, token, ctx)
			if err == nil {
				if mode != NoFileMode {
					err = os.Chmod(tmpPath, mode.Perm())
				}
				if err == nil {
					return nil // Success
				}
			}
			fmt.Printf("Failed to download from Artifactory, falling back to original URL: %v\n", err)
		}
	}
	ok := false
	algo, err := getAlgoFromIntegrity(l.Integrity)
	if err != nil {
		return err
	}

	var downloadError error = nil
	for _, u := range l.Urls {
		localName := ""
		if l.Filename != "" {
			localName = l.Filename
		} else {
			localName = path.Base(u)
		}
		resPath := filepath.Join(dir, localName)

		// Check if the destination file already exists and has the correct integrity.
		_, err := os.Stat(resPath)
		if err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("error checking destination file presence '%s': '%v'", resPath, err)
			}
		} else {
			err = checkIntegrityFromFile(resPath, algo, l.Integrity, u)
			if err != nil {
				return fmt.Errorf("existing file at '%s' with incorrect integrity: '%v'", resPath, err)
			}
			if mode != NoFileMode {
				err = os.Chmod(resPath, mode.Perm())
				if err != nil {
					return err
				}
			}
			return nil
		}

		// Download file in the target directory so that the call to
		// os.Rename is atomic.
		lpath, err := GetUrlToDir(u, dir, "", ctx)
		if err != nil {
			downloadError = err
			continue
		}
		err = checkIntegrityFromFile(lpath, algo, l.Integrity, u)
		if err != nil {
			return err
		}
		err = os.Rename(lpath, resPath)
		if err != nil {
			return err
		}
		if mode != NoFileMode {
			err = os.Chmod(resPath, mode.Perm())
			if err != nil {
				return err
			}
		}
		ok = true
		if l.CacheUri != "" && os.Getenv("NO_CACHE_UPLOAD") != "1" {
			if uploadErr := uploadToArtifactory(resPath, l.CacheUri, l.Integrity); uploadErr != nil {
				fmt.Printf("Warning: Failed to upload to cache: %v\n", uploadErr)
			}
		}
		break
	}
	if !ok {
		if err == nil {
			if downloadError != nil {
				return downloadError
			} else {
				panic("no error but no file downloaded")
			}
		}
		return err
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
