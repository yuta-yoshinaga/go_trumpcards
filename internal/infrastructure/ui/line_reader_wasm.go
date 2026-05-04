//go:build js && wasm

package ui

import "os"

// newDefaultLineReader returns a Scanner-backed reader for WASM builds.
// The Cloudflare Workers don't import this package today (they expose only
// the HTTP API), but the build tag and stub are kept so any future code path
// that pulls in `ui` from a worker doesn't break the TinyGo compile — liner
// has no cgo, but several of its dependencies don't compile under TinyGo's
// js+wasm target.
func newDefaultLineReader() LineReader {
	return newScannerLineReader(os.Stdin, os.Stdout)
}
