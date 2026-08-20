//go:build !js || !wasm || extra4

package domain

import "errors"

// GutsDeclaration はガッツ (Guts) の宣言種別。各プレイヤーはラウンドごとに手札を見て
// 「イン (勝負に残る)」か「アウト (降りる)」かを同時に宣言する。
type GutsDeclaration int

// Guts の宣言種別定数。ワイヤー値はフロントエンドの enum と一致させる。
const (
	// GutsDeclarationOut アウト (降りる)。ポットに一切関与しない。
	GutsDeclarationOut GutsDeclaration = 0
	// GutsDeclarationIn イン (勝負に残る)。負けた場合はポットをマッチする義務を負う。
	GutsDeclarationIn GutsDeclaration = 1
)

// GutsDeclarationValid は宣言種別が有効な列挙値かどうかを返す。
func GutsDeclarationValid(d GutsDeclaration) bool {
	return d == GutsDeclarationOut || d == GutsDeclarationIn
}

// Guts の設定既定値・境界値。
const (
	// GutsDefaultPlayerCount はデフォルトのプレイヤー数 (人間 1 + CPU 3)。
	GutsDefaultPlayerCount = 4
	// GutsMinPlayerCount は最小プレイヤー数。
	GutsMinPlayerCount = 2
	// GutsMaxPlayerCount は最大プレイヤー数。
	GutsMaxPlayerCount = 7
	// GutsDefaultAnte はデフォルトのアンティ額。
	GutsDefaultAnte = 10
	// GutsMinAnte は最小アンティ額。
	GutsMinAnte = 1
	// GutsMaxAnte は最大アンティ額。
	GutsMaxAnte = 1000
	// GutsDefaultStartingChips はデフォルトの初期チップ。
	GutsDefaultStartingChips = 200
	// GutsMinStartingChips は最小初期チップ。
	GutsMinStartingChips = 10
	// GutsMaxStartingChips は最大初期チップ (Validate 上限)。
	GutsMaxStartingChips = 100000
	// GutsDefaultTargetRounds はデフォルトの実施ラウンド数 (このラウンド数に達するとゲーム終了)。
	GutsDefaultTargetRounds = 10
	// GutsMinTargetRounds は最小ラウンド数。
	GutsMinTargetRounds = 1
	// GutsMaxTargetRounds は最大ラウンド数。
	GutsMaxTargetRounds = 100
)

// GutsConfig はガッツ (Guts) のローカルルール設定。
type GutsConfig struct {
	// PlayerCount は参加プレイヤー数 (人間 seat 0 + CPU、2〜7)。
	PlayerCount int `json:"pc"`
	// Ante は各ラウンドで全員が支払うアンティ額。
	Ante int `json:"an"`
	// StartingChips は各プレイヤーの初期チップ。
	StartingChips int `json:"sc"`
	// TargetRounds はこのラウンド数を消化するとゲームが終了する上限 (エスカレーションの停止条件)。
	TargetRounds int `json:"tr"`
}

// DefaultGutsConfig はデフォルトのローカルルール設定を返す。
func DefaultGutsConfig() GutsConfig {
	return GutsConfig{
		PlayerCount:   GutsDefaultPlayerCount,
		Ante:          GutsDefaultAnte,
		StartingChips: GutsDefaultStartingChips,
		TargetRounds:  GutsDefaultTargetRounds,
	}
}

// Validate は設定値のドメインバリデーションを行う。
func (c GutsConfig) Validate() error {
	if err := ValidateRange("player count", c.PlayerCount, GutsMinPlayerCount, GutsMaxPlayerCount); err != nil {
		return err
	}
	if err := ValidateRange("ante", c.Ante, GutsMinAnte, GutsMaxAnte); err != nil {
		return err
	}
	if err := ValidateRange("starting chips", c.StartingChips, GutsMinStartingChips, GutsMaxStartingChips); err != nil {
		return err
	}
	if c.StartingChips < c.Ante {
		return errors.New("starting chips must be greater than or equal to ante")
	}
	return ValidateRange("target rounds", c.TargetRounds, GutsMinTargetRounds, GutsMaxTargetRounds)
}
