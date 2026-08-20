//go:build !js || !wasm || extra4

package domain

import (
	"encoding/json"
	"fmt"
)

// RussianBankCpuDifficulty Russian Bank (Crapette) の CPU 難易度。
type RussianBankCpuDifficulty int

// Russian Bank の難易度定数。難易度は CPU の手番運用と stop 違反の発生率を決める:
//   - Easy: 強制ファウンデーション手を頻繁に取りこぼす (人間が stop で咎めやすい)
//   - Normal: 取りこぼしは時々
//   - Hard: 取りこぼさない (合法な強制手を必ず処理してから手番を終える)
const (
	// RussianBankCpuDifficultyEasy 取りこぼし多め。
	RussianBankCpuDifficultyEasy RussianBankCpuDifficulty = iota
	// RussianBankCpuDifficultyNormal 取りこぼし時々。
	RussianBankCpuDifficultyNormal
	// RussianBankCpuDifficultyHard 取りこぼしなし。
	RussianBankCpuDifficultyHard
)

// RussianBankConfig Russian Bank ゲーム設定。常に 2 人 (人間 1 + CPU 1) 固定のため
// プレイ人数は設定対象外で、CPU 難易度のみを公開する。
type RussianBankConfig struct {
	CpuDifficulty RussianBankCpuDifficulty
}

// DefaultRussianBankConfig 既定設定 (Normal) を返す。
func DefaultRussianBankConfig() RussianBankConfig {
	return RussianBankConfig{CpuDifficulty: RussianBankCpuDifficultyNormal}
}

// Validate 設定値を検証する。
func (c RussianBankConfig) Validate() error {
	if c.CpuDifficulty < RussianBankCpuDifficultyEasy || c.CpuDifficulty > RussianBankCpuDifficultyHard {
		return fmt.Errorf("russianbank: invalid cpu difficulty %d", c.CpuDifficulty)
	}
	return nil
}

// russianBankConfigJSON は RussianBankConfig の JSON ワイヤ形式。
type russianBankConfigJSON struct {
	CpuDifficulty RussianBankCpuDifficulty `json:"cd"`
}

// MarshalJSON implements json.Marshaler.
func (c RussianBankConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(russianBankConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *RussianBankConfig) UnmarshalJSON(data []byte) error {
	var j russianBankConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.CpuDifficulty = j.CpuDifficulty
	return nil
}
