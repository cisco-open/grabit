// Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package internal

import (
	"boguscontext"
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"

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
	path, err := GetUrl(url, ctx)
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

// GetUrl downloads the given resource and returns the path to it.
func GetUrl(u string, ctx context.Context) (string, error) {
	_, err := url.Parse(u)
	if err != nil {
		return "", fmt.Errorf("invalid url '%s': %s", u, err)
	}
	file, err := os.CreateTemp("", "prefix")
	if err != nil {
		log.Fatal(err)
	}
	err = requests.
		URL(u).
		ToFile(file.Name()).
		Fetch(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to download '%s': %s", u, err)
	}
	return file.Name(), nil
}

func (l *Resource) Download(dir string, ctx context.Context) error {
	ok := false
	algo, err := getAlgoFromIntegrity(l.Integrity)
	if err != nil {
		return err
	}
	for _, u := range l.Urls {
		lpath, err := GetUrl(u, ctx)
		if err != nil {
			break
		}
		err = checkIntegrityFromFile(lpath, algo, l.Integrity, u)
		if err != nil {
			return err
		}

		localName := ""
		if l.Filename != "" {
			localName = l.Filename
		} else {
			localName = path.Base(u)
		}
		err = os.Rename(lpath, filepath.Join(dir, localName))
		if err != nil {
			return err
		}
		ok = true
	}
	if !ok {
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
