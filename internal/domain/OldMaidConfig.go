package domain

// OldMaidMode ババ抜きモード
type OldMaidMode int

const (
	OldMaidModeNormal   OldMaidMode = iota // ババ抜き: ジョーカーが奇数カード
	OldMaidModeJijiNuki                    // ジジ抜き: ランダム1枚除外
)

// OldMaidConfig ババ抜き設定
type OldMaidConfig struct {
	Mode                 OldMaidMode
	CpuPlacementStrategy bool // CPU心理戦: 奇数カードを端に配置
	CpuMemoryAI          bool // CPU記憶AI: 引いた位置を記憶して戦略的に選択
}

// DefaultOldMaidConfig デフォルト設定を返す
func DefaultOldMaidConfig() OldMaidConfig {
	return OldMaidConfig{Mode: OldMaidModeNormal, CpuPlacementStrategy: false}
}
