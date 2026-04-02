GOBIN ?= $(shell go env GOPATH)/bin

.PHONY: test vet lint vulncheck build ci

test:
	go test ./...

vet:
	go vet ./...

lint:
	@test -x $(GOBIN)/golangci-lint || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOBIN)/golangci-lint run ./...

vulncheck:
	@test -x $(GOBIN)/govulncheck || go install golang.org/x/vuln/cmd/govulncheck@latest
	$(GOBIN)/govulncheck ./...

build:
	go build .

ci: test vet lint vulncheck build
	@echo "\nAll checks passed."
