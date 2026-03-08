package domain

// HoldemPlayStyle CPUプレイスタイル
type HoldemPlayStyle int

const (
	HoldemStyleTAG HoldemPlayStyle = iota // Tight-Aggressive
	HoldemStyleLAP                        // Loose-Passive
	HoldemStyleTAP                        // Tight-Passive
	HoldemStyleLAG                        // Loose-Aggressive
	HoldemStyleGTO                        // Game Theory Optimal
)

// HoldemPlayStyleNames プレイスタイル名
var HoldemPlayStyleNames = []string{
	"TAG",
	"LAP",
	"TAP",
	"LAG",
	"GTO",
}

// テーブルサイズ定数
const (
	HoldemTableSize4 = 4 // 4-max (1 human + 3 CPU)
	HoldemTableSize6 = 6 // 6-max (1 human + 5 CPU)
	HoldemTableSize9 = 9 // 9-max (1 human + 8 CPU)
)

// ValidHoldemTableSizes 有効なテーブルサイズ
var ValidHoldemTableSizes = map[int]bool{
	HoldemTableSize4: true,
	HoldemTableSize6: true,
	HoldemTableSize9: true,
}

// HoldemConfig テキサスホールデム設定
type HoldemConfig struct {
	SmallBlind      int              // スモールブラインド
	BigBlind        int              // ビッグブラインド
	InitChips       int              // 初期チップ
	TournamentMode  bool             // トーナメントモード
	BlindLevelHands int              // ブラインドレベルアップまでのハンド数
	BlindMultiplier int              // ブラインド倍率 (百分率: 200=2倍)
	BettingLimit    BettingLimitType // ベッティングリミット
	TableSize       int              // テーブルサイズ (4/6/9)
}

// DefaultHoldemConfig デフォルト設定
func DefaultHoldemConfig() HoldemConfig {
	return HoldemConfig{
		SmallBlind:      5,
		BigBlind:        10,
		InitChips:       1000,
		TournamentMode:  false,
		BlindLevelHands: 10,
		BlindMultiplier: 200,
		TableSize:       HoldemTableSize4,
	}
}

// DefaultCpuStyles テーブルサイズに応じたCPUスタイルを返す
func DefaultCpuStyles(tableSize int) []HoldemPlayStyle {
	switch tableSize {
	case HoldemTableSize6:
		return []HoldemPlayStyle{HoldemStyleTAG, HoldemStyleLAP, HoldemStyleTAP, HoldemStyleLAG, HoldemStyleGTO}
	case HoldemTableSize9:
		return []HoldemPlayStyle{HoldemStyleTAG, HoldemStyleLAP, HoldemStyleTAP, HoldemStyleLAG, HoldemStyleGTO, HoldemStyleTAG, HoldemStyleLAP, HoldemStyleLAG}
	default:
		return []HoldemPlayStyle{HoldemStyleLAP, HoldemStyleTAP, HoldemStyleGTO}
	}
}
