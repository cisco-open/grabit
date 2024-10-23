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

// getIntegrityFromFile computes the integrity hash of a file given its path and the hashing algorithm.
// It reads the file in chunks of 10MB, computes the hash using the specified algorithm, and returns
// the hash as a base64-encoded string prefixed with the algorithm name.

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

// getIntegrityFromFile computes the integrity hash of a file given its path and the hashing algorithm.
// It reads the file in chunks of 10MB, computes the hash using the specified algorithm, and returns
// the hash as a base64-encoded string prefixed with the algorithm name.

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

// getAlgoFromIntegrity extracts the hashing algorithm from a given Subresource Integrity (SRI) string.
// It splits the SRI string on the '-' character and checks if the resulting slices are valid.
// If valid, it attempts to create a new hash object using the first part of the split string.
// If successful, it returns the algorithm name; otherwise, it returns an error.

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
