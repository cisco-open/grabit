// Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package internal

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"

	toml "github.com/pelletier/go-toml/v2"
)

// This file defines a Lock structure representing a grabit lockfile.
// It contains a path to the lockfile and a configuration struct that holds
// a slice of Resource objects. Each Resource includes URLs, an integrity
// string, and tags for categorization.
var COMMENT_PREFIX = "//"

// Lock represents a grabit lockfile.
type Lock struct {
	path string
	conf config
}

type config struct {
	Resource []Resource
}

type resource struct {
	Urls      []string
	Integrity string
	Tags      []string
}

// NewLock creates a new Lock instance by checking if the specified file exists.
// If the file does not exist and newOk is true, it initializes a new Lock with the given path.
// If the file exists, it attempts to open and decode the file contents into a config structure.
// It returns an error if the file cannot be opened or if the decoding fails.

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

// AddResource adds a new resource to the Lock if it does not already exist.
// It checks if each path in the provided slice is already contained in the Lock.
// If a path is found, it returns an error indicating that the resource is already present.
// If all paths are unique, it creates a new Resource from the provided parameters
// and appends it to the Lock's configuration.

func (l *Lock) AddResource(paths []string, algo string, tags []string, filename string, dynamic bool) error {
	for _, u := range paths {
		if l.Contains(u) {
			return fmt.Errorf("resource '%s' is already present", u)
		}
	}
	r, err := NewResourceFromUrl(paths, algo, tags, filename, dynamic)
	if err != nil {
		return err
	}
	l.conf.Resource = append(l.conf.Resource, *r)
	return nil
}

// This function, DeleteResource, removes a resource identified by the given path from the Lock's configuration.
// It iterates through the existing resources and constructs a new list that excludes the resource containing the specified path.
// Finally, it updates the Lock's configuration with the new list of resources.

func (l *Lock) DeleteResource(path string) {
	newStatements := []Resource{}
	for _, r := range l.conf.Resource {
		if !r.Contains(path) {
			newStatements = append(newStatements, r)
		}
	}
	l.conf.Resource = newStatements
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
	_, err := strToFileMode(perm)
	if err != nil {
		return fmt.Errorf("'%s' is not a valid permission definition", perm)
	}

	filteredResources := l.filterResources(tags, notags)

	total := len(filteredResources)
	if total == 0 {
		return fmt.Errorf("nothing to download")
	}

	errorCh := make(chan error, total)
	for _, r := range filteredResources {
		resource := r
		go func() {
			err := resource.DownloadFile(resource.Urls[0], dir)
			if err != nil {
				errorCh <- fmt.Errorf("failed to download %s: %w", resource.Urls[0], err)
			} else {
				errorCh <- nil
			}
		}()
	}

	errs := []error{}
	for i := 0; i < total; i++ {
		if err := <-errorCh; err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// filterResources filters a slice of Resource based on the specified tags and notags.
// It returns a new slice of Resource that contains only those resources which have all the
// specified tags and do not have any of the specified notags.

func (l *Lock) filterResources(tags []string, notags []string) []Resource {
	tagFilteredResources := l.conf.Resource
	if len(tags) > 0 {
		tagFilteredResources = []Resource{}
		for _, r := range l.conf.Resource {
			if r.hasAllTags(tags) {
				tagFilteredResources = append(tagFilteredResources, r)
			}
		}
	}

	filteredResources := tagFilteredResources
	if len(notags) > 0 {
		filteredResources = []Resource{}
		for _, r := range tagFilteredResources {
			if !r.hasAnyTag(notags) {
				filteredResources = append(filteredResources, r)
			}
		}
	}

	return filteredResources
}

// hasAllTags checks if the Resource has all the specified tags.
// It iterates through each tag in the provided slice and returns false
// if any tag is not found in the Resource. If all tags are present, it returns true.

func (r *Resource) hasAllTags(tags []string) bool {
	for _, tag := range tags {
		if !r.hasTag(tag) {
			return false
		}
	}
	return true
}

// hasAnyTag checks if the Resource has any of the specified tags.
// It iterates through the provided tags and returns true if at least one tag matches,
// otherwise it returns false.

func (r *Resource) hasAnyTag(tags []string) bool {
	for _, tag := range tags {
		if r.hasTag(tag) {
			return true
		}
	}
	return false
}

// hasTag checks if a given tag exists in the Resource's Tags slice.
// It returns true if the tag is found, otherwise it returns false.

func (r *Resource) hasTag(tag string) bool {
	for _, rtag := range r.Tags {
		if tag == rtag {
			return true
		}
	}
	return false
}

// UpdateResource updates a resource in the Lock configuration based on the provided URL.
// If a resource containing the URL is found, it creates a new resource from the existing
// resource's properties and saves the updated configuration. If no resource is found,
// it returns an error indicating the resource was not found.

func (l *Lock) UpdateResource(url string) error {
	for i, r := range l.conf.Resource {
		if r.Contains(url) {
			newResource, err := NewResourceFromUrl(r.Urls, r.Integrity, r.Tags, r.Filename, r.Dynamic)
			if err != nil {
				return err
			}
			l.conf.Resource[i] = *newResource
			return l.Save()
		}
	}
	return fmt.Errorf("resource with URL '%s' not found", url)
}

// VerifyIntegrity checks the integrity of resources by validating each URL against its expected integrity value.
// It iterates through all resources and their associated URLs, returning an error if any integrity check fails.

func (l *Lock) VerifyIntegrity() error {
	for _, r := range l.conf.Resource {
		for _, url := range r.Urls {
			err := checkIntegrityFromUrl(url, r.Integrity)
			if err != nil {
				return fmt.Errorf("integrity check failed for %s: %w", url, err)
			}
		}
	}
	return nil
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
