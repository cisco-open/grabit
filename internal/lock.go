// Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package internal

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/carlmjohnson/requests"
	toml "github.com/pelletier/go-toml/v2"
	"github.com/rs/zerolog/log"
)

var COMMENT_PREFIX = "//"

// Lock represents a grabit lockfile.
type Lock struct {
	path string
	conf config
}

type config struct {
	Resource []Resource
}

func NewLock(path string, newOk bool) (*Lock, error) {
	_, error := os.Stat(path)
	if os.IsNotExist(error) {
		if newOk {
			return &Lock{path: path}, nil
		} else {
			return nil, fmt.Errorf("file '%s' does not exist", path)
		}
	}
	var conf config
	file, err := os.Open(path)
	if err != nil {
		return nil, error
	}
	d := toml.NewDecoder(file)
	err = d.Decode(&conf)
	if err != nil {
		return nil, err
	}

	return &Lock{path: path, conf: conf}, nil
}

func (l *Lock) AddResource(paths []string, algo string, tags []string, filename string, cacheURL string) error {
	for _, u := range paths {
		if l.Contains(u) {
			return fmt.Errorf("resource '%s' is already present", u)
		}
	}
	r, err := NewResourceFromUrl(paths, algo, tags, filename, cacheURL)
	if err != nil {
		return err
	}

	// If cache URL is provided, handles Artifactory upload
	if cacheURL != "" {
		token := os.Getenv("GRABIT_ARTIFACTORY_TOKEN")
		if token == "" {
			return fmt.Errorf("GRABIT_ARTIFACTORY_TOKEN environment variable is not set")
		}

		// Get file again for upload
		path, err := GetUrltoTempFile(paths[0], context.Background())
		if err != nil {
			return fmt.Errorf("failed to get file for cache: %s", err)
		}
		defer os.Remove(path)

		// Upload to Artifactory using hash as filename
		err = uploadToArtifactory(path, cacheURL, r.Integrity)
		if err != nil {
			return fmt.Errorf("failed to upload to cache: %v", err)
		}
	}

	l.conf.Resource = append(l.conf.Resource, *r)
	return nil
}

func uploadToArtifactory(filePath, cacheURL, integrity string) error {
	token := os.Getenv("GRABIT_ARTIFACTORY_TOKEN")
	if token == "" {
		return fmt.Errorf("GRABIT_ARTIFACTORY_TOKEN environment variable is not set")
	}

	// Read the content of the file
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	// Compute the SHA256 hash and generate the Artifactory URL
	hash := sha256.Sum256(fileContent)
	hexHash := fmt.Sprintf("sha256-%s", hex.EncodeToString(hash[:]))
	artifactoryURL := fmt.Sprintf("%s/%s", cacheURL, hexHash)

	// Upload the file using the requests package
	err = requests.
		URL(artifactoryURL).
		Method(http.MethodPut).
		Header("Authorization", fmt.Sprintf("Bearer %s", token)).
		BodyBytes(fileContent).
		Fetch(context.Background())

	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	return nil
}

func (l *Lock) DeleteResource(path string) {
	newStatements := []Resource{}
	for _, r := range l.conf.Resource {
		if !r.Contains(path) {
			newStatements = append(newStatements, r)
		} else if r.Contains(path) && r.CacheUri != "" {
			token := os.Getenv("GRABIT_ARTIFACTORY_TOKEN")
			if token == "" {
				log.Warn().Msg("Warning: Unable to delete from Artifcatory: GRABIT_ARTIFACTORY_TOKEN not set.")

				continue
			}

			artifactoryURL := fmt.Sprintf("%s/%s", r.CacheUri, r.Integrity)

			err := deleteCache(artifactoryURL, token)
			if err != nil {
				log.Warn().Msg("Warning: Unable to delete from Artifcatory")
			}
		}
	}
	l.conf.Resource = newStatements
}

func deleteCache(url, token string) error {
	// Create and send a DELETE request with an Authorization header.
	err := requests.
		URL(url).                                                 // Target URL for the DELETE request.
		Method(http.MethodDelete).                                // Set HTTP method to DELETE.
		Header("Authorization", fmt.Sprintf("Bearer %s", token)). // Add token in the Authorization header.
		Fetch(context.Background())                               // Execute the request.

	if err != nil {
		return err
	}

	return nil
}

const NoFileMode = os.FileMode(0)

// strToFileMode converts a string to a os.FileMode.
func strToFileMode(perm string) (os.FileMode, error) {
	if perm == "" {
		return NoFileMode, nil
	}
	parsed, err := strconv.ParseUint(perm, 8, 32)
	if err != nil {
		return NoFileMode, err
	}
	return os.FileMode(parsed), nil
}

