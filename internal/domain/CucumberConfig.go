//go:build !js || !wasm || classic

package domain

import (
	"encoding/json"
	"fmt"
)

const (
	// CucumberPlayerCntMin は最小プレイヤー数。
	CucumberPlayerCntMin = 3
	// CucumberPlayerCntMax は最大プレイヤー数。
	CucumberPlayerCntMax = 6
	// CucumberDefaultPlayerCnt は既定のプレイヤー数。
	CucumberDefaultPlayerCnt = 4
)

// CucumberHandSize は 1 ラウンドで各自が受け取る枚数。
//
// **人数で割り切ろうとしないこと。** 52 枚は 3 / 5 / 6 人で割り切れないので、
// 伝統どおり **7 枚固定**にして残りを使いません（3〜6 人で 21 / 28 / 35 / 42 枚）。
const CucumberHandSize = 7

const (
	// CucumberTargetScoreMin は失点上限の最小値。
	CucumberTargetScoreMin = 10
	// CucumberTargetScoreMax は失点上限の最大値。
	CucumberTargetScoreMax = 100
	// CucumberDefaultTargetScore は既定の失点上限。
	CucumberDefaultTargetScore = 30
)

// CucumberConfig はキューカンバーのゲーム設定。
type CucumberConfig struct {
	// PlayerCnt は参加人数。
	PlayerCnt int
	// TargetScore は誰かが到達したら終わる失点。
	TargetScore int
}

// DefaultCucumberConfig はデフォルト設定を返す。
func DefaultCucumberConfig() CucumberConfig {
	return CucumberConfig{PlayerCnt: CucumberDefaultPlayerCnt, TargetScore: CucumberDefaultTargetScore}
}

// Validate は設定値の妥当性を検証する。
func (c CucumberConfig) Validate() error {
	if c.PlayerCnt < CucumberPlayerCntMin || c.PlayerCnt > CucumberPlayerCntMax {
		return fmt.Errorf("player count must be between %d and %d, got %d",
			CucumberPlayerCntMin, CucumberPlayerCntMax, c.PlayerCnt)
	}
	return ValidateRange("target score", c.TargetScore, CucumberTargetScoreMin, CucumberTargetScoreMax)
}

// cucumberConfigJSON is the JSON wire format for CucumberConfig.
type cucumberConfigJSON struct {
	PlayerCnt   int `json:"p"`
	TargetScore int `json:"ts"`
}

// MarshalJSON implements json.Marshaler.
func (c CucumberConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(cucumberConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *CucumberConfig) UnmarshalJSON(data []byte) error {
	var j cucumberConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	*c = CucumberConfig(j)
	return c.Validate()
}
