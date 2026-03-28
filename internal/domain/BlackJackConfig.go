package domain

import "fmt"

// カウンティングシステム定数
const (
	BJCountingHiLo    = 0 // Hi-Lo
	BJCountingKO      = 1 // Knock-Out (KO)
	BJCountingZen     = 2 // Zen Count
	BJCountingOmegaII = 3 // Omega II
	BJCountingMax     = 3 // 最大値
)

// サレンダールール定数
const (
	BJSurrenderLate  = 0 // レイトサレンダー（デフォルト）
	BJSurrenderEarly = 1 // アーリーサレンダー
	BJSurrenderNone  = 2 // サレンダー無効
	BJSurrenderMax   = 2 // 最大値
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
	DealerHitsSoft17 bool `json:"dh"` // ディーラーがソフト17でヒットするか (H17 vs S17)
	CpuPlayerCount   int  `json:"cp"` // CPUプレイヤー数 (0-3)
	CountingEnabled  bool `json:"ce"` // カウンティング表示有効
	DoubleAfterSplit bool `json:"ds"` // スプリット後のダブルダウン許可 (DAS)
	CountingSystem   int  `json:"cs"` // カウンティングシステム (0=Hi-Lo, 1=KO, 2=Zen, 3=Omega II)
	DeckPenetration  int  `json:"dp"` // デッキペネトレーション率(%) (50 or 75, 0=デフォルト75)
	SurrenderRule    int  `json:"sr"` // サレンダールール (0=Late, 1=Early, 2=None)
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

// Validate 設定値のドメインバリデーション
func (c BlackJackConfig) Validate() error {
	if err := ValidateRange("CPU player count", c.CpuPlayerCount, 0, BJMaxCpuPlayers); err != nil {
		return err
	}
	if err := ValidateRange("counting system", c.CountingSystem, 0, BJCountingMax); err != nil {
		return err
	}
	if err := ValidateRange("surrender rule", c.SurrenderRule, 0, BJSurrenderMax); err != nil {
		return err
	}
	if c.DeckPenetration != 0 {
		validPen := false
		for _, v := range BJValidPenetrations {
			if v == c.DeckPenetration {
				validPen = true
				break
			}
		}
		if !validPen {
			return fmt.Errorf("deck penetration must be 50 or 75, got %d", c.DeckPenetration)
		}
	}
	return nil
}

// IsBalancedCountingSystem バランスドカウンティングシステムか (TC計算可能)
func IsBalancedCountingSystem(system int) bool {
	return system != BJCountingKO
}
