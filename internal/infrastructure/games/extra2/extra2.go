//go:build js && wasm

// Package extra2 binds the Cloudflare Worker KV-backed handlers for the games
// assigned to the fifth size bucket. A worker main blank-imports this package so
// the init below runs before games.RegisterCategory is called.
//
// Like casino/classic/solo/extra this is purely a binary-size bucket, not a
// user-facing taxonomy (ADR-0036). The colourless name is deliberate: it holds
// whatever had to move to keep every TinyGo WASM binary under the Cloudflare
// Workers free-tier 1 MB gzipped limit, and nothing about a game's genre says
// it belongs here.
//
// Currently empty. Phase 1 of ADR-0036 adds the bucket and proves the build and
// deploy path; Phase 2 moves games in.
package extra2
