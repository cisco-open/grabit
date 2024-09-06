// Copyright (c) 2023 Cisco Systems, Inc. and its affiliates
// All rights reserved.

package test

import (
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TmpFile(t *testing.T, content string) string {
	f, err := os.CreateTemp(t.TempDir(), "test")
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.WriteString(content)
	if err != nil {
		t.Fatal(err)
	}
	name := f.Name()
	t.Cleanup(func() { os.RemoveAll(name) })
	return name
}

func TmpDir(t *testing.T) string {
	dir, err := os.MkdirTemp(t.TempDir(), "test")
	if err != nil {
		log.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func HttpHandler(handler http.HandlerFunc) (int, *httptest.Server) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	s := httptest.NewUnstartedServer(http.HandlerFunc(handler))
	s.Listener.Close()
	s.Listener = l
	s.Start()
	return l.Addr().(*net.TCPAddr).Port, s
}
