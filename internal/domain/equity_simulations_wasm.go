//go:build js && wasm

package domain

// エクイティ計算のモンテカルロシミュレーション回数 (Cloudflare Workers / TinyGo WASM)。
// ネイティブビルドでは2000回だが、Workers の CPU 時間制限 (10-30ms) では
// 重いモンテカルロシミュレーションがタイムアウトするため200回に抑える。
const (
	holdemEquitySimulations    = 200
	omahaEquitySimulations     = 200
	shortDeckEquitySimulations = 200
)
