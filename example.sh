#!/bin/bash

USER=$1
PASSWORD=$2
for i in {1..1000}
do
  data='{
      "service": "api.response_time",
      "description": "Example metric",
      "metric": '$(( $RANDOM % 10 ))',
      "host": "www.example.com",
      "ttl": 60,
      "state": "ok",
      "attributes": {
        "ip": "152.206.232.215",
        "statusCode": "200",
        "method": "POST",
        "url": "/api/countries",
        "country": "CU"
      }
  }'

  #echo curl -d "$data" https://monitor.tropipay.com/metric
  #curl -u $USER:$PASSWORD -d "$data" https://monitor.tropipay.com/metric
  echo curl -d "$data" 127.0.0.1:8080/metric
  curl -u user:pass -d "$data" 127.0.0.1:8080/metric
  sleep 1
done
