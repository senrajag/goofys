#!/bin/sh
DOCKERIMAGE=senrajag/golang-build:v0.1
docker build -t $DOCKERIMAGE -f build.Dockerfile .
docker run -v $(PWD):/work $DOCKERIMAGE make build