#!/bin/bash

USER=$1
PASSWORD=$2
for i in {1..10000}
do
  data='{
      "service": "metric.test",
      "description": "Response time of api call in ms",
      "metric": '$(( $RANDOM % 10 ))',
      "host": "www.tropipay.com",
      "ttl": 60,
      "state": "ok",
      "attributes": {
          "endpoint": "/"
      }
  }'
  echo curl -d "$data" https://monitor.tropipay.com/metric
  curl -u $USER:$PASSWORD -d "$data" https://monitor.tropipay.com/metric
  #echo curl -d "$data" 127.0.0.1:8080/metric
  #curl -u user:password -d "$data" 127.0.0.1:8080/metric
  sleep 2
done
