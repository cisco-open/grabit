// Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package internal

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"strings"
)

// This Go code defines a mapping of hashing algorithms to their respective functions,
// including SHA-1, SHA-256, SHA-384, and SHA-512. It also specifies a recommended
// algorithm (SHA-256) and defines a Hasher type as a function that returns a hash.Hash.
// Additionally, a Hash struct is defined to hold the algorithm name and its corresponding
// hashing function. A variable for storing all algorithms is also declared but not initialized.

var algos = map[string]Hasher{
	"sha1":   sha1.New,
	"sha256": sha256.New,
	"sha384": sha512.New384,
	"sha512": sha512.New,
}

var RecommendedAlgo = "sha256"

type Hasher func() hash.Hash

type Hash struct {
	algo string
	hash Hasher
}

var allAlgos = ""

// This Go code initializes a list of algorithms by calling the initAlgoList function during the package initialization phase.

func init() {
	initAlgoList()
}

// This function initializes a list of algorithms by iterating over a predefined map of algorithms (algos).
// It checks for the presence of a recommended algorithm (RecommendedAlgo) within the list.
// If the recommended algorithm is not found, the function panics with an error message.
// The final list of algorithms is stored as a comma-separated string in the variable allAlgos.

func initAlgoList() {
	algoList := make([]string, 0, len(algos))
	foundRecommendedAlgo := false
	for algo := range algos {
		algoList = append(algoList, algo)
		if RecommendedAlgo == algo {
			foundRecommendedAlgo = true
		}
	}
	allAlgos = strings.Join(algoList, ", ")
	if !foundRecommendedAlgo {
		panic(fmt.Sprintf("cannot find recommended algorithm '%s'", RecommendedAlgo))
	}
}

// NewHash creates a new Hash instance for the specified algorithm.
// If the algorithm is unknown, it returns an error indicating the
// available algorithms.
func NewHash(algo string) (*Hash, error) {
	hash, ok := algos[algo]
	if !ok {
		return nil, fmt.Errorf("unknown hash algorithm '%s' (available algorithms: %s)", algo, allAlgos)
	}
	return &Hash{algo: algo, hash: hash}, nil
}
