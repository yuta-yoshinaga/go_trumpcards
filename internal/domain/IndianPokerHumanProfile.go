package domain

import (
	"encoding/json"
)

// IndianPokerHumanProfileBracketData はブラケット別アグレッシブ行動のJSON出力形式
type IndianPokerHumanProfileBracketData struct {
	Aggressive int `json:"aggressive"`
	Total      int `json:"total"`
}

// IndianPokerHumanProfileData はIndianPokerHumanProfileのJSON永続化形式
type IndianPokerHumanProfileData struct {
	AggressiveByBracket [3]IndianPokerHumanProfileBracketData `json:"aggressiveByBracket"`
	FoldToBetCount      int                                   `json:"foldToBetCount"`
	FoldToBetTotal      int                                   `json:"foldToBetTotal"`
	GamesPlayed         int                                   `json:"gamesPlayed"`
	HesitationCount     int                                   `json:"hesitationCount"`
	HesitationMean      float64                               `json:"hesitationMean"`
	HesitationM2        float64                               `json:"hesitationM2"`
}

// Export プロファイルデータをJSON永続化形式でエクスポートする
func (p *IndianPokerHumanProfile) Export() IndianPokerHumanProfileData {
	var brackets [3]IndianPokerHumanProfileBracketData
	for i := 0; i < 3; i++ {
		brackets[i] = IndianPokerHumanProfileBracketData{
			Aggressive: p.AggressiveByBracket[i].Aggressive,
			Total:      p.AggressiveByBracket[i].Total,
		}
	}
	return IndianPokerHumanProfileData{
		AggressiveByBracket: brackets,
		FoldToBetCount:      p.FoldToBetCount,
		FoldToBetTotal:      p.FoldToBetTotal,
		GamesPlayed:         p.GamesPlayed,
		HesitationCount:     p.HesitationCount,
		HesitationMean:      p.HesitationMean,
		HesitationM2:        p.HesitationM2,
	}
}

// Import JSON永続化形式のデータからプロファイルを復元する
func (p *IndianPokerHumanProfile) Import(data IndianPokerHumanProfileData) {
	for i := 0; i < 3; i++ {
		p.AggressiveByBracket[i].Aggressive = data.AggressiveByBracket[i].Aggressive
		p.AggressiveByBracket[i].Total = data.AggressiveByBracket[i].Total
	}
	p.FoldToBetCount = data.FoldToBetCount
	p.FoldToBetTotal = data.FoldToBetTotal
	p.GamesPlayed = data.GamesPlayed
	p.HesitationCount = data.HesitationCount
	p.HesitationMean = data.HesitationMean
	p.HesitationM2 = data.HesitationM2
}

// ImportIndianPokerHumanProfileJSON JSONバイトからIndianPokerHumanProfileDataをデコードする
func ImportIndianPokerHumanProfileJSON(data []byte) (IndianPokerHumanProfileData, error) {
	var d IndianPokerHumanProfileData
	err := json.Unmarshal(data, &d)
	return d, err
}

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
	welfordUpdate(ms, &p.HesitationCount, &p.HesitationMean, &p.HesitationM2)
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
	return computeAdaptStrength(p.GamesPlayed)
}

// HesitationStdDev 迷い時間の標準偏差を返す (データ不足の場合0)
func (p *IndianPokerHumanProfile) HesitationStdDev() float64 {
	return welfordStdDev(p.HesitationCount, p.HesitationM2)
}

// HesitationZScore 指定msの迷い時間のz-scoreを返す (データ不足の場合0)
func (p *IndianPokerHumanProfile) HesitationZScore(ms int) float64 {
	return welfordZScore(ms, p.HesitationCount, p.HesitationMean, p.HesitationM2)
}

// HesitationBoost 指定msの迷い時間に対するコール確率ブーストを返す
func (p *IndianPokerHumanProfile) HesitationBoost(ms int) float64 {
	return hesitationBoost(p.HesitationCount, p.HesitationMean, p.HesitationM2, ms)
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
