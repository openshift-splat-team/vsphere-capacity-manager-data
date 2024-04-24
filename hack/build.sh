#!/bin/bash
go build -o ./bin/generate -gcflags "all=-N -l" cmd/generate.go

