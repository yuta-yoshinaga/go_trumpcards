//go:build !js || !wasm || extra

package interfaces

// BiribaGame はビリバゲームのインタフェース。Biriba は「ポゼットを有効化した
// Canasta」として実装されるため、ドメイン型と同様にインタフェースも CanastaGame の
// 型エイリアスとして公開する（GetPozzettoCount を含む）。
type BiribaGame = CanastaGame
