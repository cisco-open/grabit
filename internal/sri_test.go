// Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAlgoFromIntegrity(t *testing.T) {
	tests := []struct {
		sriString     string
		valid         bool
		errorContains string
		resString     string
	}{
		{
			sriString:     "test123-aaa",
			valid:         false,
			errorContains: "unknown hash algorithm",
		},
		{
			sriString:     "test123",
			valid:         false,
			errorContains: "invalid SRI",
		},
		{
			sriString: "sha1-aaa",
			valid:     true,
			resString: "sha1",
		},
	}

	for _, data := range tests {
		algo, err := getAlgoFromIntegrity(data.sriString)
		assert.Equal(t, data.valid, err == nil)
		if err != nil {
			assert.Contains(t, err.Error(), data.errorContains)
		} else {
			assert.Equal(t, data.resString, algo)
		}
	}
}

func TestGetIntegrityFromFile(t *testing.T) {
	path := tmpFile(t, "abcdef")
	sri, err := getIntegrityFromFile(path, "sha256")
	assert.Nil(t, err)
	assert.Equal(t, "sha256-vvV+x/U6bUC+tkCngKY5yDvCmsipgW8fxsXG3Nk8RyE=", sri)
}

func TestCheckIntegrityFromFile(t *testing.T) {
	err := checkIntegrityFromFile("/non/existant/path/test", "sha256", "not-used", "")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "cannot open file")
}

func TestCheckIntegrityFromFileInvalid(t *testing.T) {
	path := tmpFile(t, "abcdef")
	err := checkIntegrityFromFile(path, "sha256", "invalid", "")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "integrity mismatch")
}
