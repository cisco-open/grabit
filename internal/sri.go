// Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package internal

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

func getIntegrityFromFile(path string, algo string) (string, error) {
	hash, err := NewHash(algo)
	if err != nil {
		return "", err
	}
	f, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("cannot open file '%s'", path)
	}
	hasher := hash.hash()
	hasher.Write(f)
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
