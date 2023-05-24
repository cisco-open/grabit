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

func init() {
	initAlgoList()
}

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

func NewHash(algo string) (*Hash, error) {
	hash, ok := algos[algo]
	if !ok {
		return nil, fmt.Errorf("unknown hash algorithm '%s' (available algorithms: %s)", algo, allAlgos)
	}
	return &Hash{algo: algo, hash: hash}, nil
}
