//go:build !js || !wasm || extra2

package domain

import "encoding/json"

// KempsPlayerCnt はケムプスのプレイヤー数 (人間 1 + CPU 3)。
const KempsPlayerCnt = 4

// KempsHandSize は各プレイヤーが持つ手札枚数。4 枚すべてが同じランクになると
// 「フォーオブアカインド」成立 (= Kemps 宣言の権利が発生) となる。
const KempsHandSize = 4

// KempsFieldSize は場 (フィールド) に表向きで並ぶ共有カードの枚数。
const KempsFieldSize = 4

// KempsTargetScore は勝利に必要なチーム得点 (デフォルト 5)。
const KempsTargetScore = 5

// KempsMaxRounds は無限ループ防止のためのラウンド数上限 (停止保証のフェイルセーフ)。
const KempsMaxRounds = 2000

// KempsMaxSwapsPerRound は 1 ラウンドあたりのスワップ/パス回数上限。これを超えると
// フェイルセーフとして強制的にラウンドを締める (各ラウンドの停止保証)。
const KempsMaxSwapsPerRound = 2000

// KempsTeamCnt はチーム数 (2)。
const KempsTeamCnt = 2

// KempsTeamOf はプレイヤーインデックスからチーム番号 (0 または 1) を返す。
// 偶数席 (0,2) = チーム A、奇数席 (1,3) = チーム B。
func KempsTeamOf(i int) int { return i % 2 }

// KempsPartnerOf はプレイヤー i のパートナー席 (対角) を返す。
func KempsPartnerOf(i int) int { return (i + 2) % KempsPlayerCnt }

// KempsCpuDifficulty は CPU の難易度。
type KempsCpuDifficulty int

// ケムプス CPU 難易度定数
const (
	// KempsCpuEasy 反応が鈍く、相手のシグナルを見抜きにくい初心者向け
	KempsCpuEasy KempsCpuDifficulty = 0
	// KempsCpuNormal 標準的な反応
	KempsCpuNormal KempsCpuDifficulty = 1
	// KempsCpuHard 反応が速く、相手のシグナルを見抜きやすい
	KempsCpuHard KempsCpuDifficulty = 2
)

// SignalType は人間プレイヤーが事前に決める秘密のシグナル種別。
type SignalType int

// シグナル種別定数
const (
	// SignalSound 音 (咳払いなど) によるシグナル
	SignalSound SignalType = 0
	// SignalBlink 瞬き (合図) によるシグナル
	SignalBlink SignalType = 1
)

// SignalTypeMin / SignalTypeMax はシグナル種別の範囲 (復元時の検証用)。
const (
	SignalTypeMin = SignalSound
	SignalTypeMax = SignalBlink
)

// KempsConfig はケムプスのゲーム設定。
type KempsConfig struct {
	// CpuDifficulty CPU の難易度
	CpuDifficulty KempsCpuDifficulty
	// TargetScore 勝利に必要なチーム得点
	TargetScore int
}

// DefaultKempsConfig はデフォルト設定を返す。
func DefaultKempsConfig() KempsConfig {
	return KempsConfig{CpuDifficulty: KempsCpuNormal, TargetScore: KempsTargetScore}
}

// Validate は設定値の妥当性を検証する。
func (c KempsConfig) Validate() error {
	if err := ValidateRange("cpu difficulty", int(c.CpuDifficulty),
		int(KempsCpuEasy), int(KempsCpuHard)); err != nil {
		return err
	}
	return ValidateRange("target score", c.TargetScore, 1, 99)
}

// CounterChance は相手チームのフォーオブアカインドに対して、CPU が
// カウンターケムプスを正しく見抜く確率を難易度別に返す。Hard ほど高い。
func (c KempsConfig) CounterChance() float64 {
	switch c.CpuDifficulty {
	case KempsCpuEasy:
		return 0.10
	case KempsCpuHard:
		return 0.45
	default:
		return 0.25
	}
}

// kempsConfigJSON is the JSON wire format for KempsConfig.
type kempsConfigJSON struct {
	CpuDifficulty int `json:"cd"`
	TargetScore   int `json:"ts"`
}

// MarshalJSON implements json.Marshaler.
func (c KempsConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(kempsConfigJSON{
		CpuDifficulty: int(c.CpuDifficulty),
		TargetScore:   c.TargetScore,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *KempsConfig) UnmarshalJSON(data []byte) error {
	var j kempsConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.CpuDifficulty = KempsCpuDifficulty(j.CpuDifficulty)
	c.TargetScore = j.TargetScore
	if c.TargetScore <= 0 {
		c.TargetScore = KempsTargetScore
	}
	return nil
}
