TINYGO ?= tinygo
WASM_OPT ?= wasm-opt
GOPATH ?= $(shell go env GOPATH)
WORKERS_JS := $(GOPATH)/pkg/mod/github.com/syumai/workers@v0.32.0/cmd/_assets/main.js

WORKERS := casino classic solo

.PHONY: build-workers $(addprefix build-worker-,$(WORKERS)) clean-workers

build-workers: $(addprefix build-worker-,$(WORKERS))

build-worker-%:
	@echo "Building worker: $*"
	@mkdir -p build/$*
	$(TINYGO) build -o build/$*/app.wasm -target wasi -no-debug -opt=z ./cmd/workers/$*
	$(WASM_OPT) --enable-bulk-memory --enable-nontrapping-float-to-int --enable-sign-ext -Oz build/$*/app.wasm -o build/$*/app.wasm
	@cp $(WORKERS_JS) build/$*/worker.js
	@GZIP=$$(gzip -c build/$*/app.wasm | wc -c); \
	echo "  $*: $$(stat -c%s build/$*/app.wasm) bytes raw, $$GZIP bytes gzip"

clean-workers:
	rm -rf build/casino build/classic build/solo

dev-worker-%: build-worker-%
	cd build/$* && npx wrangler dev

deploy-worker-%: build-worker-%
	cd workers/$* && npx wrangler deploy

deploy-workers: $(addprefix deploy-worker-,$(WORKERS))
