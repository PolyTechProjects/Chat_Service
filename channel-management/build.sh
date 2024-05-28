#!/bin/bash

if [[ "$OSTYPE" == "linux-gnu"* ]]; then
  go build -C src/ -o ../build/ChannelManagementService-001-build
elif [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" ]]; then
  GOOS=linux GOARCH=amd64 go build -C src/ -o ../build/ChannelManagementService-001-build
else
  echo "Unsupported OS: $OSTYPE"
  exit 1
fi
