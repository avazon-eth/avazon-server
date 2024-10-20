#!/bin/bash

docker build -t avazon-api:latest .
docker tag avazon-api:latest seongduk/avazon-api:latest
docker push seongduk/avazon-api:latest
