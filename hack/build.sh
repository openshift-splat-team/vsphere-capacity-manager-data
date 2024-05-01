#!/bin/bash
go build -o ./bin/vcmd -gcflags "all=-N -l" main.go

