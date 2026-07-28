#!/bin/bash

ENV_FILE=.env.local

set -e

check_api() {
  echo docker build --target=api --tag=revo_web --file Dockerfile .
  docker build --target=api --tag=revo_web --file Dockerfile .
  echo docker run --rm --env-file ${ENV_FILE} revo_web
  docker run --rm --env-file ${ENV_FILE} revo_web
}

check_release() {
  echo docker build --target=release --tag=revo_release --file Dockerfile .
  docker build --target=release --tag=revo_release --file Dockerfile .
  echo docker run --rm --env-file ${ENV_FILE} revo_release
  docker run --rm --env-file ${ENV_FILE} revo_release
}

check_release
check_api
