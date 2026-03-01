package domain

// BlackJackConfig ブラックジャックゲーム設定
type BlackJackConfig struct {
	DealerHitsSoft17 bool // ディーラーがソフト17でヒットするか (H17 vs S17)
	CpuPlayerCount   int  // CPUプレイヤー数 (0-3)
	CountingEnabled  bool // カウンティング表示有効
}

// DefaultBlackJackConfig デフォルト設定 (全機能無効)
func DefaultBlackJackConfig() BlackJackConfig {
	return BlackJackConfig{
		DealerHitsSoft17: false,
		CpuPlayerCount:   0,
		CountingEnabled:  false,
	}
}
