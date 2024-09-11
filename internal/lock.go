// Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package internal

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	toml "github.com/pelletier/go-toml/v2"
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

func (l *Lock) AddResource(paths []string, algo string, tags []string, filename string) error {
	for _, u := range paths {
		if l.Contains(u) {
			return fmt.Errorf("resource '%s' is already present", u)
		}
	}
	r, err := NewResourceFromUrl(paths, algo, tags, filename)
	if err != nil {
		return err
	}
	l.conf.Resource = append(l.conf.Resource, *r)
	return nil
}

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
func (l *Lock) Download(dir string, tags []string, notags []string, perm string, bar bool) error {
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

	//This progress goroutine will run concurrently with the download goroutines.
	//When each download goroutine is finished, it places a 1 in the progressCh channel.
	//The progress goroutine keeps a tally of how many downloads have finished by adding whatever is in the channel to the total (a 1 for each completed download).
	//Example:
	//[■■■■■■■#___]   7 of 11 Complete
	//		■   completed download
	//		#   current, active download
	//		_	download yet to be started
	progressCh := make(chan int)
	if bar {
		spinChars := [6]string{"-", "\\", "|", "/", "-", "\\"}
		spinI := 0 //Current char in spinChars.

		//The progress bar goroutine blocks and waits for items to enter the progressCh channel.
		//So the spinner would only update when a download completes (when a download completes, it places a 1 in progressCh)
		//We want the spinner to continuously update, so we continuously feed in 0's to the progressCh channel (every 50 milliseconds).
		//This keeps the goroutine running and printing, and the extra 0's don't change downloadTotal.
		ticker := time.NewTicker(50 * time.Millisecond)
		go func() {
			for {
				select {
				case <-ticker.C:
					progressCh <- 0
				}
			}
		}()

		go func() {
			downloadTotal := 0
			for {
				downloadTotal += <-progressCh

				var spinner string
				if downloadTotal < len(filteredResources) {
					spinner = spinChars[spinI]
					spinI += 1
					if spinI == 5 {
						spinI = 0
					}
				} else {
					spinner = "✔"
				}

				//Bar is yellow while downloading, green when complete.
				var color string
				if downloadTotal < len(filteredResources) {
					color = "yellow"
				} else {
					color = "green"
				}

				bar := "["
				for i := 0; i < downloadTotal; i += 1 {
					bar += "█"
				}

				if downloadTotal < len(filteredResources) {
					bar += "░"
				}

				for i := downloadTotal + 1; i < len(filteredResources); i += 1 {
					bar += "_"
				}

				bar += "]"

				//"\r" allows the bar to clear and update on one line.
				line := "\r" + spinner + bar + "   " + strconv.Itoa(downloadTotal) + " of " + strconv.Itoa(len(filteredResources)) + " Complete"
				fmt.Print(Color_Text(line, color))

				if downloadTotal == len(filteredResources) {
					fmt.Println()
					break
				}
			}
		}()
	}
	for _, r := range filteredResources {
		resource := r
		go func() {
			err := resource.Download(dir, mode, ctx)

			errorCh <- err

			progressCh <- 1
		}()
	}
	done := 0
	for err := range errorCh {
		if err != nil {
			return err
		}
		done += 1
		if done == total {
			return nil
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
