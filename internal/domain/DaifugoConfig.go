package domain

import "fmt"

// DaifugoFiveSkipCountMax 5飛びスキップ数最大
const DaifugoFiveSkipCountMax = 5

// DaifugoDifficultyNames 難易度名マップ
var DaifugoDifficultyNames = map[DaifugoCpuDifficulty]string{
	DaifugoDifficultyNormal: "Normal",
	DaifugoDifficultyEasy:   "Easy",
	DaifugoDifficultyHard:   "Hard",
}

// DaifugoConfig 大富豪ローカルルール設定
type DaifugoConfig struct {
	JokerCount                int                  // ジョーカー枚数 (default: 2)
	EightCutEnabled           bool                 // 8切り
	SuitLockMode              DaifugoSuitLockMode  // スート縛りモード (0=なし, 1=片縛り, 2=両縛り)
	ElevenBackEnabled         bool                 // 11バック
	SequenceEnabled           bool                 // 階段
	CardExchangeEnabled       bool                 // カード交換
	BlindExchangeEnabled      bool                 // ブラインドカード交換 (上位が弱い札ではなくランダム)
	FiveSkipEnabled           bool                 // 5飛び
	FiveSkipCount             int                  // 5飛びスキップ数 (default: 1, FiveSkipEnabled時のみ有効)
	SevenPassEnabled          bool                 // 7渡し
	TenDiscardEnabled         bool                 // 10捨て
	SpadeThreeEnabled         bool                 // スペ3返し
	CapitalFallEnabled        bool                 // 都落ち
	NineReverseEnabled        bool                 // 9リバース
	CoupDetatEnabled          bool                 // クーデター (3枚の9で革命)
	NumberLockEnabled         bool                 // 数縛り (連番縛り)
	SandstormEnabled          bool                 // 砂嵐 (3枚の3で場をクリア)
	EmperorEnabled            bool                 // エンペラー (4枚連番・全スート異なる→革命+場クリア)
	SequenceRevolutionEnabled bool                 // 階段革命 (4枚以上の階段で革命)
	SequenceLockEnabled       bool                 // 階段縛り (階段に階段を出すと以降階段のみ)
	IllegalFinishEnabled      bool                 // 反則上がり (8切り/ジョーカー/革命で上がりはペナルティ)
	QueenBomberEnabled        bool                 // 12ボンバー (Qを出したら数字を選んで全員からその数字を除去)
	CpuDifficulty             DaifugoCpuDifficulty // CPU難易度 (0=Normal, 1=Easy, 2=Hard)
}

// DefaultDaifugoConfig デフォルトのローカルルール設定 (全て有効)
func DefaultDaifugoConfig() DaifugoConfig {
	return DaifugoConfig{
		JokerCount:          DaifugoJokerCount,
		EightCutEnabled:     true,
		SuitLockMode:        DaifugoSuitLockFull,
		ElevenBackEnabled:   true,
		SequenceEnabled:     true,
		CardExchangeEnabled: true,
		FiveSkipEnabled:     false,
		FiveSkipCount:       1,
		SevenPassEnabled:    false,
		TenDiscardEnabled:   false,
		SpadeThreeEnabled:   false,
		CapitalFallEnabled:  false,
		NineReverseEnabled:  false,
		CoupDetatEnabled:    false,
		NumberLockEnabled:   false,
		SandstormEnabled:    false,
		EmperorEnabled:      false,
	}
}

// Validate 設定値のドメインバリデーション
func (c DaifugoConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(DaifugoDifficultyNormal), int(DaifugoDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateRange("joker count", c.JokerCount, 0, DaifugoJokerCount); err != nil {
		return err
	}
	if err := ValidateRange("suit lock mode", int(c.SuitLockMode), int(DaifugoSuitLockNone), int(DaifugoSuitLockFull)); err != nil {
		return err
	}
	if err := ValidateRange("five skip count", c.FiveSkipCount, 1, DaifugoFiveSkipCountMax); err != nil {
		return err
	}
	if c.SequenceLockEnabled && !c.SequenceEnabled {
		return fmt.Errorf("sequence lock requires sequence to be enabled")
	}
	if c.BlindExchangeEnabled && !c.CardExchangeEnabled {
		return fmt.Errorf("blind exchange requires card exchange to be enabled")
	}
	return nil
}
