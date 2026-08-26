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
	// **ドラマハは 4-max 固定。** 1 席あたり最悪 10 枚 (ホール 5 + 交換 5) 使い、
	// ボードに 5 枚要るので、必要枚数は 10N+5。52 枚に収まるのは N=4 まで
	// (6-max は 65 枚、9-max は 95 枚)。足りないまま配ると山が尽き、DrawCard が
	// nil を返し、ボードが 3 枚のままショーダウンに行く —— 実測で 9-max の
	// 85% が 3 枚ボード、15% が panic。クローン元 (Three Card Brag / Omaha) は
	// 1 席 4 枚以下なので、この上限を持っていない。
	if !IsValidHoldemTableSize(tableSize) || tableSize != HoldemTableSize4 {
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
