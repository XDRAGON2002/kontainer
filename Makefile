run:
	@go run ./main.go

build:
	@go build -o ./bin/kontainer ./main.go

setup:
	@mkdir -p ./rootfs && \
	cd ./rootfs && \
	curl -s -o ./alpine.tar.gz http://dl-cdn.alpinelinux.org/alpine/v3.16/releases/x86_64/alpine-minirootfs-3.16.0-x86_64.tar.gz && \
	tar xf ./alpine.tar.gz && \
	rm -rf ./alpine.tar.gz && \
	touch ./ALPINE_ROOTFS

teardown:
	@rm -rf ./rootfs

clean:
	@rm -rf ./bin

build-all: build setup

clean-all: teardown clean
