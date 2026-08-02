# Format-check + lint targets. Run `make` for the full suite, or invoke a
# single target like `make fmt`.

GO            ?= go
PKG           := ./...
LOCAL_MODULE  := github.com/admin-template/backend

.PHONY: help
help:
	@echo "Targets:"
	@echo "  fmt       - gofumpt + goimports (write)"
	@echo "  fmtcheck  - gofumpt + goimports (CI gate)"
	@echo "  lint      - golangci-lint run"
	@echo "  test      - go test ./..."
	@echo "  e2e       - go test -tags=e2e ./e2e/..."
	@echo "  all       - fmtcheck + lint + test"

.PHONY: fmt
fmt:
	gofumpt -w .
	goimports -w -local $(LOCAL_MODULE) .

.PHONY: fmtcheck
fmtcheck:
	@out=$$(gofumpt -l .); if [ -n "$$out" ]; then echo "gofumpt would reformat:"; echo "$$out"; exit 1; fi
	@out=$$(goimports -l -local $(LOCAL_MODULE) .); if [ -n "$$out" ]; then echo "goimports would reformat:"; echo "$$out"; exit 1; fi

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: test
test:
	$(GO) test -shuffle=on -count=1 $(PKG)

.PHONY: e2e
e2e:
	$(GO) test -tags=e2e -count=1 -timeout=120s ./e2e/...

.PHONY: all
all: fmtcheck lint test