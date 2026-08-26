//go:build !js || !wasm || extra4

package domain

import "errors"

// SevenTwentySeven の設定既定値・境界値。
const (
	// SevenTwentySevenDefaultPlayerCount はデフォルトのプレイヤー数 (人間 1 + CPU 3)。
	SevenTwentySevenDefaultPlayerCount = 4
	// SevenTwentySevenMinPlayerCount は最小プレイヤー数。
	SevenTwentySevenMinPlayerCount = 2
	// SevenTwentySevenMaxPlayerCount は最大プレイヤー数。
	SevenTwentySevenMaxPlayerCount = 7
	// SevenTwentySevenDefaultAnte はデフォルトのアンティ額。
	SevenTwentySevenDefaultAnte = 10
	// SevenTwentySevenMinAnte は最小アンティ額。
	SevenTwentySevenMinAnte = 1
	// SevenTwentySevenMaxAnte は最大アンティ額。
	SevenTwentySevenMaxAnte = 1000
	// SevenTwentySevenDefaultStartingChips はデフォルトの初期チップ。
	SevenTwentySevenDefaultStartingChips = 200
	// SevenTwentySevenMinStartingChips は最小初期チップ。
	SevenTwentySevenMinStartingChips = 10
	// SevenTwentySevenMaxStartingChips は最大初期チップ (Validate 上限)。
	SevenTwentySevenMaxStartingChips = 100000
	// SevenTwentySevenDefaultTargetRounds はデフォルトの実施ラウンド数 (このラウンド数に達するとゲーム終了)。
	SevenTwentySevenDefaultTargetRounds = 10
	// SevenTwentySevenMinTargetRounds は最小ラウンド数。
	SevenTwentySevenMinTargetRounds = 1
	// SevenTwentySevenMaxTargetRounds は最大ラウンド数。
	SevenTwentySevenMaxTargetRounds = 100
)

// SevenTwentySevenConfig はセブン・トゥエンティセブン (SevenTwentySeven) のローカルルール設定。
type SevenTwentySevenConfig struct {
	// PlayerCount は参加プレイヤー数 (人間 seat 0 + CPU、2〜7)。
	PlayerCount int `json:"pc"`
	// Ante は各ラウンドで全員が支払うアンティ額。
	Ante int `json:"an"`
	// StartingChips は各プレイヤーの初期チップ。
	StartingChips int `json:"sc"`
	// TargetRounds はこのラウンド数を消化するとゲームが終了する上限 (エスカレーションの停止条件)。
	TargetRounds int `json:"tr"`
}

// DefaultSevenTwentySevenConfig はデフォルトのローカルルール設定を返す。
func DefaultSevenTwentySevenConfig() SevenTwentySevenConfig {
	return SevenTwentySevenConfig{
		PlayerCount:   SevenTwentySevenDefaultPlayerCount,
		Ante:          SevenTwentySevenDefaultAnte,
		StartingChips: SevenTwentySevenDefaultStartingChips,
		TargetRounds:  SevenTwentySevenDefaultTargetRounds,
	}
}

// Validate は設定値のドメインバリデーションを行う。
func (c SevenTwentySevenConfig) Validate() error {
	if err := ValidateRange("player count", c.PlayerCount, SevenTwentySevenMinPlayerCount, SevenTwentySevenMaxPlayerCount); err != nil {
		return err
	}
	if err := ValidateRange("ante", c.Ante, SevenTwentySevenMinAnte, SevenTwentySevenMaxAnte); err != nil {
		return err
	}
	if err := ValidateRange("starting chips", c.StartingChips, SevenTwentySevenMinStartingChips, SevenTwentySevenMaxStartingChips); err != nil {
		return err
	}
	if c.StartingChips < c.Ante {
		return errors.New("starting chips must be greater than or equal to ante")
	}
	return ValidateRange("target rounds", c.TargetRounds, SevenTwentySevenMinTargetRounds, SevenTwentySevenMaxTargetRounds)
}
