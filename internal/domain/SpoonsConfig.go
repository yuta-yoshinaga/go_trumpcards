//go:build !js || !wasm || extra2

package domain

import "encoding/json"

// SpoonsPlayerCnt はスプーンのプレイヤー数 (人間 1 + CPU 3)。
const SpoonsPlayerCnt = 4

// SpoonsHandSize は各プレイヤーが通常時に持つ手札枚数。
// 4 枚すべてが同じランクになると「フォーオブアカインド」成立となる。
const SpoonsHandSize = 4

// SpoonsMaxLetters はプレイヤーが脱落するまでに溜められる文字数 (S-P-O-O-N-S = 6)。
const SpoonsMaxLetters = 6

// SpoonsMaxRounds は無限ループ防止のためのラウンド数上限 (フルCPU対戦の停止保証)。
const SpoonsMaxRounds = 1000

// SpoonsMaxPassesPerRound は 1 ラウンドあたりのパス回数上限。これを超えると
// フェイルセーフとして現在の手番プレイヤーがフォーオブアカインドを揃えたものと
// みなしてグラブウィンドウを強制的に開く (各ラウンドの停止保証)。
const SpoonsMaxPassesPerRound = 500

// SpoonsCpuDifficulty は CPU の難易度。
type SpoonsCpuDifficulty int

// スプーン CPU 難易度定数
const (
	// SpoonsCpuEasy 反応が鈍くグラブに乗り遅れやすい初心者向け
	SpoonsCpuEasy SpoonsCpuDifficulty = 0
	// SpoonsCpuNormal 標準的な反応
	SpoonsCpuNormal SpoonsCpuDifficulty = 1
	// SpoonsCpuHard 反応が速くグラブに乗り遅れにくい
	SpoonsCpuHard SpoonsCpuDifficulty = 2
)

// SpoonsConfig はスプーンのゲーム設定。
type SpoonsConfig struct {
	// CpuDifficulty CPU の難易度
	CpuDifficulty SpoonsCpuDifficulty
}

// DefaultSpoonsConfig はデフォルト設定を返す。
func DefaultSpoonsConfig() SpoonsConfig {
	return SpoonsConfig{CpuDifficulty: SpoonsCpuNormal}
}

// Validate は設定値の妥当性を検証する。
func (c SpoonsConfig) Validate() error {
	return ValidateRange("cpu difficulty", int(c.CpuDifficulty),
		int(SpoonsCpuEasy), int(SpoonsCpuHard))
}

// GrabMissChance はグラブウィンドウが開いたとき、CPU がスプーンの取得に
// 失敗 (=遅れる) 確率を難易度別に返す。Hard ほど低い。
func (c SpoonsConfig) GrabMissChance() float64 {
	switch c.CpuDifficulty {
	case SpoonsCpuEasy:
		return 0.45
	case SpoonsCpuHard:
		return 0.10
	default:
		return 0.25
	}
}

// spoonsConfigJSON is the JSON wire format for SpoonsConfig.
type spoonsConfigJSON struct {
	CpuDifficulty int `json:"cd"`
}

// MarshalJSON implements json.Marshaler.
func (c SpoonsConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(spoonsConfigJSON{CpuDifficulty: int(c.CpuDifficulty)})
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *SpoonsConfig) UnmarshalJSON(data []byte) error {
	var j spoonsConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.CpuDifficulty = SpoonsCpuDifficulty(j.CpuDifficulty)
	return nil
}
