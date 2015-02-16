.PHONY: default run build

default: build run

.built:
	docker build -t hallo .

build: .built

run:
	docker run --rm -it -v `pwd`:/go/src/app -p 3000:3000 hallo
