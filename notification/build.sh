#!/bin/bash

GOOS=linux GOARCH=amd64 go build -C src/ -o ../build/NotificationApp-001-build
