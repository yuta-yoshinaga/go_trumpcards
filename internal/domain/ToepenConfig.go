//go:build !js || !wasm || extra3

package domain

// ToepenCpuDifficulty は CPU の難易度レベル。
type ToepenCpuDifficulty int

// ToepenのCPU難易度定数
const (
	// ToepenCpuDifficultyNormal 標準難易度 (v1で唯一サポート)
	ToepenCpuDifficultyNormal ToepenCpuDifficulty = iota
)

// ToepenConfig はトゥーペンのゲーム設定。
type ToepenConfig struct {
	CpuDifficulty ToepenCpuDifficulty `json:"cd"`
	// PlayerCnt は参加人数。pagat は 3〜8 人としている。
	PlayerCnt int `json:"pc"`
}

// DefaultToepenConfig はデフォルト設定を返す。
func DefaultToepenConfig() ToepenConfig {
	return ToepenConfig{
		CpuDifficulty: ToepenCpuDifficultyNormal,
		PlayerCnt:     ToepenPlayerCnt,
	}
}

// Validate は設定値のドメインバリデーション。
func (c ToepenConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(ToepenCpuDifficultyNormal), int(ToepenCpuDifficultyNormal)); err != nil {
		return err
	}
	return ValidateRange("player count", c.PlayerCnt, ToepenMinPlayers, ToepenMaxPlayers)
}
