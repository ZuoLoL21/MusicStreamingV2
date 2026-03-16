#!/bin/sh
set -e

echo "Waiting for MinIO..."
until /usr/bin/mc alias set minio http://minio-test:9000 testuser testpassword; do
  sleep 1
done

echo "Creating bucket and setting policy..."
/usr/bin/mc mb minio/test-bucket --ignore-existing
/usr/bin/mc anonymous set download minio/test-bucket

echo "MinIO initialized successfully"
