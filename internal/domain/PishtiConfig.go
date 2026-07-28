//go:build !js || !wasm || extra2

package domain

import "encoding/json"

// PishtiCpuDifficulty は Pişti の CPU 難易度レベル。
type PishtiCpuDifficulty int

// PishtiCpuDifficulty 定数
const (
	// PishtiDifficultyEasy 簡単 (ランダムな合法手)
	PishtiDifficultyEasy PishtiCpuDifficulty = 0
	// PishtiDifficultyNormal 通常 (捕獲できる手を優先)
	PishtiDifficultyNormal PishtiCpuDifficulty = 1
	// PishtiDifficultyHard 難しい (Pişti / 高得点札を狙う)
	PishtiDifficultyHard PishtiCpuDifficulty = 2
)

// PishtiDifficultyNames は難易度名マップ。
var PishtiDifficultyNames = map[PishtiCpuDifficulty]string{
	PishtiDifficultyEasy:   "Easy",
	PishtiDifficultyNormal: "Normal",
	PishtiDifficultyHard:   "Hard",
}

// Pişti のプレイヤー数定数。
const (
	// PishtiMinPlayerCnt 最小プレイヤー数 (2人)
	PishtiMinPlayerCnt = 2
	// PishtiMaxPlayerCnt 最大プレイヤー数 (4人)
	PishtiMaxPlayerCnt = 4
	// PishtiDefaultPlayerCnt デフォルトのプレイヤー数 (4人)
	PishtiDefaultPlayerCnt = 4
)

// Pişti の配札・場札関連定数。
const (
	// PishtiHandSize 一度に各プレイヤーへ配る枚数 (4枚)
	PishtiHandSize = 4
	// PishtiInitialPileSize ゲーム開始時に場へ伏せ置く枚数 (4枚)
	PishtiInitialPileSize = 4
)

// Pişti のスコア定数。
const (
	// PishtiScoreMostCards 最多捕獲枚数ボーナス (+3)
	PishtiScoreMostCards = 3
	// PishtiScoreAce エース1枚あたり (+1)
	PishtiScoreAce = 1
	// PishtiScoreTwoClubs 2♣ (+2)
	PishtiScoreTwoClubs = 2
	// PishtiScoreTenDiamonds 10♦ (+3)
	PishtiScoreTenDiamonds = 3
	// PishtiScoreJack ジャック1枚あたり (+1)
	PishtiScoreJack = 1
	// PishtiBonusSingle 単独札を捕獲した Pişti ボーナス (+10)
	PishtiBonusSingle = 10
	// PishtiBonusJackOnJack ジャックを単独ジャックに重ねた Pişti ボーナス (+20)
	PishtiBonusJackOnJack = 20
)

// PishtiJackValue はジャックのカード値。ジャックは常に捕獲を成立させるワイルド札。
const PishtiJackValue = 11

// PishtiConfig は Pişti のローカルルール設定。
type PishtiConfig struct {
	// PlayerCnt プレイヤー数 (2-4, デフォルト 4)。seat 0 が人間。
	PlayerCnt int
	// CpuDifficulty CPU 難易度
	CpuDifficulty PishtiCpuDifficulty
}

// DefaultPishtiConfig はデフォルトのローカルルール設定を返す。
//   - プレイヤー数: 4 (1 human + 3 CPU)
//   - CPU 難易度: Normal
func DefaultPishtiConfig() PishtiConfig {
	return PishtiConfig{
		PlayerCnt:     PishtiDefaultPlayerCnt,
		CpuDifficulty: PishtiDifficultyNormal,
	}
}

// Validate は設定値のドメインバリデーションを行う。
func (c PishtiConfig) Validate() error {
	if err := ValidateRange("player count", c.PlayerCnt, PishtiMinPlayerCnt, PishtiMaxPlayerCnt); err != nil {
		return err
	}
	return ValidateRange(
		"CPU difficulty",
		int(c.CpuDifficulty),
		int(PishtiDifficultyEasy),
		int(PishtiDifficultyHard),
	)
}

// pishtiConfigJSON is the JSON wire format for PishtiConfig.
type pishtiConfigJSON struct {
	PlayerCnt     int                 `json:"pc"`
	CpuDifficulty PishtiCpuDifficulty `json:"di"`
}

// MarshalJSON implements json.Marshaler.
func (c PishtiConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(pishtiConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *PishtiConfig) UnmarshalJSON(data []byte) error {
	var j pishtiConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	*c = PishtiConfig(j)
	return nil
}
