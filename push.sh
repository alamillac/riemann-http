#!/bin/sh

docker build -t riemannhttp .
docker tag riemannhttp alamilla/riemann-http
docker push alamilla/riemann-http
