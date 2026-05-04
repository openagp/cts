.PHONY: build test vectors sync-spec install clean

BIN := agp-cts
SRC := ./cmd/agp-cts

# Build a static binary for the host platform.
build:
	@CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BIN) $(SRC)
	@echo "built: ./$(BIN)"

# Cross-compile a release matrix.
release: clean
	@mkdir -p dist
	@for goos in linux darwin windows; do \
	  for goarch in amd64 arm64; do \
	    out=dist/$(BIN)-$$goos-$$goarch; \
	    [ "$$goos" = "windows" ] && out=$$out.exe; \
	    echo "building $$out ..."; \
	    CGO_ENABLED=0 GOOS=$$goos GOARCH=$$goarch \
	      go build -ldflags="-s -w" -o $$out $(SRC); \
	  done; \
	done
	@ls -la dist

# Run all unit + vector tests.
test:
	@go test ./... -count=1

# Run the embedded vectors via the CLI (sanity-check post-build).
vectors: build
	@./$(BIN) vectors

# Sync schemas + test vectors from a sibling spec checkout.
# Run after pulling new spec changes; commit any resulting diff.
sync-spec:
	@scripts/sync-spec.sh

install: build
	@install -m 755 $(BIN) $(GOPATH)/bin/$(BIN) 2>/dev/null || \
	 install -m 755 $(BIN) $$HOME/go/bin/$(BIN) 2>/dev/null || \
	 (echo "GOPATH not set and ~/go/bin missing; copy ./$(BIN) to your PATH manually"; exit 1)
	@echo "installed: $(BIN)"

clean:
	@rm -rf $(BIN) dist
	@go clean -testcache
