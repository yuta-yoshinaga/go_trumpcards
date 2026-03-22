package domain

import "math"

// IndianPokerHumanProfile セッション内でインディアンポーカーにおける人間プレイヤーの行動を学習するプロファイル
// CPUは人間のカードが見えるため、人間のカード強度ブラケット別にアグレッシブ行動を追跡する
type IndianPokerHumanProfile struct {
	// AggressiveByBracket カード強度ブラケット別のアグレッシブ行動追跡:
	// [0]=弱(2-5), [1]=中(6-9), [2]=強(10-A)
	AggressiveByBracket [3]struct{ Aggressive, Total int }
	// FoldToBetCount ベット/レイズに対してフォールドした回数
	FoldToBetCount int
	// FoldToBetTotal ベット/レイズに対するフォールド機会の合計
	FoldToBetTotal int
	// GamesPlayed セッション内のゲーム数
	GamesPlayed int
	// HesitationCount 迷い時間の計測回数 (Welford's online algorithm)
	HesitationCount int
	// HesitationMean 迷い時間の平均 (ms)
	HesitationMean float64
	// HesitationM2 迷い時間の分散計算用 M2 (Welford's online algorithm)
	HesitationM2 float64
}

// indianPokerCardBracket カードランクをブラケット(0-2)に分類する
// 0=弱(2-5), 1=中(6-9), 2=強(10-A=14)
func indianPokerCardBracket(cardRank int) int {
	if cardRank <= 5 {
		return 0
	}
	if cardRank <= 9 {
		return 1
	}
	return 2
}

// RecordAction 人間のベッティングアクションを記録する
// cardRank: カードランク(2-14), action: bettingAction* 定数
func (p *IndianPokerHumanProfile) RecordAction(cardRank, action int) {
	bracket := indianPokerCardBracket(cardRank)
	p.AggressiveByBracket[bracket].Total++
	if action == bettingActionBet || action == bettingActionRaise {
		p.AggressiveByBracket[bracket].Aggressive++
	}
}

// RecordFoldToBet ベット/レイズに対するフォールド機会を記録する
func (p *IndianPokerHumanProfile) RecordFoldToBet(folded bool) {
	p.FoldToBetTotal++
	if folded {
		p.FoldToBetCount++
	}
}

// RecordHesitation 迷い時間(ms)を記録する (Welford's online algorithm)
// ms <= 0 の場合は何もしない (CUI等で計測不可の場合)
func (p *IndianPokerHumanProfile) RecordHesitation(ms int) {
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

// BluffRate 指定ブラケットのアグレッシブ率を返す (データなしの場合0.5)
func (p *IndianPokerHumanProfile) BluffRate(bracket int) float64 {
	if bracket < 0 || bracket > 2 || p.AggressiveByBracket[bracket].Total == 0 {
		return 0.5
	}
	return float64(p.AggressiveByBracket[bracket].Aggressive) / float64(p.AggressiveByBracket[bracket].Total)
}

// FoldRate フォールド率を返す (データなしの場合0.5)
func (p *IndianPokerHumanProfile) FoldRate() float64 {
	if p.FoldToBetTotal == 0 {
		return 0.5
	}
	return float64(p.FoldToBetCount) / float64(p.FoldToBetTotal)
}

// AdaptStrength 適応強度を返す (0.0 ~ 0.2)
func (p *IndianPokerHumanProfile) AdaptStrength() float64 {
	games := min(p.GamesPlayed, metaAIMaxAdaptGames)
	return float64(games) * metaAIAdaptPerGame
}

// HesitationStdDev 迷い時間の標準偏差を返す (データ不足の場合0)
func (p *IndianPokerHumanProfile) HesitationStdDev() float64 {
	if p.HesitationCount < 2 {
		return 0
	}
	return math.Sqrt(p.HesitationM2 / float64(p.HesitationCount-1))
}

// HesitationZScore 指定msの迷い時間のz-scoreを返す (データ不足の場合0)
func (p *IndianPokerHumanProfile) HesitationZScore(ms int) float64 {
	sd := p.HesitationStdDev()
	if sd == 0 {
		return 0
	}
	return (float64(ms) - p.HesitationMean) / sd
}

// HesitationBoost 指定msの迷い時間に対するコール確率ブーストを返す
func (p *IndianPokerHumanProfile) HesitationBoost(ms int) float64 {
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

// AdjustedCallChance ブラフ率と迷い時間に基づいて調整済みコール確率を返す
// base: ベースコール確率, bracket: カード強度ブラケット, hesitationMs: 迷い時間(ms)
func (p *IndianPokerHumanProfile) AdjustedCallChance(base float64, bracket int, hesitationMs int) float64 {
	adapt := p.AdaptStrength()
	return base + (p.BluffRate(bracket)-0.5)*adapt + p.HesitationBoost(hesitationMs)*adapt
}

// AdjustedBluffChance 人間のフォールド率に基づいて調整済みブラフ確率を返す
// 人間がフォールドしやすい → CPUがブラフを増やす
func (p *IndianPokerHumanProfile) AdjustedBluffChance(base float64) float64 {
	return base * (1.0 + (p.FoldRate()-0.5)*p.AdaptStrength())
}
