//go:build !js || !wasm || extra3

package domain

// KaiserCpuDifficulty CPU の難易度レベル
type KaiserCpuDifficulty int

// Kaiser の CPU 難易度定数
const (
	// KaiserCpuDifficultyNormal 中難易度 (v1 はこれのみ)
	KaiserCpuDifficultyNormal KaiserCpuDifficulty = iota
)

// KaiserConfig カイザーのゲーム設定
type KaiserConfig struct {
	CpuDifficulty KaiserCpuDifficulty `json:"cd"`
	// AllowNoTrump はノートランプ系のビッドを許すか。
	AllowNoTrump bool `json:"nt"`
}

// DefaultKaiserConfig デフォルト設定を返す
func DefaultKaiserConfig() KaiserConfig {
	return KaiserConfig{
		CpuDifficulty: KaiserCpuDifficultyNormal,
		AllowNoTrump:  true,
	}
}

// Validate 設定値のドメインバリデーション
func (c KaiserConfig) Validate() error {
	return ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(KaiserCpuDifficultyNormal), int(KaiserCpuDifficultyNormal))
}
