package domain

import "math"

// BettingHumanProfile セッション内でベッティングゲーム(Poker/Holdem)における人間プレイヤーの行動を学習するプロファイル
type BettingHumanProfile struct {
	// AggressiveByBracket ハンド強度ブラケット別のアグレッシブ行動追跡:
	// [0]=弱(HighCard), [1]=中(OnePair/TwoPair), [2]=強(ThreeOfAKind+)
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

// bettingHandBracket ハンドランクをブラケット(0-2)に分類する
// 0=弱(HighCard), 1=中(OnePair/TwoPair), 2=強(ThreeOfAKind+)
func bettingHandBracket(handRank int) int {
	if handRank <= PokerHandHighCard {
		return 0
	}
	if handRank <= PokerHandTwoPair {
		return 1
	}
	return 2
}

// RecordAction 人間のベッティングアクションを記録する
// handRank: PokerHand* 定数, action: bettingAction* 定数
func (p *BettingHumanProfile) RecordAction(handRank, action int) {
	bracket := bettingHandBracket(handRank)
	p.AggressiveByBracket[bracket].Total++
	if action == bettingActionBet || action == bettingActionRaise {
		p.AggressiveByBracket[bracket].Aggressive++
	}
}

// RecordFoldToBet ベット/レイズに対するフォールド機会を記録する
func (p *BettingHumanProfile) RecordFoldToBet(folded bool) {
	p.FoldToBetTotal++
	if folded {
		p.FoldToBetCount++
	}
}

// RecordHesitation 迷い時間(ms)を記録する (Welford's online algorithm)
// ms <= 0 の場合は何もしない (CUI等で計測不可の場合)
func (p *BettingHumanProfile) RecordHesitation(ms int) {
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
func (p *BettingHumanProfile) BluffRate(bracket int) float64 {
	if bracket < 0 || bracket > 2 || p.AggressiveByBracket[bracket].Total == 0 {
		return 0.5
	}
	return float64(p.AggressiveByBracket[bracket].Aggressive) / float64(p.AggressiveByBracket[bracket].Total)
}

// FoldRate フォールド率を返す (データなしの場合0.5)
func (p *BettingHumanProfile) FoldRate() float64 {
	if p.FoldToBetTotal == 0 {
		return 0.5
	}
	return float64(p.FoldToBetCount) / float64(p.FoldToBetTotal)
}

// AdaptStrength 適応強度を返す (0.0 ~ 0.2)
func (p *BettingHumanProfile) AdaptStrength() float64 {
	games := p.GamesPlayed
	if games > metaAIMaxAdaptGames {
		games = metaAIMaxAdaptGames
	}
	return float64(games) * metaAIAdaptPerGame
}

// HesitationStdDev 迷い時間の標準偏差を返す (データ不足の場合0)
func (p *BettingHumanProfile) HesitationStdDev() float64 {
	if p.HesitationCount < 2 {
		return 0
	}
	return math.Sqrt(p.HesitationM2 / float64(p.HesitationCount-1))
}

// HesitationZScore 指定msの迷い時間のz-scoreを返す (データ不足の場合0)
func (p *BettingHumanProfile) HesitationZScore(ms int) float64 {
	sd := p.HesitationStdDev()
	if sd == 0 {
		return 0
	}
	return (float64(ms) - p.HesitationMean) / sd
}

// HesitationBoost 指定msの迷い時間に対するコール確率ブーストを返す
func (p *BettingHumanProfile) HesitationBoost(ms int) float64 {
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
// base: ベースコール確率, bracket: ハンド強度ブラケット, hesitationMs: 迷い時間(ms)
func (p *BettingHumanProfile) AdjustedCallChance(base float64, bracket int, hesitationMs int) float64 {
	adapt := p.AdaptStrength()
	return base + (p.BluffRate(bracket)-0.5)*adapt + p.HesitationBoost(hesitationMs)*adapt
}

// AdjustedBluffChance 人間のフォールド率に基づいて調整済みブラフ確率を返す
// 人間がフォールドしやすい → CPUがブラフを増やす
func (p *BettingHumanProfile) AdjustedBluffChance(base float64) float64 {
	return base * (1.0 + (p.FoldRate()-0.5)*p.AdaptStrength())
}
