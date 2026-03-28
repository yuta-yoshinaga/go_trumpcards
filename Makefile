TINYGO ?= tinygo
WASM_OPT ?= wasm-opt
ASSETS_GEN := go run github.com/syumai/workers/cmd/workers-assets-gen

.PHONY: build-workers build-worker-casino build-worker-classic build-worker-solo clean-workers deploy-workers

build-workers: build-worker-casino build-worker-classic build-worker-solo

define build_worker
	@echo "Building worker: $(1)"
	@mkdir -p workers/$(1)/build
	$(ASSETS_GEN) -mode=tinygo -o workers/$(1)/build
	$(TINYGO) build -o workers/$(1)/build/app.wasm -target wasi -no-debug -opt=z ./cmd/workers/$(1)
	$(WASM_OPT) --enable-bulk-memory --enable-nontrapping-float-to-int --enable-sign-ext -Oz workers/$(1)/build/app.wasm -o workers/$(1)/build/app.wasm
	@RAW=$$(stat -c%s workers/$(1)/build/app.wasm); GZIP=$$(gzip -c workers/$(1)/build/app.wasm | wc -c); \
	echo "  $(1): $$RAW bytes raw, $$GZIP bytes gzip"
endef

build-worker-casino:
	$(call build_worker,casino)

build-worker-classic:
	$(call build_worker,classic)

build-worker-solo:
	$(call build_worker,solo)

clean-workers:
	rm -rf workers/casino/build workers/classic/build workers/solo/build

deploy-workers: build-workers
	cd workers/casino && bunx wrangler deploy
	cd workers/classic && bunx wrangler deploy
	cd workers/solo && bunx wrangler deploy
