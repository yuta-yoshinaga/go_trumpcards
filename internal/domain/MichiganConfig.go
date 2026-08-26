//go:build !js || !wasm || extra4

package domain

import "errors"

// Michigan の設定既定値・境界値。
const (
	// MichiganDefaultPlayerCount はデフォルトのプレイヤー数 (人間 1 + CPU 3)。
	MichiganDefaultPlayerCount = 4
	// MichiganMinPlayerCount は最小プレイヤー数。
	MichiganMinPlayerCount = 3
	// MichiganMaxPlayerCount は最大プレイヤー数。
	MichiganMaxPlayerCount = 8
	// MichiganDefaultAnte はデフォルトのアンティ額 (4 つのブードルに分配して賭ける)。
	MichiganDefaultAnte = 8
	// MichiganMinAnte は最小アンティ額 (ブードル数と一致させると均等分配できる)。
	MichiganMinAnte = 4
	// MichiganMaxAnte は最大アンティ額。
	MichiganMaxAnte = 1000
	// MichiganDefaultStartingChips はデフォルトの初期チップ。
	MichiganDefaultStartingChips = 200
	// MichiganMinStartingChips は最小初期チップ。
	MichiganMinStartingChips = 10
	// MichiganMaxStartingChips は最大初期チップ (Validate 上限)。
	MichiganMaxStartingChips = 100000
	// MichiganDefaultTargetRounds はデフォルトの実施ラウンド数 (このラウンド数に達するとゲーム終了)。
	MichiganDefaultTargetRounds = 10
	// MichiganMinTargetRounds は最小ラウンド数。
	MichiganMinTargetRounds = 1
	// MichiganMaxTargetRounds は最大ラウンド数。
	MichiganMaxTargetRounds = 100
	// MichiganBoodleCount はブードル (中央の賭け札) の枚数 (常に 4)。
	MichiganBoodleCount = 4
)

// MichiganConfig はミシガン (Michigan / Newmarket) のローカルルール設定。
type MichiganConfig struct {
	// PlayerCount は参加プレイヤー数 (人間 seat 0 + CPU、3〜8)。
	PlayerCount int `json:"pc"`
	// Ante は各ラウンドで全員が 4 つのブードルに分配して賭ける総額。
	Ante int `json:"an"`
	// StartingChips は各プレイヤーの初期チップ。
	StartingChips int `json:"sc"`
	// TargetRounds はこのラウンド数を消化するとゲームが終了する上限。
	TargetRounds int `json:"tr"`
}

// DefaultMichiganConfig はデフォルトのローカルルール設定を返す。
func DefaultMichiganConfig() MichiganConfig {
	return MichiganConfig{
		PlayerCount:   MichiganDefaultPlayerCount,
		Ante:          MichiganDefaultAnte,
		StartingChips: MichiganDefaultStartingChips,
		TargetRounds:  MichiganDefaultTargetRounds,
	}
}

// Validate は設定値のドメインバリデーションを行う。
func (c MichiganConfig) Validate() error {
	if err := ValidateRange("player count", c.PlayerCount, MichiganMinPlayerCount, MichiganMaxPlayerCount); err != nil {
		return err
	}
	if err := ValidateRange("ante", c.Ante, MichiganMinAnte, MichiganMaxAnte); err != nil {
		return err
	}
	if err := ValidateRange("starting chips", c.StartingChips, MichiganMinStartingChips, MichiganMaxStartingChips); err != nil {
		return err
	}
	if c.StartingChips < c.Ante {
		return errors.New("starting chips must be greater than or equal to ante")
	}
	return ValidateRange("target rounds", c.TargetRounds, MichiganMinTargetRounds, MichiganMaxTargetRounds)
}