// Download gets all the resources in this lock file and moves them to
// the destination directory.
func (l *Lock) Download(dir string, tags []string, notags []string, perm string) error {
	if stat, err := os.Stat(dir); err != nil || !stat.IsDir() {
		return fmt.Errorf("'%s' is not a directory", dir)
	}
	mode, err := strToFileMode(perm)
	if err != nil {
		return fmt.Errorf("'%s' is not a valid permission definition", perm)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Filter in the resources that have all the required tags.
	tagFilteredResources := []Resource{}
	if len(tags) > 0 {
		for _, r := range l.conf.Resource {
			hasAllTags := true
			for _, tag := range tags {
				hasTag := false
				for _, rtag := range r.Tags {
					if tag == rtag {
						hasTag = true
						break
					}
				}
				if !hasTag {
					hasAllTags = false
					break
				}
			}
			if hasAllTags {
				tagFilteredResources = append(tagFilteredResources, r)
			}
		}
	} else {
		tagFilteredResources = l.conf.Resource
	}
	// Filter out the resources that have any 'notag' tag.
	filteredResources := []Resource{}
	if len(notags) > 0 {
		for _, r := range tagFilteredResources {
			hasTag := false
			for _, notag := range notags {
				for _, rtag := range r.Tags {
					if notag == rtag {
						hasTag = true
					}
				}
			}
			if !hasTag {
				filteredResources = append(filteredResources, r)
			}
		}
	} else {
		filteredResources = tagFilteredResources
	}

	total := len(filteredResources)
	if total == 0 {
		return fmt.Errorf("nothing to download")
	}
	errorCh := make(chan error, total)
	for _, r := range filteredResources {
		resource := r
		go func() {
			// See if the resource has an Artifactory URL
			if resource.CacheUri != "" {
				// Find the correct filename
				filename := resource.Filename
				if filename == "" {
					filename = path.Base(resource.Urls[0])
				}
				// Build the full file path
				fullPath := filepath.Join(dir, filename)

				err := downloadFromArtifactory(ctx, resource.CacheUri, resource.Integrity, fullPath, mode)
				if err == nil {
					errorCh <- nil
					return
				}
				// Show a warning only for connection errors
				if strings.Contains(err.Error(), "lookup invalid") || strings.Contains(err.Error(), "dial tcp") {
					fmt.Printf("Failed to download from Artifactory, falling back to original URL: %v\n", err)
				}
			}

			err := resource.Download(dir, mode, ctx)
			errorCh <- err
		}()
	}
	done := 0
	errs := []error{}
	for range total {
		err = <-errorCh
		if err != nil {
			errs = append(errs, err)
		} else {
			done += 1
		}
	}
	if done == total {
		return nil
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func downloadFromArtifactory(ctx context.Context, cacheURL string, integrity string, filePath string, mode os.FileMode) error {
	// Check if Artifactory token is set in environment
	token := os.Getenv("GRABIT_ARTIFACTORY_TOKEN")
	if token == "" {
		return fmt.Errorf("GRABIT_ARTIFACTORY_TOKEN environment variable is not set")
	}

	// Create the full Artifactory URL using the cache URL and the file's integrity hash
	artifactoryURL := fmt.Sprintf("%s/%s", cacheURL, integrity)
	req, err := http.NewRequestWithContext(ctx, "GET", artifactoryURL, nil)
	if err != nil {
		return err
	}

	// Add authentication token to request
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check if the request to Artifactory was successful
	if resp.StatusCode >= 400 {
		return fmt.Errorf("failed to download from Artifactory, status code: %d", resp.StatusCode)
	}

	// Create a temporary file to download the content into
	tmpFile, err := os.CreateTemp(filepath.Dir(filePath), "download-*")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Save the downloaded content into the temporary file
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return err
	}
	tmpFile.Close()

	// Check if the file's hash matches the expected hash
	algo, err := getAlgoFromIntegrity(integrity)
	if err != nil {
		return err
	}
	err = checkIntegrityFromFile(tmpPath, algo, integrity, artifactoryURL)
	if err != nil {
		// Show a warning if the file validation fails
		log.Warn().Msgf("Cache validation failed for %s, falling back to original URL\n", artifactoryURL)
		return err
	}

	// Set the permissions for the downloaded file
	if err := os.Chmod(tmpPath, mode); err != nil {
		return err
	}

	// Move the temporary file to the final location
	return os.Rename(tmpPath, filePath)
}

// Save this lock file to disk.
func (l *Lock) Save() error {
	res, err := toml.Marshal(l.conf)
	if err != nil {
		return err
	}
	file, err := os.Create(l.path)
	if err != nil {
		return err
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	_, err = w.Write(res)
	if err != nil {
		return err
	}
	w.Flush()
	return nil
}

// Contains returns true if this lock file contains the
// given resource url.
func (l *Lock) Contains(url string) bool {
	for _, r := range l.conf.Resource {
		for _, u := range r.Urls {
			if url == u {
				return true
			}
		}
	}
	return false
}
