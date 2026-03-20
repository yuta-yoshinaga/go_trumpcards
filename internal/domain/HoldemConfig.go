package domain

import "fmt"

// HoldemPlayStyle CPUプレイスタイル
type HoldemPlayStyle int

// Holdemのプレイスタイル定数
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
	SmallBlind       int              // スモールブラインド
	BigBlind         int              // ビッグブラインド
	InitChips        int              // 初期チップ
	TournamentMode   bool             // トーナメントモード
	BlindLevelHands  int              // ブラインドレベルアップまでのハンド数
	BlindMultiplier  int              // ブラインド倍率 (百分率: 200=2倍)
	BettingLimit     BettingLimitType // ベッティングリミット
	TableSize        int              // テーブルサイズ (4/6/9)
	RebuyEnabled     bool             // リバイ有効
	RebuyMaxCount    int              // リバイ最大回数
	RebuyChips       int              // リバイ時の補充チップ
	RebuyPeriodHands int              // リバイ可能期間 (ハンド数)
	AddonEnabled     bool             // アドオン有効
	AddonChips       int              // アドオン時の補充チップ
	AddonAfterHand   int              // アドオン提供ハンド番号
}

// DefaultHoldemConfig デフォルト設定
func DefaultHoldemConfig() HoldemConfig {
	return HoldemConfig{
		SmallBlind:       5,
		BigBlind:         10,
		InitChips:        1000,
		TournamentMode:   false,
		BlindLevelHands:  10,
		BlindMultiplier:  200,
		TableSize:        HoldemTableSize4,
		RebuyEnabled:     false,
		RebuyMaxCount:    3,
		RebuyChips:       1000,
		RebuyPeriodHands: 20,
		AddonEnabled:     false,
		AddonChips:       1500,
		AddonAfterHand:   20,
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

// Validate 設定値のドメインバリデーション
func (c HoldemConfig) Validate() error {
	if c.BettingLimit < BettingLimitFixed || c.BettingLimit > BettingLimitNoLimit {
		return fmt.Errorf("betting limit must be %d-%d, got %d", int(BettingLimitFixed), int(BettingLimitNoLimit), int(c.BettingLimit))
	}
	if c.SmallBlind < 1 {
		return fmt.Errorf("small blind must be >= 1, got %d", c.SmallBlind)
	}
	if c.BigBlind < 2 {
		return fmt.Errorf("big blind must be >= 2, got %d", c.BigBlind)
	}
	if c.SmallBlind >= c.BigBlind {
		return fmt.Errorf("small blind (%d) must be less than big blind (%d)", c.SmallBlind, c.BigBlind)
	}
	if c.BlindLevelHands < 1 {
		return fmt.Errorf("blind level hands must be >= 1, got %d", c.BlindLevelHands)
	}
	// TableSize == 0 means "keep current size / no change"; only validate non-zero values.
	if c.TableSize != 0 && !IsValidHoldemTableSize(c.TableSize) {
		return fmt.Errorf("invalid table size %d, must be %d, %d, or %d", c.TableSize, HoldemTableSize4, HoldemTableSize6, HoldemTableSize9)
	}
	return nil
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
