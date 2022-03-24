FROM golang:1.18 as build

# Install certificates
# hadolint ignore=DL3008,DL3015
RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /go/src/github.com/calyptia/enterprise-plugin-dummy
# Allow us to cache go module download if source code changes
COPY go.* ./
RUN go mod download

# Now do the rest of the source code - this way we can speed up local iteration
COPY . .
RUN go build -buildmode=c-shared -trimpath -v ./...
