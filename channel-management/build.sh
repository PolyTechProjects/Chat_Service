#!/bin/bash

GOOS=linux GOARCH=amd64 go build -C src/ -o ../build/ChannelManagementService-001-build
