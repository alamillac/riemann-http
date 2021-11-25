#!/bin/bash

data='{
    "service": "api.response_time",
    "description": "Response time of api call in ms",
    "metric": 12,
    "host": "www.tropipay.com",
    "ttl": 60,
    "state": "ok",
    "attributes": {
        "endpoint": "/"
    }
}'

echo curl -d $data 127.0.0.1:8080/metric
for i in {1..5}
do
curl -d "$data" 127.0.0.1:8080/metric &
done
