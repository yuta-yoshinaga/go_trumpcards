//go:build !js || !wasm || casino

package domain

import "fmt"

// FollowTheQueenPlayStyle CPUプレイスタイル (Holdemと共通)
type FollowTheQueenPlayStyle = HoldemPlayStyle

// FollowTheQueenPlayStyleNames プレイスタイル名 (Holdemと共通)
var FollowTheQueenPlayStyleNames = HoldemPlayStyleNames

// テーブルサイズ定数
const (
	FollowTheQueenTableSize4 = 4 // 4-max (1 human + 3 CPU)
	FollowTheQueenTableSize6 = 6 // 6-max (1 human + 5 CPU)
	FollowTheQueenTableSize7 = 7 // 7-max (1 human + 6 CPU)
)

// IsValidFollowTheQueenTableSize テーブルサイズが有効か判定する
func IsValidFollowTheQueenTableSize(n int) bool {
	return n >= 2 && n <= FollowTheQueenTableSize7
}

// FollowTheQueenConfig フォロー・ザ・クイーン設定
type FollowTheQueenConfig struct {
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

// DefaultFollowTheQueenConfig デフォルト設定
func DefaultFollowTheQueenConfig() FollowTheQueenConfig {
	return FollowTheQueenConfig{
		Ante:             1,
		BringIn:          2,
		SmallBet:         5,
		BigBet:           10,
		InitChips:        1000,
		BettingLimit:     BettingLimitFixed,
		TableSize:        FollowTheQueenTableSize4,
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

// Validate 設定値のドメインバリデーション
func (c FollowTheQueenConfig) Validate() error {
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
	if c.TableSize != 0 && !IsValidFollowTheQueenTableSize(c.TableSize) {
		return fmt.Errorf("invalid table size %d, must be 2-%d", c.TableSize, FollowTheQueenTableSize7)
	}
	return nil
}

// FollowTheQueenCpuStyles テーブルサイズ別CPUスタイル
var (
	followTheQueenStyles3 = []FollowTheQueenPlayStyle{HoldemStyleLAP, HoldemStyleTAP}
	followTheQueenStyles4 = []FollowTheQueenPlayStyle{HoldemStyleLAP, HoldemStyleTAP, HoldemStyleGTO}
	followTheQueenStyles5 = []FollowTheQueenPlayStyle{HoldemStyleTAG, HoldemStyleLAP, HoldemStyleTAP, HoldemStyleGTO}
	followTheQueenStyles6 = []FollowTheQueenPlayStyle{HoldemStyleTAG, HoldemStyleLAP, HoldemStyleTAP, HoldemStyleLAG, HoldemStyleGTO}
	followTheQueenStyles7 = []FollowTheQueenPlayStyle{HoldemStyleTAG, HoldemStyleLAP, HoldemStyleTAP, HoldemStyleLAG, HoldemStyleGTO, HoldemStyleTAG}
)

// DefaultFollowTheQueenCpuStyles テーブルサイズに応じたCPUスタイルを返す
func DefaultFollowTheQueenCpuStyles(tableSize int) []FollowTheQueenPlayStyle {
	switch tableSize {
	case 3:
		return followTheQueenStyles3
	case 4:
		return followTheQueenStyles4
	case 5:
		return followTheQueenStyles5
	case 6:
		return followTheQueenStyles6
	case 7:
		return followTheQueenStyles7
	default:
		if tableSize <= 2 {
			return []FollowTheQueenPlayStyle{HoldemStyleGTO}
		}
		return followTheQueenStyles7
	}
}

// NewFollowTheQueenPlayersForTable 指定されたテーブルサイズに応じたプレイヤースライスを生成する
func NewFollowTheQueenPlayersForTable(tableSize int) []*FollowTheQueenPlayer {
	if !IsValidFollowTheQueenTableSize(tableSize) {
		tableSize = FollowTheQueenTableSize7
	}
	styles := DefaultFollowTheQueenCpuStyles(tableSize)
	players := make([]*FollowTheQueenPlayer, 0, tableSize)
	players = append(players, NewFollowTheQueenPlayer(true, HoldemStyleTAG))
	for _, s := range styles {
		players = append(players, NewFollowTheQueenPlayer(false, s))
	}
	return players
}
