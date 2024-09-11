// Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package internal

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
)

func getIntegrityFromFile(path string, algo string) (string, error) {
	hash, err := NewHash(algo)
	if err != nil {
		return "", err
	}
	hasher := hash.hash()
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("cannot open file '%s'", path)
	}
	defer f.Close()
	reader := bufio.NewReader(f)
	// each block size is 10MB
	const chunkSize = 10 * 1024 * 1024
	buf := make([]byte, chunkSize)
	for {
		n, err := reader.Read(buf)
		buf = buf[:n]

		if n == 0 {
			if err != nil {
				if err == io.EOF {
					break
				}
				return "", err
			}
			break
		}
		hasher.Write(buf)
	}
	h := hasher.Sum(nil)
	return fmt.Sprintf("%s-%s", algo, base64.StdEncoding.EncodeToString(h)), nil
}

func checkIntegrityFromFile(path string, algo string, integrity string, u string) error {
	computedIntegrity, err := getIntegrityFromFile(path, algo)
	if err != nil {
		return fmt.Errorf("failed to compute ressource integrity: %s", err)
	}
	if computedIntegrity != integrity {
		return fmt.Errorf("integrity mismatch for '%s': got '%s' expected '%s'", u, computedIntegrity, integrity)
	}
	return nil
}

func getAlgoFromIntegrity(integrity string) (string, error) {
	splits := strings.Split(integrity, "-")
	if len(splits) < 2 {
		return "", fmt.Errorf("invalid SRI '%s'", integrity)
	}
	hash, err := NewHash(splits[0])
	if err != nil {
		return "", err
	}
	return hash.algo, nil
}
