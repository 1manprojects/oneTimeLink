#!/bin/bash

env CGO_ENABLED=0 GOOS=linux /usr/local/go/bin/go build *.go
docker build -t onetimelink .