//go:build !js || !wasm || solo

package interfaces

// BurracoGame はブラーコゲームのインタフェース。Burraco は「ポゼットを有効化した
// Canasta」として実装されるため、ドメイン型と同様にインタフェースも CanastaGame の
// 型エイリアスとして公開する（GetPozzettoCount を含む）。
type BurracoGame = CanastaGame
