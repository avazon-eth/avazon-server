#!/bin/bash

touch app.log
touch test.db

go mod tidy
docker-compose up --build -d