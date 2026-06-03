//go:build !js || !wasm || casino

package domain

// PineappleConfig パイナップルポーカー設定 (HoldemConfigと同一構造)
type PineappleConfig = HoldemConfig

// PineapplePlayStyle CPUプレイスタイル (HoldemPlayStyleと同一)
type PineapplePlayStyle = HoldemPlayStyle

// DefaultPineappleConfig デフォルト設定
func DefaultPineappleConfig() PineappleConfig {
	return DefaultHoldemConfig()
}

// NewPineapplePlayersForTable 指定されたテーブルサイズに応じたパイナップルプレイヤースライスを生成する
func NewPineapplePlayersForTable(tableSize int) []*PineapplePlayer {
	if !IsValidHoldemTableSize(tableSize) {
		tableSize = HoldemTableSize4
	}
	styles := DefaultCpuStyles(tableSize)
	players := make([]*PineapplePlayer, 0, tableSize)
	players = append(players, NewPineapplePlayer(true, HoldemStyleTAG))
	for _, s := range styles {
		players = append(players, NewPineapplePlayer(false, s))
	}
	return players
}
