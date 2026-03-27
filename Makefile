TINYGO ?= tinygo
WASM_OPT ?= wasm-opt
ASSETS_GEN := go run github.com/syumai/workers/cmd/workers-assets-gen

.PHONY: build-workers build-worker-casino build-worker-classic build-worker-solo clean-workers deploy-workers

build-workers: build-worker-casino build-worker-classic build-worker-solo

define build_worker
	@echo "Building worker: $(1)"
	@mkdir -p build/$(1)
	$(ASSETS_GEN) -mode=tinygo -o build/$(1)
	$(TINYGO) build -o build/$(1)/app.wasm -target wasi -no-debug -opt=z ./cmd/workers/$(1)
	$(WASM_OPT) --enable-bulk-memory --enable-nontrapping-float-to-int --enable-sign-ext -Oz build/$(1)/app.wasm -o build/$(1)/app.wasm
	@RAW=$$(stat -c%s build/$(1)/app.wasm); GZIP=$$(gzip -c build/$(1)/app.wasm | wc -c); \
	echo "  $(1): $$RAW bytes raw, $$GZIP bytes gzip"
endef

build-worker-casino:
	$(call build_worker,casino)

build-worker-classic:
	$(call build_worker,classic)

build-worker-solo:
	$(call build_worker,solo)

clean-workers:
	rm -rf build/casino build/classic build/solo

deploy-workers: build-workers
	cd workers/casino && npx wrangler deploy
	cd workers/classic && npx wrangler deploy
	cd workers/solo && npx wrangler deploy
