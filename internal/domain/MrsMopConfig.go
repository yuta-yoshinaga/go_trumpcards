//go:build !js || !wasm || extra4

package domain

// MrsMopDifficulty ミセス・モップソリティア難易度
type MrsMopDifficulty int

// MrsMopの難易度定数
//
// **4スートが Mrs. Mop 本来の形** ── 2デッキ104枚をそのまま使う。1/2スートは
// 遊びやすさのための緩和版で、クローン元の Spider から引き継いだもの。
// **既定を Spider のまま 1スートにしてはいけない**: 最初に開いた盤が
// Mrs. Mop でなくなる。
const (
	// MrsMopDifficulty1Suit 1スート（初級・緩和版）
	MrsMopDifficulty1Suit MrsMopDifficulty = 1
	// MrsMopDifficulty2Suit 2スート（中級・緩和版）
	MrsMopDifficulty2Suit MrsMopDifficulty = 2
	// MrsMopDifficulty4Suit 4スート（Mrs. Mop 本来の形）
	MrsMopDifficulty4Suit MrsMopDifficulty = 4
)

// MrsMopConfig ミセス・モップソリティア設定
type MrsMopConfig struct {
	Difficulty MrsMopDifficulty
}

// DefaultMrsMopConfig デフォルト設定（4スート = 本来の Mrs. Mop）
func DefaultMrsMopConfig() MrsMopConfig {
	return MrsMopConfig{Difficulty: MrsMopDifficulty4Suit}
}
