.PHONY: default run build

default: build run

.built:
	docker build -t hallo .
	touch .built

build: .built

run:
	docker run --rm -it -v `pwd`:/go/src/app --name john1 -p `curl curlmyip.com`::3000 hallo
