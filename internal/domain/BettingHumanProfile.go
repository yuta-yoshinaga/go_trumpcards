//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
)

// BettingHumanProfileBracketData はブラケット別アグレッシブ行動のJSON出力形式
type BettingHumanProfileBracketData struct {
	Aggressive int `json:"aggressive"`
	Total      int `json:"total"`
}

// BettingHumanProfileData はBettingHumanProfileのJSON永続化形式
type BettingHumanProfileData struct {
	AggressiveByBracket [3]BettingHumanProfileBracketData `json:"aggressiveByBracket"`
	FoldToBetCount      int                               `json:"foldToBetCount"`
	FoldToBetTotal      int                               `json:"foldToBetTotal"`
	GamesPlayed         int                               `json:"gamesPlayed"`
	HesitationCount     int                               `json:"hesitationCount"`
	HesitationMean      float64                           `json:"hesitationMean"`
	HesitationM2        float64                           `json:"hesitationM2"`
}

// Export プロファイルデータをJSON永続化形式でエクスポートする
func (p *BettingHumanProfile) Export() BettingHumanProfileData {
	var brackets [3]BettingHumanProfileBracketData
	for i := 0; i < 3; i++ {
		brackets[i] = BettingHumanProfileBracketData{
			Aggressive: p.AggressiveByBracket[i].Aggressive,
			Total:      p.AggressiveByBracket[i].Total,
		}
	}
	return BettingHumanProfileData{
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
func (p *BettingHumanProfile) Import(data BettingHumanProfileData) {
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

// ImportBettingHumanProfileJSON JSONバイトからBettingHumanProfileDataをデコードする
func ImportBettingHumanProfileJSON(data []byte) (BettingHumanProfileData, error) {
	var d BettingHumanProfileData
	err := json.Unmarshal(data, &d)
	return d, err
}

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
	welfordUpdate(&p.HesitationCount, &p.HesitationMean, &p.HesitationM2, ms)
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
	return computeAdaptStrength(p.GamesPlayed)
}

// HesitationStdDev 迷い時間の標準偏差を返す (データ不足の場合0)
func (p *BettingHumanProfile) HesitationStdDev() float64 {
	return welfordStdDev(p.HesitationCount, p.HesitationM2)
}

// HesitationZScore 指定msの迷い時間のz-scoreを返す (データ不足の場合0)
func (p *BettingHumanProfile) HesitationZScore(ms int) float64 {
	return welfordZScore(p.HesitationCount, p.HesitationMean, p.HesitationM2, ms)
}

// HesitationBoost 指定msの迷い時間に対するコール確率ブーストを返す
func (p *BettingHumanProfile) HesitationBoost(ms int) float64 {
	return hesitationBoost(p.HesitationCount, p.HesitationMean, p.HesitationM2, ms)
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

// importBettingProfile decodes a saved human profile, returning nil for empty
// input so callers can leave their existing profile untouched. 8 betting games
// had this written out.
//
// Lives here rather than in player_helpers.go because that file is compiled
// into every Worker, while this one is tagged `!js || !wasm || casino` -- the
// split that lets TinyGo drop the betting games from the other five binaries.
// A helper naming BettingHumanProfile from an untagged file breaks those
// builds, which `go test ./...` cannot show because it compiles everything.
func importBettingProfile(data []byte) (*BettingHumanProfile, error) {
	if len(data) == 0 {
		return nil, nil
	}
	d, err := ImportBettingHumanProfileJSON(data)
	if err != nil {
		return nil, err
	}
	p := &BettingHumanProfile{}
	p.Import(d)
	return p, nil
}
