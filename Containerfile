FROM golang:1.25

# Install dependencies
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.5.0
RUN go install github.com/goreleaser/goreleaser/v2@latest

# Install security tools
RUN go install golang.org/x/vuln/cmd/govulncheck@latest
RUN go install github.com/securego/gosec/v2/cmd/gosec@latest

WORKDIR /app