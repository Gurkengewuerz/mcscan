#!/bin/bash
USERNAME="gurken2108"
PROJECT="mcscan"
REGISTRY="docker.io"

docker build --no-cache -t ${REGISTRY}/${USERNAME}/${PROJECT}:latest .
#docker image push ${REGISTRY}/${USERNAME}/${PROJECT}:latest

echo -e "Done!"