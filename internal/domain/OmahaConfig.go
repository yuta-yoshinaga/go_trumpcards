package domain

// OmahaConfig オマハホールデム設定 (HoldemConfigと同一構造)
type OmahaConfig = HoldemConfig

// OmahaPlayStyle CPUプレイスタイル (HoldemPlayStyleと同一)
type OmahaPlayStyle = HoldemPlayStyle

// DefaultOmahaConfig デフォルト設定
func DefaultOmahaConfig() OmahaConfig {
	return DefaultHoldemConfig()
}

// NewOmahaPlayersForTable 指定されたテーブルサイズに応じたオマハプレイヤースライスを生成する
func NewOmahaPlayersForTable(tableSize int) []*OmahaPlayer {
	if !IsValidHoldemTableSize(tableSize) {
		tableSize = HoldemTableSize4
	}
	styles := DefaultCpuStyles(tableSize)
	players := make([]*OmahaPlayer, 0, tableSize)
	players = append(players, NewOmahaPlayer(true, HoldemStyleTAG))
	for _, s := range styles {
		players = append(players, NewOmahaPlayer(false, s))
	}
	return players
}
