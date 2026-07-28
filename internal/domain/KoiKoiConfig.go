//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// KoiKoiCpuDifficulty は CPU の難易度レベル。
type KoiKoiCpuDifficulty int

// Koi-Koi の CPU 難易度定数
const (
	// KoiKoiCpuDifficultyEasy 低難易度 (合法手からランダム、決断は常に止める)
	KoiKoiCpuDifficultyEasy KoiKoiCpuDifficulty = iota
	// KoiKoiCpuDifficultyNormal 中難易度 (捕獲価値を優先)
	KoiKoiCpuDifficultyNormal
	// KoiKoiCpuDifficultyHard 高難易度 (役を狙い、こいこいで積極的に続行)
	KoiKoiCpuDifficultyHard
)

// KoiKoiDifficultyNames 難易度名マップ
var KoiKoiDifficultyNames = map[KoiKoiCpuDifficulty]string{
	KoiKoiCpuDifficultyEasy:   "Easy",
	KoiKoiCpuDifficultyNormal: "Normal",
	KoiKoiCpuDifficultyHard:   "Hard",
}

// KoiKoiTargetScoreMin / Max は目標得点の許容範囲。
const (
	KoiKoiTargetScoreMin = 1
	KoiKoiTargetScoreMax = 200
)

// KoiKoiConfig はこいこい (Koi-Koi) のローカルルール設定。
type KoiKoiConfig struct {
	// CpuDifficulty CPU 難易度
	CpuDifficulty KoiKoiCpuDifficulty `json:"cd"`
	// TargetScore この累計得点に到達したプレイヤーが出た時点でゲーム終了。
	// 到達者がいなくても KoiKoiMaxRounds ラウンドで打ち切る。
	TargetScore int `json:"ts"`
}

// DefaultKoiKoiConfig はデフォルトのローカルルール設定を返す。
//   - プレイヤー数: 2 (1 human + 1 CPU)
//   - デッキ: 花札 48 枚 (12 か月 × 4)
//   - 手札 8 枚 / 場 8 枚
//   - 目標得点: 15 (これに到達で終局、最長 KoiKoiMaxRounds ラウンド)
//   - CPU 難易度: 3 段階
func DefaultKoiKoiConfig() KoiKoiConfig {
	return KoiKoiConfig{
		CpuDifficulty: KoiKoiCpuDifficultyNormal,
		TargetScore:   15,
	}
}

// Validate は設定値のドメインバリデーションを行う。
func (c KoiKoiConfig) Validate() error {
	if err := ValidateRange(
		"CPU difficulty",
		int(c.CpuDifficulty),
		int(KoiKoiCpuDifficultyEasy),
		int(KoiKoiCpuDifficultyHard),
	); err != nil {
		return err
	}
	return ValidateRange("target score", c.TargetScore, KoiKoiTargetScoreMin, KoiKoiTargetScoreMax)
}

// koikoiConfigJSON is the JSON wire format for KoiKoiConfig.
type koikoiConfigJSON struct {
	CpuDifficulty KoiKoiCpuDifficulty `json:"cd"`
	TargetScore   int                 `json:"ts"`
}

// MarshalJSON implements json.Marshaler.
func (c KoiKoiConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(koikoiConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *KoiKoiConfig) UnmarshalJSON(data []byte) error {
	var j koikoiConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	*c = KoiKoiConfig(j)
	return nil
}
