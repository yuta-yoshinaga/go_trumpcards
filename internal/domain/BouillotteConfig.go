//go:build !js || !wasm || extra3

package domain

import "errors"

// Bouillotte の設定既定値・境界値。
const (
	// BouillotteDefaultPlayerCount はデフォルトのプレイヤー数 (人間 1 + CPU 3)。
	BouillotteDefaultPlayerCount = 4
	// BouillotteMinPlayerCount は最小プレイヤー数。
	BouillotteMinPlayerCount = 3
	// BouillotteMaxPlayerCount は最大プレイヤー数。
	BouillotteMaxPlayerCount = 4
	// BouillotteDefaultAnte はデフォルトのアンティ額 (レイズ増分もこの額)。
	BouillotteDefaultAnte = 10
	// BouillotteMinAnte は最小アンティ額。
	BouillotteMinAnte = 1
	// BouillotteMaxAnte は最大アンティ額。
	BouillotteMaxAnte = 1000
	// BouillotteDefaultStartingChips はデフォルトの初期チップ。
	BouillotteDefaultStartingChips = 200
	// BouillotteMinStartingChips は最小初期チップ。
	BouillotteMinStartingChips = 10
	// BouillotteMaxStartingChips は最大初期チップ (Validate 上限)。
	BouillotteMaxStartingChips = 100000
	// BouillotteDefaultTargetRounds はデフォルトの実施ラウンド数 (このラウンド数に達するとゲーム終了)。
	BouillotteDefaultTargetRounds = 10
	// BouillotteMinTargetRounds は最小ラウンド数。
	BouillotteMinTargetRounds = 1
	// BouillotteMaxTargetRounds は最大ラウンド数。
	BouillotteMaxTargetRounds = 100
	// BouillotteMaxRaises は 1 ラウンドあたりの最大レイズ (ヴィ) 回数。
	BouillotteMaxRaises = 4
)

// BouillotteConfig はブイヨット (Bouillotte) のローカルルール設定。
type BouillotteConfig struct {
	// PlayerCount は参加プレイヤー数 (人間 seat 0 + CPU、3〜4)。
	PlayerCount int `json:"pc"`
	// Ante は各ラウンドで全員が支払うアンティ額。レイズ (ヴィ) の増分もこの額に等しい。
	Ante int `json:"an"`
	// StartingChips は各プレイヤーの初期チップ。
	StartingChips int `json:"sc"`
	// TargetRounds はこのラウンド数を消化するとゲームが終了する上限。
	TargetRounds int `json:"tr"`
}

// DefaultBouillotteConfig はデフォルトのローカルルール設定を返す。
func DefaultBouillotteConfig() BouillotteConfig {
	return BouillotteConfig{
		PlayerCount:   BouillotteDefaultPlayerCount,
		Ante:          BouillotteDefaultAnte,
		StartingChips: BouillotteDefaultStartingChips,
		TargetRounds:  BouillotteDefaultTargetRounds,
	}
}

// Validate は設定値のドメインバリデーションを行う。
func (c BouillotteConfig) Validate() error {
	if err := ValidateRange("player count", c.PlayerCount, BouillotteMinPlayerCount, BouillotteMaxPlayerCount); err != nil {
		return err
	}
	if err := ValidateRange("ante", c.Ante, BouillotteMinAnte, BouillotteMaxAnte); err != nil {
		return err
	}
	if err := ValidateRange("starting chips", c.StartingChips, BouillotteMinStartingChips, BouillotteMaxStartingChips); err != nil {
		return err
	}
	if c.StartingChips < c.Ante {
		return errors.New("starting chips must be greater than or equal to ante")
	}
	return ValidateRange("target rounds", c.TargetRounds, BouillotteMinTargetRounds, BouillotteMaxTargetRounds)
}
