//go:build js && wasm

// Package extra binds the Cloudflare Worker KV-backed handlers for the games
// assigned to the fourth ("extra") size bucket. A worker main blank-imports
// this package so the init below runs before games.RegisterCategory is called.
//
// Like casino/classic/solo this is purely a binary-size bucket, not a
// user-facing taxonomy: it holds an overflow mix of games moved off the other
// three workers to keep every TinyGo WASM binary under the Cloudflare Workers
// free-tier 1 MB gzipped limit. Game RegisterKVGame calls are appended to the
// init below as games are rebucketed here.
package extra

func init() {
}
