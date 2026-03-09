package domain

// 迷い時間ディレイ (ミリ秒) の共通範囲定数
const (
	hesitationFastMin   = 300 // 速い反応 (ペア成立, ブラフ急ぎ)
	hesitationFastMax   = 500
	hesitationMediumMin = 600 // 中間の反応 (正直, 通常)
	hesitationMediumMax = 1000
)
