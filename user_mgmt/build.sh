#!/bin/bash

GOOS=linux GOARCH=amd64 go build -C src/ -o ../build/UserMgmtApp-001-build
