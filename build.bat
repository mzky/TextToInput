@echo off
go env -w CGO_ENABLED=0
go mod tidy
go build -ldflags "-w -s"
