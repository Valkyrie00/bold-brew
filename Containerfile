FROM golang:1.25

# Install dependencies
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.1.6
RUN go install github.com/goreleaser/goreleaser/v2@latest

WORKDIR /app