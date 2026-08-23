//go:build !js || !wasm || extra

package domain

// BauernschnapsenCpuDifficulty CPU の難易度レベル
type BauernschnapsenCpuDifficulty int

// BauernschnapsenのCPU難易度定数
const (
	// BauernschnapsenCpuDifficultyEasy 低難易度
	BauernschnapsenCpuDifficultyEasy BauernschnapsenCpuDifficulty = iota
	// BauernschnapsenCpuDifficultyNormal 中難易度
	BauernschnapsenCpuDifficultyNormal
	// BauernschnapsenCpuDifficultyHard 高難易度
	BauernschnapsenCpuDifficultyHard
)

// BauernschnapsenConfig バウエルンシュナプセンゲーム設定
type BauernschnapsenConfig struct {
	CpuDifficulty BauernschnapsenCpuDifficulty `json:"cd"`
	// TargetScore ゲーム終了スコア (先に到達したチームが勝利)。
	//
	// **既定は 24。** クローン元のガイゲルはカード点をそのまま積むので 101 が
	// 妥当だったが、こちらは契約の成否で 1〜3 点しか動かない。101 のままだと
	// 1 ゲームが 50 ラウンド近くかかる。
	TargetScore int `json:"ts"`
}

// DefaultBauernschnapsenConfig デフォルト設定を返す
func DefaultBauernschnapsenConfig() BauernschnapsenConfig {
	return BauernschnapsenConfig{
		CpuDifficulty: BauernschnapsenCpuDifficultyNormal,
		TargetScore:   24,
	}
}

// Validate 設定値のドメインバリデーション
func (c BauernschnapsenConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(BauernschnapsenCpuDifficultyEasy), int(BauernschnapsenCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target score", c.TargetScore, 1); err != nil {
		return err
	}
	return nil
}
