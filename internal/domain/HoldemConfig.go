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

// IsValidHoldemTableSize テーブルサイズが有効か判定する
func IsValidHoldemTableSize(n int) bool {
	return n == HoldemTableSize4 || n == HoldemTableSize6 || n == HoldemTableSize9
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

// テーブルサイズ別CPUスタイル
var (
	cpuStyles4Max = []HoldemPlayStyle{HoldemStyleLAP, HoldemStyleTAP, HoldemStyleGTO}
	cpuStyles6Max = []HoldemPlayStyle{HoldemStyleTAG, HoldemStyleLAP, HoldemStyleTAP, HoldemStyleLAG, HoldemStyleGTO}
	cpuStyles9Max = []HoldemPlayStyle{HoldemStyleTAG, HoldemStyleLAP, HoldemStyleTAP, HoldemStyleLAG, HoldemStyleGTO, HoldemStyleTAG, HoldemStyleLAP, HoldemStyleLAG}
)

// DefaultCpuStyles テーブルサイズに応じたCPUスタイルを返す
func DefaultCpuStyles(tableSize int) []HoldemPlayStyle {
	switch tableSize {
	case HoldemTableSize6:
		return cpuStyles6Max
	case HoldemTableSize9:
		return cpuStyles9Max
	default:
		return cpuStyles4Max
	}
}

// NewPlayersForTable 指定されたテーブルサイズに応じたプレイヤースライスを生成する
func NewPlayersForTable(tableSize int) []*HoldemPlayer {
	if !IsValidHoldemTableSize(tableSize) {
		tableSize = HoldemTableSize4
	}
	styles := DefaultCpuStyles(tableSize)
	players := make([]*HoldemPlayer, 0, tableSize)
	players = append(players, NewHoldemPlayer(true, HoldemStyleTAG))
	for _, s := range styles {
		players = append(players, NewHoldemPlayer(false, s))
	}
	return players
}
