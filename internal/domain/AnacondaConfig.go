//go:build !js || !wasm || extra4

package domain

import "errors"

// Anaconda の設定既定値・境界値。
const (
	// AnacondaDefaultPlayerCount はデフォルトのプレイヤー数 (人間 1 + CPU 3)。
	AnacondaDefaultPlayerCount = 4
	// AnacondaMinPlayerCount は最小プレイヤー数。
	AnacondaMinPlayerCount = 3
	// AnacondaMaxPlayerCount は最大プレイヤー数。
	AnacondaMaxPlayerCount = 7
	// AnacondaDefaultAnte はデフォルトのアンティ額 (ロール時の固定ベット増分も兼ねる)。
	AnacondaDefaultAnte = 10
	// AnacondaMinAnte は最小アンティ額。
	AnacondaMinAnte = 1
	// AnacondaMaxAnte は最大アンティ額。
	AnacondaMaxAnte = 1000
	// AnacondaDefaultStartingChips はデフォルトの初期チップ。
	AnacondaDefaultStartingChips = 200
	// AnacondaMinStartingChips は最小初期チップ。
	AnacondaMinStartingChips = 10
	// AnacondaMaxStartingChips は最大初期チップ (Validate 上限)。
	AnacondaMaxStartingChips = 100000
	// AnacondaDefaultTargetRounds はデフォルトの実施ラウンド数 (このラウンド数に達するとゲーム終了)。
	AnacondaDefaultTargetRounds = 10
	// AnacondaMinTargetRounds は最小ラウンド数。
	AnacondaMinTargetRounds = 1
	// AnacondaMaxTargetRounds は最大ラウンド数。
	AnacondaMaxTargetRounds = 100
)

// AnacondaConfig はアナコンダ (Anaconda / Pass the Trash) のローカルルール設定。
type AnacondaConfig struct {
	// PlayerCount は参加プレイヤー数 (人間 seat 0 + CPU、3〜7)。
	PlayerCount int `json:"pc"`
	// Ante は各ラウンドで全員が支払うアンティ額。ロールフェーズの固定ベット増分も兼ねる。
	Ante int `json:"an"`
	// StartingChips は各プレイヤーの初期チップ。
	StartingChips int `json:"sc"`
	// TargetRounds はこのラウンド数を消化するとゲームが終了する上限。
	TargetRounds int `json:"tr"`
}

// DefaultAnacondaConfig はデフォルトのローカルルール設定を返す。
func DefaultAnacondaConfig() AnacondaConfig {
	return AnacondaConfig{
		PlayerCount:   AnacondaDefaultPlayerCount,
		Ante:          AnacondaDefaultAnte,
		StartingChips: AnacondaDefaultStartingChips,
		TargetRounds:  AnacondaDefaultTargetRounds,
	}
}

// Validate は設定値のドメインバリデーションを行う。
func (c AnacondaConfig) Validate() error {
	if err := ValidateRange("player count", c.PlayerCount, AnacondaMinPlayerCount, AnacondaMaxPlayerCount); err != nil {
		return err
	}
	if err := ValidateRange("ante", c.Ante, AnacondaMinAnte, AnacondaMaxAnte); err != nil {
		return err
	}
	if err := ValidateRange("starting chips", c.StartingChips, AnacondaMinStartingChips, AnacondaMaxStartingChips); err != nil {
		return err
	}
	if c.StartingChips < c.Ante {
		return errors.New("starting chips must be greater than or equal to ante")
	}
	return ValidateRange("target rounds", c.TargetRounds, AnacondaMinTargetRounds, AnacondaMaxTargetRounds)
}
