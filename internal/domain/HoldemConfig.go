package domain

// HoldemPlayStyle CPUプレイスタイル
type HoldemPlayStyle int

const (
	HoldemStyleTAG HoldemPlayStyle = iota // Tight-Aggressive
	HoldemStyleLAP                        // Loose-Passive
	HoldemStyleTAP                        // Tight-Passive
	HoldemStyleLAG                        // Loose-Aggressive
)

// HoldemPlayStyleNames プレイスタイル名
var HoldemPlayStyleNames = []string{
	"TAG",
	"LAP",
	"TAP",
	"LAG",
}

// HoldemConfig テキサスホールデム設定
type HoldemConfig struct {
	SmallBlind      int  // スモールブラインド
	BigBlind        int  // ビッグブラインド
	InitChips       int  // 初期チップ
	TournamentMode  bool // トーナメントモード
	BlindLevelHands int  // ブラインドレベルアップまでのハンド数
	BlindMultiplier int  // ブラインド倍率 (百分率: 200=2倍)
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
	}
}
