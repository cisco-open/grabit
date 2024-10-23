// Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package internal

import (
	"testing"

	"github.com/cisco-open/grabit/test"
	"github.com/stretchr/testify/assert"
)

// TestGetAlgoFromIntegrity tests the getAlgoFromIntegrity function by validating various SRI strings.
// It checks for valid and invalid cases, ensuring that the function returns the expected results and error messages.

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

// TestGetIntegrityFromFile verifies that the function getIntegrityFromFile correctly computes
// the Subresource Integrity (SRI) hash for a given file path and hash algorithm (sha256).
// It checks that no error occurs during the process and that the computed SRI matches the expected value.

func TestGetIntegrityFromFile(t *testing.T) {
	path := test.TmpFile(t, "abcdef")
	sri, err := getIntegrityFromFile(path, "sha256")
	assert.Nil(t, err)
	assert.Equal(t, "sha256-vvV+x/U6bUC+tkCngKY5yDvCmsipgW8fxsXG3Nk8RyE=", sri)
}

// TestCheckIntegrityFromFile tests the checkIntegrityFromFile function by attempting to open a non-existent file,
// expecting an error that indicates the file cannot be opened.

func TestCheckIntegrityFromFile(t *testing.T) {
	err := checkIntegrityFromFile("/non/existant/path/test", "sha256", "not-used", "")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "cannot open file")
}

// TestCheckIntegrityFromFileInvalid tests the checkIntegrityFromFile function for handling an invalid integrity value.
// It verifies that the function returns an error indicating an integrity mismatch when provided with an invalid checksum.

func TestCheckIntegrityFromFileInvalid(t *testing.T) {
	path := test.TmpFile(t, "abcdef")
	err := checkIntegrityFromFile(path, "sha256", "invalid", "")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "integrity mismatch")
}
