.PHONY: build run run-main run-etcd run-proxy run-dnsmasq

RUN_CMD := sh -c 'go install && app $$ETCD_PORT_4001_TCP_ADDR:$$ETCD_PORT_4001_TCP_PORT'

default: build run

build: .built

.built:
	docker build -t `basename $(shell pwd)` .
	touch .built

run: run-etcd run-proxy run-dnsmasq run-main

run-dnsmasq:
	docker run -d --name dnsmasq -p 53:53/udp -p 53:53 cstpdk/dnsmasq

run-etcd:
	docker run --name etcd -d -h etcd -p 172.17.42.1::4001 \
		quay.io/coreos/etcd:v2.0.0 \
		--listen-client-urls http://etcd:4001 \
		--advertise-client-urls http://etcd:4001
	until docker run --link etcd:etcd --rm speg03/curl \
		etcd -L http://etcd:4001/v2/keys/config -XPUT \
		-d dir=true ; do \
		sleep 0.1 ; \
	done

run-proxy:
	docker run -d -p 80:80 --name proxy --entrypoint sh \
		--link etcd:etcd cstpdk/haproxy-confd \
		-c 'confd -node=$$ETCD_PORT_4001_TCP_ADDR:$$ETCD_PORT_4001_TCP_PORT -interval=1'

run-main:
	docker run -d -v `pwd`:/go/src/app --link etcd:etcd \
		-v /var/run/docker.sock:/var/run/docker.sock \
		`basename $(shell pwd)` $(RUN_CMD)

run-gin:
	docker run -it -v `pwd`:/go/src/app `basename $(shell pwd)` gin

stop:
	-docker stop -t 2 `basename $(shell pwd)` | xargs docker rm -v
	-docker stop -t 2 proxy | xargs docker rm -v
	-docker stop -t 2 etcd | xargs docker rm -v
	-docker stop -t 2 dnsmasq | xargs docker rm -v
