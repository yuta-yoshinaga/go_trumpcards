//go:build !js || !wasm || casino

package domain

import "fmt"

// FiveCardStudPlayStyle CPUプレイスタイル (Holdemと共通)
type FiveCardStudPlayStyle = HoldemPlayStyle

// FiveCardStudPlayStyleNames プレイスタイル名 (Holdemと共通)
var FiveCardStudPlayStyleNames = HoldemPlayStyleNames

// テーブルサイズ定数
const (
	FiveCardStudTableSize4 = 4 // 4-max (1 human + 3 CPU)
	FiveCardStudTableSize6 = 6 // 6-max (1 human + 5 CPU)
)

// IsValidFiveCardStudTableSize テーブルサイズが有効か判定する
func IsValidFiveCardStudTableSize(n int) bool {
	return n >= 2 && n <= FiveCardStudTableSize6
}

// FiveCardStudConfig ファイブカードスタッド設定
type FiveCardStudConfig struct {
	Ante             int              // アンティ
	BringIn          int              // ブリングイン
	SmallBet         int              // スモールベット (2nd-3rd Street)
	BigBet           int              // ビッグベット (4th-5th Street)
	InitChips        int              // 初期チップ
	BettingLimit     BettingLimitType // ベッティングリミット
	TableSize        int              // テーブルサイズ (2-6)
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

// DefaultFiveCardStudConfig デフォルト設定 (6-max)
func DefaultFiveCardStudConfig() FiveCardStudConfig {
	return FiveCardStudConfig{
		Ante:             1,
		BringIn:          2,
		SmallBet:         5,
		BigBet:           10,
		InitChips:        1000,
		BettingLimit:     BettingLimitFixed,
		TableSize:        FiveCardStudTableSize6,
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

// DefaultSokoConfig Soko (Canadian Stud) デフォルト設定。
// ディール・ブリングイン・ベット構造は Five Card Stud と同一で、違うのは
// ショウダウンの役序列だけなので設定値も同じ。
func DefaultSokoConfig() FiveCardStudConfig {
	return DefaultFiveCardStudConfig()
}

// Validate 設定値のドメインバリデーション
func (c FiveCardStudConfig) Validate() error {
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
	if c.TableSize != 0 && !IsValidFiveCardStudTableSize(c.TableSize) {
		return fmt.Errorf("invalid table size %d, must be 2-%d", c.TableSize, FiveCardStudTableSize6)
	}
	return nil
}

// FiveCardStudCpuStyles テーブルサイズ別CPUスタイル
var (
	fiveCardStudStyles3 = []FiveCardStudPlayStyle{HoldemStyleLAP, HoldemStyleTAP}
	fiveCardStudStyles4 = []FiveCardStudPlayStyle{HoldemStyleLAP, HoldemStyleTAP, HoldemStyleGTO}
	fiveCardStudStyles5 = []FiveCardStudPlayStyle{HoldemStyleTAG, HoldemStyleLAP, HoldemStyleTAP, HoldemStyleGTO}
	fiveCardStudStyles6 = []FiveCardStudPlayStyle{HoldemStyleTAG, HoldemStyleLAP, HoldemStyleTAP, HoldemStyleLAG, HoldemStyleGTO}
)

// DefaultFiveCardStudCpuStyles テーブルサイズに応じたCPUスタイルを返す
func DefaultFiveCardStudCpuStyles(tableSize int) []FiveCardStudPlayStyle {
	switch tableSize {
	case 3:
		return fiveCardStudStyles3
	case 4:
		return fiveCardStudStyles4
	case 5:
		return fiveCardStudStyles5
	case 6:
		return fiveCardStudStyles6
	default:
		if tableSize <= 2 {
			return []FiveCardStudPlayStyle{HoldemStyleGTO}
		}
		return fiveCardStudStyles6
	}
}

// NewFiveCardStudPlayersForTable 指定されたテーブルサイズに応じたプレイヤースライスを生成する
func NewFiveCardStudPlayersForTable(tableSize int) []*FiveCardStudPlayer {
	if !IsValidFiveCardStudTableSize(tableSize) {
		tableSize = FiveCardStudTableSize6
	}
	styles := DefaultFiveCardStudCpuStyles(tableSize)
	players := make([]*FiveCardStudPlayer, 0, tableSize)
	players = append(players, NewFiveCardStudPlayer(true, HoldemStyleTAG))
	for _, s := range styles {
		players = append(players, NewFiveCardStudPlayer(false, s))
	}
	return players
}
