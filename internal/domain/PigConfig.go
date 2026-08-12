//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"fmt"
)

const (
	// PigPlayerCntMin は最小プレイヤー数。
	PigPlayerCntMin = 3
	// PigPlayerCntMax は最大プレイヤー数。
	PigPlayerCntMax = 6
	// PigDefaultPlayerCnt は既定のプレイヤー数。
	PigDefaultPlayerCnt = 4
)

// PigHandSize は各プレイヤーが常に持つ手札枚数。
//
// **枚数は最初から最後まで変わりません。** 全員が同時に 1 枚渡して 1 枚受け取る
// ので、渡す前後で 4 枚のままです。
const PigHandSize = 4

// PigMaxLetters は脱落するまでに溜められる文字数 (P-I-G = 3)。
const PigMaxLetters = 3

// PigMaxRounds はゲーム全体のラウンド数上限 (停止保証のフェイルセーフ)。
const PigMaxRounds = 200

// PigMaxPassesPerRound は 1 ラウンドあたりのパス回数上限。
//
// **揃わないまま回り続ける配りが存在します。** 全員が同じ札を押し付け合うと
// 循環しうるので、上限を超えたら現在の手番が揃えたものとみなして合図を開きます。
const PigMaxPassesPerRound = 200

// PigCpuDifficulty は CPU の難易度。
type PigCpuDifficulty int

// ピッグ CPU 難易度定数
const (
	// PigCpuEasy 合図に気づくのが遅い初心者向け
	PigCpuEasy PigCpuDifficulty = 0
	// PigCpuNormal 標準的な反応
	PigCpuNormal PigCpuDifficulty = 1
	// PigCpuHard 合図にすぐ気づく
	PigCpuHard PigCpuDifficulty = 2
)

// PigConfig はピッグのゲーム設定。
type PigConfig struct {
	// PlayerCnt は参加人数。デッキの大きさもこれで決まる。
	PlayerCnt int
	// CpuDifficulty は CPU の難易度。
	CpuDifficulty PigCpuDifficulty
}

// DefaultPigConfig はデフォルト設定を返す。
func DefaultPigConfig() PigConfig {
	return PigConfig{PlayerCnt: PigDefaultPlayerCnt, CpuDifficulty: PigCpuNormal}
}

// Validate は設定値の妥当性を検証する。
func (c PigConfig) Validate() error {
	if c.PlayerCnt < PigPlayerCntMin || c.PlayerCnt > PigPlayerCntMax {
		return fmt.Errorf("player count must be between %d and %d, got %d",
			PigPlayerCntMin, PigPlayerCntMax, c.PlayerCnt)
	}
	return ValidateRange("cpu difficulty", int(c.CpuDifficulty), int(PigCpuEasy), int(PigCpuHard))
}

// NoticeMissChance は合図が出た 1 手ごとに CPU が気づき損ねる確率を返す。
func (c PigConfig) NoticeMissChance() float64 {
	switch c.CpuDifficulty {
	case PigCpuEasy:
		return 0.55
	case PigCpuHard:
		return 0.15
	default:
		return 0.35
	}
}

// PigDeckSize は人数に対するデッキ枚数を返す。
//
// **人数と同じ種類のランクを 4 スート揃えて使います。** 4 人なら 4 ランク ×
// 4 スート = 16 枚で、これがちょうど 4 人 × 4 枚。**新しい Card 型は要りません**
// ——標準 52 枚の部分集合です。
func PigDeckSize(playerCnt int) int { return playerCnt * PigHandSize }

// pigConfigJSON is the JSON wire format for PigConfig.
type pigConfigJSON struct {
	PlayerCnt     int `json:"p"`
	CpuDifficulty int `json:"cd"`
}

// MarshalJSON implements json.Marshaler.
func (c PigConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(pigConfigJSON{PlayerCnt: c.PlayerCnt, CpuDifficulty: int(c.CpuDifficulty)})
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *PigConfig) UnmarshalJSON(data []byte) error {
	var j pigConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.PlayerCnt = j.PlayerCnt
	c.CpuDifficulty = PigCpuDifficulty(j.CpuDifficulty)
	return c.Validate()
}
