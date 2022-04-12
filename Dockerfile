FROM golang:1.18 as build

# Install certificates
# hadolint ignore=DL3008,DL3015
RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /go/src/github.com/calyptia/enterprise-plugin-dummy
# Allow us to cache go module download if source code changes
COPY go.mod .
COPY go.sum .
RUN go mod download

ENV CGO_ENABLED=1
# Now do the rest of the source code - this way we can speed up local iteration
COPY plugin/dummy/dummy.go ./plugin/dummy/
RUN go build -trimpath -buildmode=plugin -o bin/go-dummy-plugin.so ./plugin/dummy/dummy.go

COPY bridge.go .
ENV LD_PRELOAD=/go/src/github.com/calyptia/enterprise-plugin-dummy/bin/go-dummy-plugin.so
RUN go build -trimpath -buildmode=c-shared -o bin/go-bridge-cshared.so ./bridge.go
