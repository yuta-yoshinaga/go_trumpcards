//go:build !js || !wasm || extra2

package domain

import "encoding/json"

// RistikontraCpuDifficulty は Pişti の CPU 難易度レベル。
type RistikontraCpuDifficulty int

// RistikontraCpuDifficulty 定数
const (
	// RistikontraDifficultyEasy 簡単 (ランダムな合法手)
	RistikontraDifficultyEasy RistikontraCpuDifficulty = 0
	// RistikontraDifficultyNormal 通常 (捕獲できる手を優先)
	RistikontraDifficultyNormal RistikontraCpuDifficulty = 1
	// RistikontraDifficultyHard 難しい (Pişti / 高得点札を狙う)
	RistikontraDifficultyHard RistikontraCpuDifficulty = 2
)

// RistikontraDifficultyNames は難易度名マップ。
var RistikontraDifficultyNames = map[RistikontraCpuDifficulty]string{
	RistikontraDifficultyEasy:   "Easy",
	RistikontraDifficultyNormal: "Normal",
	RistikontraDifficultyHard:   "Hard",
}

// Pişti のプレイヤー数定数。
const (
	// RistikontraMinPlayerCnt 最小プレイヤー数 (4人)。
	//
	// **リスティコントラは常に 2 対 2 の固定パートナーシップ**なので、
	// 席数は 4 で固定。クローン元のピシュティは 2〜4 人の可変人数戦で、
	// そこを引き継ぐと成立しないチーム編成 (2 人卓や 3 人卓) を作れてしまう。
	RistikontraMinPlayerCnt = 4
	// RistikontraMaxPlayerCnt 最大プレイヤー数 (4人)
	RistikontraMaxPlayerCnt = 4
	// RistikontraDefaultPlayerCnt デフォルトのプレイヤー数 (4人)
	RistikontraDefaultPlayerCnt = 4
)

// Pişti の配札・場札関連定数。
const (
	// RistikontraHandSize 一度に各プレイヤーへ配る枚数 (4枚)
	RistikontraHandSize = 4
	// RistikontraInitialPileSize ゲーム開始時に場へ伏せ置く枚数 (4枚)
	RistikontraInitialPileSize = 4
)

// Pişti のスコア定数。
const (
	// RistikontraScoreMostCards 最多捕獲枚数ボーナス (+3)
	RistikontraScoreMostCards = 3
	// RistikontraScoreAce エース1枚あたり (+1)
	RistikontraScoreAce = 1
	// RistikontraScoreTwoClubs 2♣ (+2)
	RistikontraScoreTwoClubs = 2
	// RistikontraScoreTenDiamonds 10♦ (+3)
	RistikontraScoreTenDiamonds = 3
	// RistikontraScoreJack ジャック1枚あたり (+1)
	RistikontraScoreJack = 1
)

// RistikontraJackValue はジャックのカード値。ジャックは常に捕獲を成立させるワイルド札。
const RistikontraJackValue = 11

// RistikontraConfig は Pişti のローカルルール設定。
type RistikontraConfig struct {
	// PlayerCnt プレイヤー数 (2-4, デフォルト 4)。seat 0 が人間。
	PlayerCnt int
	// CpuDifficulty CPU 難易度
	CpuDifficulty RistikontraCpuDifficulty
}

// DefaultRistikontraConfig はデフォルトのローカルルール設定を返す。
//   - プレイヤー数: 4 (1 human + 3 CPU)
//   - CPU 難易度: Normal
func DefaultRistikontraConfig() RistikontraConfig {
	return RistikontraConfig{
		PlayerCnt:     RistikontraDefaultPlayerCnt,
		CpuDifficulty: RistikontraDifficultyNormal,
	}
}

// Validate は設定値のドメインバリデーションを行う。
func (c RistikontraConfig) Validate() error {
	if err := ValidateRange("player count", c.PlayerCnt, RistikontraMinPlayerCnt, RistikontraMaxPlayerCnt); err != nil {
		return err
	}
	return ValidateRange(
		"CPU difficulty",
		int(c.CpuDifficulty),
		int(RistikontraDifficultyEasy),
		int(RistikontraDifficultyHard),
	)
}

// ristikontraConfigJSON is the JSON wire format for RistikontraConfig.
type ristikontraConfigJSON struct {
	PlayerCnt     int                      `json:"pc"`
	CpuDifficulty RistikontraCpuDifficulty `json:"di"`
}

// MarshalJSON implements json.Marshaler.
func (c RistikontraConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(ristikontraConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *RistikontraConfig) UnmarshalJSON(data []byte) error {
	var j ristikontraConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	*c = RistikontraConfig(j)
	return nil
}
