TINYGO ?= tinygo
WASM_OPT ?= wasm-opt
ASSETS_GEN := go run github.com/syumai/workers/cmd/workers-assets-gen
COVERAGE_DIR := build/coverage

.PHONY: build-workers build-worker-casino build-worker-classic build-worker-solo clean-workers deploy-workers coverage clean-cov

build-workers: build-worker-casino build-worker-classic build-worker-solo

define build_worker
	@echo "Building worker: $(1)"
	@mkdir -p workers/$(1)/build
	$(ASSETS_GEN) -mode=tinygo -o workers/$(1)/build
	# -stack-size=64KB: TinyGo's default 16KB main stack overflows when
	# encoding/json walks the deep MarshalJSON chain produced by the
	# Player → GamePlayer → RankedGamePlayer → SevensPlayer (etc.) embedded
	# pointer hierarchy. 64KB gives ~4× headroom and fits within the
	# Cloudflare Workers free-tier 1MB gzipped limit. See PR #1640 follow-up.
	$(TINYGO) build -o workers/$(1)/build/app.wasm -target wasm -stack-size=64KB -no-debug -opt=z ./cmd/workers/$(1)
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

coverage: ## Run tests with coverage, writing profile and HTML report to build/coverage/.
	@mkdir -p $(COVERAGE_DIR)
	go test -tags test -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic ./...
	go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "Coverage written to $(COVERAGE_DIR)/"

clean-cov: ## Remove coverage profile artifacts (root-level *.out, coverage.html, build/coverage).
	@find . -maxdepth 1 -type f \( -name '*.out' -o -name 'coverage.html' \) -delete
	@rm -rf $(COVERAGE_DIR)
	@echo "Removed coverage artifacts"
