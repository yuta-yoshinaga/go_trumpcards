package domain

// SevensConfig 7並べゲーム設定
type SevensConfig struct {
	TunnelEnabled          bool // トンネルルール (A↔K循環)
	TunnelSkipWidth        int  // カスタムトンネル: スキップ幅 (0=無効, 2以上で±N接続を追加。TunnelEnabled時は循環ラップあり)
	JokerCount             int  // ジョーカー枚数
	CpuStrategy            bool // CPU戦略思考
	MaxPasses              int  // 最大パス回数 (0 = 無制限)
	NoJokerFinish          bool // ジョーカー上がり禁止
	JokerReclaimEnabled    bool // ジョーカー回収 (ジョーカー配置位置に本物のカードを出すとジョーカーが手札に戻る)
	EndStopEnabled         bool // 片側ストップ (Aを置くと上側8-Kがブロック、Kを置くと下側A-6がブロック)
	JokerConsecutiveBanned bool // ジョーカー連続禁止 (前のターンにジョーカーを出した場合、次のターンにジョーカーを出せない)
}

// DefaultSevensConfig デフォルト設定 (全機能無効)
func DefaultSevensConfig() SevensConfig {
	return SevensConfig{
		TunnelEnabled:          false,
		TunnelSkipWidth:        0,
		JokerCount:             0,
		CpuStrategy:            false,
		MaxPasses:              SevensMaxPasses,
		NoJokerFinish:          false,
		JokerReclaimEnabled:    false,
		EndStopEnabled:         false,
		JokerConsecutiveBanned: false,
	}
}
