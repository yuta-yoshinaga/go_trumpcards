//go:build !js || !wasm || casino

package domain

// ShortDeckConfig ショートデックホールデム設定 (HoldemConfigと同一構造)
type ShortDeckConfig = HoldemConfig

// ShortDeckPlayStyle CPUプレイスタイル (HoldemPlayStyleと同一)
type ShortDeckPlayStyle = HoldemPlayStyle

// DefaultShortDeckConfig デフォルト設定
func DefaultShortDeckConfig() ShortDeckConfig {
	return DefaultHoldemConfig()
}

// NewShortDeckPlayersForTable 指定されたテーブルサイズに応じたショートデックプレイヤースライスを生成する
func NewShortDeckPlayersForTable(tableSize int) []*ShortDeckPlayer {
	if !IsValidHoldemTableSize(tableSize) {
		tableSize = HoldemTableSize4
	}
	styles := DefaultCpuStyles(tableSize)
	players := make([]*ShortDeckPlayer, 0, tableSize)
	players = append(players, NewShortDeckPlayer(true, HoldemStyleTAG))
	for _, s := range styles {
		players = append(players, NewShortDeckPlayer(false, s))
	}
	return players
}
