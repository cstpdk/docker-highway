#!/bin/bash

docker run --name etcd-public --rm -it -p 4001:4001 \
	quay.io/coreos/etcd:v2.0.0 \
	--listen-client-urls http://0.0.0.0:4001 \
	--advertise-client-urls http://`curl curlmyip.com`:4001
