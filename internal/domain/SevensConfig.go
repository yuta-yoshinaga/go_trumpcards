package domain

// SevensConfig 7並べゲーム設定
type SevensConfig struct {
	TunnelEnabled       bool // トンネルルール (A↔K循環)
	JokerCount          int  // ジョーカー枚数
	CpuStrategy         bool // CPU戦略思考
	MaxPasses           int  // 最大パス回数 (0 = 無制限)
	NoJokerFinish       bool // ジョーカー上がり禁止
	JokerReclaimEnabled bool // ジョーカー回収 (ジョーカー配置位置に本物のカードを出すとジョーカーが手札に戻る)
	EndStopEnabled      bool // 片側ストップ (Aを置くと上側8-Kがブロック、Kを置くと下側A-6がブロック)
}

// DefaultSevensConfig デフォルト設定 (全機能無効)
func DefaultSevensConfig() SevensConfig {
	return SevensConfig{
		TunnelEnabled: false,
		JokerCount:    0,
		CpuStrategy:   false,
		MaxPasses:     SevensMaxPasses,
		NoJokerFinish: false,
	}
}
