package domain

import "math"

// DoubtHumanProfile セッション内で人間プレイヤーの行動を学習するプロファイル
type DoubtHumanProfile struct {
	// BluffsByBracket 手札枚数ブラケット別のブラフ追跡: [0]=小(1-4), [1]=中(5-9), [2]=大(10+)
	BluffsByBracket [3]struct{ Bluffs, Total int }
	// DoubtCorrect 人間のダウト成功回数
	DoubtCorrect int
	// DoubtTotal 人間のダウト合計回数
	DoubtTotal int
	// GamesPlayed セッション内のゲーム数
	GamesPlayed int
	// HesitationCount 迷い時間の計測回数 (Welford's online algorithm)
	HesitationCount int
	// HesitationMean 迷い時間の平均 (ms)
	HesitationMean float64
	// HesitationM2 迷い時間の分散計算用 M2 (Welford's online algorithm)
	HesitationM2 float64
}

// hesitationMinPlays 迷い時間ブーストを有効にするための最小データ点数
const hesitationMinPlays = 3

// hesitationZThreshold z-scoreがこの値を超えた場合にブーストが発生する
const hesitationZThreshold = 1.0

// hesitationWeight z-score超過分に対するブースト重み
const hesitationWeight = 0.05

// maxHesitationBoost 迷い時間ブーストの上限
const maxHesitationBoost = 0.10

// doubtHandSizeBracket 手札枚数をブラケット(0-2)に分類する
func doubtHandSizeBracket(handSize int) int {
	if handSize <= 4 {
		return 0
	}
	if handSize <= 9 {
		return 1
	}
	return 2
}

// RecordPlay カードプレイを記録する
func (p *DoubtHumanProfile) RecordPlay(handSize int, isBluff bool) {
	bracket := doubtHandSizeBracket(handSize)
	p.BluffsByBracket[bracket].Total++
	if isBluff {
		p.BluffsByBracket[bracket].Bluffs++
	}
}

// RecordDoubt ダウト結果を記録する
func (p *DoubtHumanProfile) RecordDoubt(wasCorrect bool) {
	p.DoubtTotal++
	if wasCorrect {
		p.DoubtCorrect++
	}
}

// BluffRate 指定ブラケットのブラフ率を返す (データなしの場合0.5)
func (p *DoubtHumanProfile) BluffRate(bracket int) float64 {
	if bracket < 0 || bracket > 2 || p.BluffsByBracket[bracket].Total == 0 {
		return 0.5
	}
	return float64(p.BluffsByBracket[bracket].Bluffs) / float64(p.BluffsByBracket[bracket].Total)
}

// DoubtAccuracy 人間のダウト正解率を返す (データなしの場合0.5)
func (p *DoubtHumanProfile) DoubtAccuracy() float64 {
	if p.DoubtTotal == 0 {
		return 0.5
	}
	return float64(p.DoubtCorrect) / float64(p.DoubtTotal)
}

// metaAIMaxAdaptGames 適応が最大に達するゲーム数
const metaAIMaxAdaptGames = 5

// metaAIAdaptPerGame 1ゲームあたりの適応強度
const metaAIAdaptPerGame = 0.04

// AdaptStrength 適応強度を返す (0.0 ~ 0.2)
func (p *DoubtHumanProfile) AdaptStrength() float64 {
	games := p.GamesPlayed
	if games > metaAIMaxAdaptGames {
		games = metaAIMaxAdaptGames
	}
	return float64(games) * metaAIAdaptPerGame
}

// RecordHesitation 迷い時間(ms)を記録する (Welford's online algorithm)
// ms <= 0 の場合は何もしない (CUI等で計測不可の場合)
func (p *DoubtHumanProfile) RecordHesitation(ms int) {
	if ms <= 0 {
		return
	}
	p.HesitationCount++
	x := float64(ms)
	delta := x - p.HesitationMean
	p.HesitationMean += delta / float64(p.HesitationCount)
	delta2 := x - p.HesitationMean
	p.HesitationM2 += delta * delta2
}

// HesitationStdDev 迷い時間の標準偏差を返す (データ不足の場合0)
func (p *DoubtHumanProfile) HesitationStdDev() float64 {
	if p.HesitationCount < 2 {
		return 0
	}
	return math.Sqrt(p.HesitationM2 / float64(p.HesitationCount))
}

// HesitationZScore 指定msの迷い時間のz-scoreを返す (データ不足の場合0)
func (p *DoubtHumanProfile) HesitationZScore(ms int) float64 {
	sd := p.HesitationStdDev()
	if sd == 0 {
		return 0
	}
	return (float64(ms) - p.HesitationMean) / sd
}

// HesitationBoost 指定msの迷い時間に対するダウト確率ブーストを返す
func (p *DoubtHumanProfile) HesitationBoost(ms int) float64 {
	if p.HesitationCount < hesitationMinPlays {
		return 0
	}
	z := p.HesitationZScore(ms)
	if z <= hesitationZThreshold {
		return 0
	}
	boost := (z - hesitationZThreshold) * hesitationWeight
	if boost > maxHesitationBoost {
		return maxHesitationBoost
	}
	return boost
}

// AdjustedDoubtChance ブラフ率と迷い時間に基づいて調整済みダウト確率を返す
func (p *DoubtHumanProfile) AdjustedDoubtChance(base float64, bracket int, humanPlayMs int) float64 {
	adapt := p.AdaptStrength()
	return base + (p.BluffRate(bracket)-0.5)*adapt + p.HesitationBoost(humanPlayMs)*adapt
}

// AdjustedBluffChance 人間のダウト正解率に基づいて調整済みブラフ確率を返す
func (p *DoubtHumanProfile) AdjustedBluffChance(base float64) float64 {
	return base * (1.0 - p.DoubtAccuracy()*p.AdaptStrength())
}
