//go:build !js || !wasm || casino

package domain

// DramahaConfig ドラマハホールデム設定 (HoldemConfigと同一構造)
type DramahaConfig = HoldemConfig

// DramahaPlayStyle CPUプレイスタイル (HoldemPlayStyleと同一)
type DramahaPlayStyle = HoldemPlayStyle

// DefaultDramahaConfig デフォルト設定
func DefaultDramahaConfig() DramahaConfig {
	return DefaultHoldemConfig()
}

// NewDramahaPlayersForTable 指定されたテーブルサイズに応じたドラマハプレイヤースライスを生成する
func NewDramahaPlayersForTable(tableSize int) []*DramahaPlayer {
	if !IsValidHoldemTableSize(tableSize) {
		tableSize = HoldemTableSize4
	}
	styles := DefaultCpuStyles(tableSize)
	players := make([]*DramahaPlayer, 0, tableSize)
	players = append(players, NewDramahaPlayer(true, HoldemStyleTAG))
	for _, s := range styles {
		players = append(players, NewDramahaPlayer(false, s))
	}
	return players
}
