#!/bin/bash
set -e
REPO=$1
FUNC=$2
REGION=${3:-us-east-1}

for i in {1..3}; do
  docker build -t "$REPO:$FUNC" . && break
  sleep 2
done
