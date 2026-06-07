//go:build !js || !wasm || casino

package domain

import "fmt"

// SevenCardStudPlayStyle CPUプレイスタイル (Holdemと共通)
type SevenCardStudPlayStyle = HoldemPlayStyle

// SevenCardStudPlayStyleNames プレイスタイル名 (Holdemと共通)
var SevenCardStudPlayStyleNames = HoldemPlayStyleNames

// テーブルサイズ定数
const (
	SevenCardStudTableSize4 = 4 // 4-max (1 human + 3 CPU)
	SevenCardStudTableSize6 = 6 // 6-max (1 human + 5 CPU)
	SevenCardStudTableSize7 = 7 // 7-max (1 human + 6 CPU)
)

// IsValidSevenCardStudTableSize テーブルサイズが有効か判定する
func IsValidSevenCardStudTableSize(n int) bool {
	return n >= 2 && n <= SevenCardStudTableSize7
}

// SevenCardStudConfig セブンカードスタッド設定
type SevenCardStudConfig struct {
	Ante             int              // アンティ
	BringIn          int              // ブリングイン
	SmallBet         int              // スモールベット (3rd-4th Street)
	BigBet           int              // ビッグベット (5th-7th Street)
	InitChips        int              // 初期チップ
	BettingLimit     BettingLimitType // ベッティングリミット
	TableSize        int              // テーブルサイズ (2-7)
	TournamentMode   bool             // トーナメントモード
	AnteLevelHands   int              // アンティレベルアップまでのハンド数
	AnteMultiplier   int              // アンティ倍率 (百分率: 200=2倍)
	RebuyEnabled     bool             // リバイ有効
	RebuyMaxCount    int              // リバイ最大回数
	RebuyChips       int              // リバイ時の補充チップ
	RebuyPeriodHands int              // リバイ可能期間 (ハンド数)
	AddonEnabled     bool             // アドオン有効
	AddonChips       int              // アドオン時の補充チップ
	AddonAfterHand   int              // アドオン提供ハンド番号
	CpuMetaAI        bool             // メタAI: セッション内学習
}

// DefaultSevenCardStudConfig デフォルト設定
func DefaultSevenCardStudConfig() SevenCardStudConfig {
	return SevenCardStudConfig{
		Ante:             1,
		BringIn:          2,
		SmallBet:         5,
		BigBet:           10,
		InitChips:        1000,
		BettingLimit:     BettingLimitFixed,
		TableSize:        SevenCardStudTableSize4,
		TournamentMode:   false,
		AnteLevelHands:   10,
		AnteMultiplier:   200,
		RebuyEnabled:     false,
		RebuyMaxCount:    3,
		RebuyChips:       1000,
		RebuyPeriodHands: 20,
		AddonEnabled:     false,
		AddonChips:       1500,
		AddonAfterHand:   20,
	}
}

// DefaultRazzConfig Razz デフォルト設定
func DefaultRazzConfig() SevenCardStudConfig {
	return DefaultSevenCardStudConfig()
}

// Validate 設定値のドメインバリデーション
func (c SevenCardStudConfig) Validate() error {
	if err := ValidateRange("betting limit", int(c.BettingLimit), int(BettingLimitFixed), int(BettingLimitNoLimit)); err != nil {
		return err
	}
	if err := ValidateMin("ante", c.Ante, 1); err != nil {
		return err
	}
	if err := ValidateMin("bring-in", c.BringIn, 1); err != nil {
		return err
	}
	if err := ValidateMin("small bet", c.SmallBet, 1); err != nil {
		return err
	}
	if err := ValidateMin("big bet", c.BigBet, 1); err != nil {
		return err
	}
	if c.SmallBet > c.BigBet {
		return fmt.Errorf("small bet (%d) must be <= big bet (%d)", c.SmallBet, c.BigBet)
	}
	if err := ValidateMin("ante level hands", c.AnteLevelHands, 1); err != nil {
		return err
	}
	if err := ValidateMin("init chips", c.InitChips, 1); err != nil {
		return err
	}
	// TableSize == 0 means "keep current size / no change"; only validate non-zero values.
	if c.TableSize != 0 && !IsValidSevenCardStudTableSize(c.TableSize) {
		return fmt.Errorf("invalid table size %d, must be 2-%d", c.TableSize, SevenCardStudTableSize7)
	}
	return nil
}

// SevenCardStudCpuStyles テーブルサイズ別CPUスタイル
var (
	sevenCardStudStyles3 = []SevenCardStudPlayStyle{HoldemStyleLAP, HoldemStyleTAP}
	sevenCardStudStyles4 = []SevenCardStudPlayStyle{HoldemStyleLAP, HoldemStyleTAP, HoldemStyleGTO}
	sevenCardStudStyles5 = []SevenCardStudPlayStyle{HoldemStyleTAG, HoldemStyleLAP, HoldemStyleTAP, HoldemStyleGTO}
	sevenCardStudStyles6 = []SevenCardStudPlayStyle{HoldemStyleTAG, HoldemStyleLAP, HoldemStyleTAP, HoldemStyleLAG, HoldemStyleGTO}
	sevenCardStudStyles7 = []SevenCardStudPlayStyle{HoldemStyleTAG, HoldemStyleLAP, HoldemStyleTAP, HoldemStyleLAG, HoldemStyleGTO, HoldemStyleTAG}
)

// DefaultSevenCardStudCpuStyles テーブルサイズに応じたCPUスタイルを返す
func DefaultSevenCardStudCpuStyles(tableSize int) []SevenCardStudPlayStyle {
	switch tableSize {
	case 3:
		return sevenCardStudStyles3
	case 4:
		return sevenCardStudStyles4
	case 5:
		return sevenCardStudStyles5
	case 6:
		return sevenCardStudStyles6
	case 7:
		return sevenCardStudStyles7
	default:
		if tableSize <= 2 {
			return []SevenCardStudPlayStyle{HoldemStyleGTO}
		}
		return sevenCardStudStyles7
	}
}

// NewSevenCardStudPlayersForTable 指定されたテーブルサイズに応じたプレイヤースライスを生成する
func NewSevenCardStudPlayersForTable(tableSize int) []*SevenCardStudPlayer {
	if !IsValidSevenCardStudTableSize(tableSize) {
		tableSize = SevenCardStudTableSize7
	}
	styles := DefaultSevenCardStudCpuStyles(tableSize)
	players := make([]*SevenCardStudPlayer, 0, tableSize)
	players = append(players, NewSevenCardStudPlayer(true, HoldemStyleTAG))
	for _, s := range styles {
		players = append(players, NewSevenCardStudPlayer(false, s))
	}
	return players
}
