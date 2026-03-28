package domain

// CPU戦略モード定数
const (
	SevensCpuSimple     int = iota // シンプル (最初の出せるカード)
	SevensCpuStrategic             // 戦略的 (自己利益重視)
	SevensCpuHarassment            // 嫌がらせ特化 (相手妨害重視)
)

// SevensConfig 7並べゲーム設定
type SevensConfig struct {
	TunnelEnabled          bool `json:"te"` // トンネルルール (A↔K循環)
	TunnelSkipWidth        int  `json:"tw"` // カスタムトンネル: スキップ幅 (0=無効, 2以上で±N接続を追加。TunnelEnabled時は循環ラップあり)
	JokerCount             int  `json:"jc"` // ジョーカー枚数
	CpuStrategy            int  `json:"cs"` // CPU戦略モード (0=シンプル, 1=戦略的, 2=嫌がらせ特化)
	MaxPasses              int  `json:"mp"` // 最大パス回数 (0 = 無制限)
	NoJokerFinish          bool `json:"nj"` // ジョーカー上がり禁止
	JokerReclaimEnabled    bool `json:"jr"` // ジョーカー回収 (ジョーカー配置位置に本物のカードを出すとジョーカーが手札に戻る)
	EndStopEnabled         bool `json:"es"` // 片側ストップ (Aを置くと上側8-Kがブロック、Kを置くと下側A-6がブロック)
	JokerConsecutiveBanned bool `json:"jb"` // ジョーカー連続禁止 (前のターンにジョーカーを出した場合、次のターンにジョーカーを出せない)
}

// DefaultSevensConfig デフォルト設定 (全機能無効)
func DefaultSevensConfig() SevensConfig {
	return SevensConfig{
		TunnelEnabled:          false,
		TunnelSkipWidth:        0,
		JokerCount:             0,
		CpuStrategy:            SevensCpuSimple,
		MaxPasses:              SevensMaxPasses,
		NoJokerFinish:          false,
		JokerReclaimEnabled:    false,
		EndStopEnabled:         false,
		JokerConsecutiveBanned: false,
	}
}
