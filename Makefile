# Copyright 2023 Cisco Systems, Inc. and its affiliates
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# SPDX-License-Identifier: Apache-2.0

all: build 
.PHONY: dep fmt build test lint

dep:
	@go mod tidy

update-dep:
	@go get -u ./...

fmt:
	@go fmt ./...

build: dep
	@go build -o grabit

test: dep
	@go test -race -v -coverpkg=./... -coverprofile=coverage.out ./...
	@go tool cover -func coverage.out | tail -n 1

check: dep
	@golangci-lint run --sort-results ./...
