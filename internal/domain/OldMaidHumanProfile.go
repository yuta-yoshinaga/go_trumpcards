package domain

import "math/rand"

// OldMaidHumanProfile セッション内で人間プレイヤーの行動を学習するプロファイル
type OldMaidHumanProfile struct {
	// PositionBuckets 位置別のピック追跡: [0]=先頭1/3, [1]=中央1/3, [2]=末尾1/3
	PositionBuckets [3]int
	// TotalPicks 合計ピック数
	TotalPicks int
	// ShuffleCount シャッフル回数
	ShuffleCount int
	// DrawCount 引いた回数
	DrawCount int
	// GamesPlayed セッション内のゲーム数
	GamesPlayed int
}

// oldMaidPickBucket カードインデックスを位置バケット(0-2)に分類する
func oldMaidPickBucket(cardIdx, handSize int) int {
	if handSize <= 0 {
		return 0
	}
	third := handSize / 3
	if third == 0 {
		third = 1
	}
	if cardIdx < third {
		return 0
	}
	if cardIdx < third*2 {
		return 1
	}
	return 2
}

// RecordPick カードピックを記録する
func (p *OldMaidHumanProfile) RecordPick(cardIdx, handSize int) {
	bucket := oldMaidPickBucket(cardIdx, handSize)
	p.PositionBuckets[bucket]++
	p.TotalPicks++
}

// RecordShuffle シャッフルを記録する
func (p *OldMaidHumanProfile) RecordShuffle() {
	p.ShuffleCount++
}

// RecordDraw 引きを記録する
func (p *OldMaidHumanProfile) RecordDraw() {
	p.DrawCount++
}

// PickRate 指定バケットのピック率を返す (データなしの場合1/3)
func (p *OldMaidHumanProfile) PickRate(bucket int) float64 {
	if bucket < 0 || bucket > 2 || p.TotalPicks == 0 {
		return 1.0 / 3.0
	}
	return float64(p.PositionBuckets[bucket]) / float64(p.TotalPicks)
}

// ShuffleRate シャッフル率を返す (draws == 0 の場合0.0)
func (p *OldMaidHumanProfile) ShuffleRate() float64 {
	if p.DrawCount == 0 {
		return 0.0
	}
	return float64(p.ShuffleCount) / float64(p.DrawCount)
}

// AdaptStrength 適応強度を返す (0.0 ~ 0.2)
func (p *OldMaidHumanProfile) AdaptStrength() float64 {
	games := p.GamesPlayed
	if games > metaAIMaxAdaptGames {
		games = metaAIMaxAdaptGames
	}
	return float64(games) * metaAIAdaptPerGame
}

// oldMaidHighShuffleThreshold シャッフル率が高いとみなす閾値
const oldMaidHighShuffleThreshold = 0.5

// StrategicPlacement 奇数カードの戦略的配置位置を返す
// 最もピックされにくいゾーンに配置する。
// 適応強度が低い場合やシャッフル率が高い場合はランダムな端に配置する。
func (p *OldMaidHumanProfile) StrategicPlacement(size int) int {
	if size <= 1 {
		return 0
	}

	// シャッフル率が高い場合はランダムな端にフォールバック
	if p.ShuffleRate() >= oldMaidHighShuffleThreshold {
		if rand.Intn(2) == 0 {
			return 0
		}
		return size - 1
	}

	// 最もピックされにくいバケットを見つける
	minRate := p.PickRate(0)
	minBucket := 0
	for i := 1; i <= 2; i++ {
		rate := p.PickRate(i)
		if rate < minRate {
			minRate = rate
			minBucket = i
		}
	}

	// バケットに応じた位置を返す
	third := size / 3
	if third == 0 {
		third = 1
	}
	switch minBucket {
	case 0:
		return 0
	case 1:
		return third + rand.Intn(max(third, 1))
	default:
		return size - 1
	}
}
