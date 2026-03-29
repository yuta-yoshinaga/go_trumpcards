//go:build !js

package domain

// エクイティ計算のモンテカルロシミュレーション回数 (ネイティブビルド)。
// 50000回では Free tier サーバー (Render等) でリクエストタイムアウトが発生するため
// 2000回に削減。精度は±3%程度でゲームプレイには十分。
const (
	holdemEquitySimulations    = 2000
	omahaEquitySimulations     = 2000
	shortDeckEquitySimulations = 2000
)
