go-dummy-plugin.so:
	docker run --rm --platform=linux/amd64 -v $(shell pwd):/src -w /src golang:1.18 go build -trimpath -buildmode=plugin -o /src/data/go-dummy-plugin.so -v /src/plugin/dummy/dummy.go

go-bridge-plugin.so:
	docker run --rm --platform=linux/amd64 -v $(shell pwd):/src -w /src golang:1.18 go build -trimpath -buildmode=c-shared -o /src/data/go-bridge-cshared.so -v /src/bridge.go

all: go-dummy-plugin.so go-bridge-plugin.so
	docker run -it --rm --platform=linux/amd64 -v $(shell pwd)/data:/data ghcr.io/calyptia/enterprise:main /fluent-bit/bin/fluent-bit -c /data/fluentbit.conf
