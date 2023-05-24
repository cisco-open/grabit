// Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package internal

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewResourceFromUrl(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`abcdef`))
		if err != nil {
			t.Fatal(err)
		}
	}
	port, server := httpHandler(handler)
	defer server.Close()
	algo := "sha256"
	tests := []struct {
		urls          []string
		valid         bool
		errorContains string
		res           Resource
	}{
		{
			urls:  []string{fmt.Sprintf("http://localhost:%d/test.html", port)},
			valid: true,
			res:   Resource{Urls: []string{fmt.Sprintf("http://localhost:%d/test.html", port)}, Integrity: fmt.Sprintf("%s-vvV+x/U6bUC+tkCngKY5yDvCmsipgW8fxsXG3Nk8RyE=", algo), Tags: []string{}, Filename: ""},
		},
		{
			urls:          []string{"invalid url"},
			valid:         false,
			errorContains: "failed to download",
		},
	}

	for _, data := range tests {
		resource, err := NewResourceFromUrl(data.urls, algo, []string{}, "")
		assert.Equal(t, err == nil, data.valid)
		if err != nil {
			assert.Contains(t, err.Error(), data.errorContains)
		} else {
			assert.Equal(t, *resource, data.res)
		}
	}
}
