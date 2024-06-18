#!/bin/sh

docker build -t riemannhttp .
docker tag riemannhttp alamilla/riemann-http:v4
docker tag riemannhttp alamilla/riemann-http:latest
docker push alamilla/riemann-http:v4
