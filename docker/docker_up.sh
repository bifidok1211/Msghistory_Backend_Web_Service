#! /bin/bash

cd /home/rubinshtern/work/RIP/docker
cd minio
docker compose up -d
cd ..
cd postgres
docker compose up -d
cd ..
cd redis
docker compose up -d
