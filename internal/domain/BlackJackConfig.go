package domain

// カウンティングシステム定数
const (
	BJCountingHiLo    = 0 // Hi-Lo
	BJCountingKO      = 1 // Knock-Out (KO)
	BJCountingZen     = 2 // Zen Count
	BJCountingOmegaII = 3 // Omega II
	BJCountingMax     = 3 // 最大値
)

// デッキペネトレーション定数
const (
	BJDefaultPenetration = 75 // デフォルトペネトレーション率(%)
	BJPenetrationMin     = 50 // 最小ペネトレーション率(%)
	BJPenetrationMax     = 75 // 最大ペネトレーション率(%)
)

// BJValidPenetrations 有効なペネトレーション値
var BJValidPenetrations = []int{50, 75}

// BlackJackConfig ブラックジャックゲーム設定
type BlackJackConfig struct {
	DealerHitsSoft17 bool // ディーラーがソフト17でヒットするか (H17 vs S17)
	CpuPlayerCount   int  // CPUプレイヤー数 (0-3)
	CountingEnabled  bool // カウンティング表示有効
	DoubleAfterSplit bool // スプリット後のダブルダウン許可 (DAS)
	CountingSystem   int  // カウンティングシステム (0=Hi-Lo, 1=KO, 2=Zen, 3=Omega II)
	DeckPenetration  int  // デッキペネトレーション率(%) (50 or 75, 0=デフォルト75)
}

// DefaultBlackJackConfig デフォルト設定 (全機能無効)
func DefaultBlackJackConfig() BlackJackConfig {
	return BlackJackConfig{
		DealerHitsSoft17: false,
		CpuPlayerCount:   0,
		CountingEnabled:  false,
		DoubleAfterSplit: true,
	}
}

// IsBalancedCountingSystem バランスドカウンティングシステムか (TC計算可能)
func IsBalancedCountingSystem(system int) bool {
	return system != BJCountingKO
}
