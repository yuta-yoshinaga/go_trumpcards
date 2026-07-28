//go:build !js || !wasm || extra3

package domain

import "errors"

// Primero の設定既定値・境界値。
const (
	// PrimeroDefaultPlayerCount はデフォルトのプレイヤー数 (人間 1 + CPU 3)。
	PrimeroDefaultPlayerCount = 4
	// PrimeroMinPlayerCount は最小プレイヤー数。
	PrimeroMinPlayerCount = 2
	// PrimeroMaxPlayerCount は最大プレイヤー数。
	PrimeroMaxPlayerCount = 6
	// PrimeroDefaultAnte はデフォルトのアンティ額 (レイズ増分もこの額)。
	PrimeroDefaultAnte = 10
	// PrimeroMinAnte は最小アンティ額。
	PrimeroMinAnte = 1
	// PrimeroMaxAnte は最大アンティ額。
	PrimeroMaxAnte = 1000
	// PrimeroDefaultStartingChips はデフォルトの初期チップ。
	PrimeroDefaultStartingChips = 200
	// PrimeroMinStartingChips は最小初期チップ。
	PrimeroMinStartingChips = 10
	// PrimeroMaxStartingChips は最大初期チップ (Validate 上限)。
	PrimeroMaxStartingChips = 100000
	// PrimeroDefaultTargetRounds はデフォルトの実施ラウンド数 (このラウンド数に達するとゲーム終了)。
	PrimeroDefaultTargetRounds = 10
	// PrimeroMinTargetRounds は最小ラウンド数。
	PrimeroMinTargetRounds = 1
	// PrimeroMaxTargetRounds は最大ラウンド数。
	PrimeroMaxTargetRounds = 100
	// PrimeroMaxRaises は 1 ラウンドあたりの最大レイズ (ヴィ) 回数。
	PrimeroMaxRaises = 4
)

// PrimeroConfig はプリメロ (Primero) のローカルルール設定。
type PrimeroConfig struct {
	// PlayerCount は参加プレイヤー数 (人間 seat 0 + CPU、2〜6)。
	PlayerCount int `json:"pc"`
	// Ante は各ラウンドで全員が支払うアンティ額。レイズ (ヴィ) の増分もこの額に等しい。
	Ante int `json:"an"`
	// StartingChips は各プレイヤーの初期チップ。
	StartingChips int `json:"sc"`
	// TargetRounds はこのラウンド数を消化するとゲームが終了する上限。
	TargetRounds int `json:"tr"`
}

// DefaultPrimeroConfig はデフォルトのローカルルール設定を返す。
func DefaultPrimeroConfig() PrimeroConfig {
	return PrimeroConfig{
		PlayerCount:   PrimeroDefaultPlayerCount,
		Ante:          PrimeroDefaultAnte,
		StartingChips: PrimeroDefaultStartingChips,
		TargetRounds:  PrimeroDefaultTargetRounds,
	}
}

// Validate は設定値のドメインバリデーションを行う。
func (c PrimeroConfig) Validate() error {
	if err := ValidateRange("player count", c.PlayerCount, PrimeroMinPlayerCount, PrimeroMaxPlayerCount); err != nil {
		return err
	}
	if err := ValidateRange("ante", c.Ante, PrimeroMinAnte, PrimeroMaxAnte); err != nil {
		return err
	}
	if err := ValidateRange("starting chips", c.StartingChips, PrimeroMinStartingChips, PrimeroMaxStartingChips); err != nil {
		return err
	}
	if c.StartingChips < c.Ante {
		return errors.New("starting chips must be greater than or equal to ante")
	}
	return ValidateRange("target rounds", c.TargetRounds, PrimeroMinTargetRounds, PrimeroMaxTargetRounds)
}
