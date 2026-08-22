//go:build !js || !wasm || extra

package domain

// スペキュレーションの卓設定の既定値と範囲。
const (
	// SpeculationDefaultPlayers は既定の参加人数 (人間 1 + CPU)。
	SpeculationDefaultPlayers = 4
	// SpeculationMinPlayers / SpeculationMaxPlayers は参加人数の範囲。
	// 原作は 3〜8 人だが、ここは 1 人プレイへの翻案なので人間 1 + CPU 1〜7。
	SpeculationMinPlayers = 2
	SpeculationMaxPlayers = 8

	// SpeculationDefaultChips は各プレイヤーの初期チップ。
	SpeculationDefaultChips = 200
	// SpeculationDefaultStake は 1 ラウンドの参加料 (全員がポットに出す)。
	SpeculationDefaultStake = 10
	// SpeculationDefaultRounds は既定のラウンド数。
	SpeculationDefaultRounds = 5
	// SpeculationMinRounds / SpeculationMaxRounds はラウンド数の範囲。
	SpeculationMinRounds = 1
	SpeculationMaxRounds = 20

	// SpeculationCardsPerPlayer は各プレイヤーに配る伏せ札の枚数。
	// **3 枚固定。** 8 人卓でも 24 枚 + 切り札 1 枚で 52 枚に収まる。
	SpeculationCardsPerPlayer = 3
)

// SpeculationConfig はスペキュレーションの卓設定。
type SpeculationConfig struct {
	// Players は人間 1 人を含む参加人数。
	Players int
	// InitialChips は各プレイヤーの初期チップ。
	InitialChips int
	// Stake は 1 ラウンドの参加料。全員が出し、勝者がまとめて取る。
	Stake int
	// Rounds は規定ラウンド数。
	Rounds int
}

// NewDefaultSpeculationConfig は既定の卓設定を返す。
func NewDefaultSpeculationConfig() SpeculationConfig {
	return SpeculationConfig{
		Players:      SpeculationDefaultPlayers,
		InitialChips: SpeculationDefaultChips,
		Stake:        SpeculationDefaultStake,
		Rounds:       SpeculationDefaultRounds,
	}
}

// Normalize は範囲外の値を既定値に丸める。**復元した設定も通す** ——
// KV から戻した卓が 0 人や 0 ラウンドだと配りも進行も成立しない。
func (c *SpeculationConfig) Normalize() {
	if c.Players < SpeculationMinPlayers || c.Players > SpeculationMaxPlayers {
		c.Players = SpeculationDefaultPlayers
	}
	if c.InitialChips <= 0 {
		c.InitialChips = SpeculationDefaultChips
	}
	if c.Stake <= 0 {
		c.Stake = SpeculationDefaultStake
	}
	if c.Rounds < SpeculationMinRounds || c.Rounds > SpeculationMaxRounds {
		c.Rounds = SpeculationDefaultRounds
	}
}
