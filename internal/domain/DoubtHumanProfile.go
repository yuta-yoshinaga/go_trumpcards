package domain

import (
	"encoding/json"
)

// DoubtHumanProfileBracketData はブラケット別ブラフ行動のJSON出力形式
type DoubtHumanProfileBracketData struct {
	Bluffs int `json:"bluffs"`
	Total  int `json:"total"`
}

// DoubtHumanProfileData はDoubtHumanProfileのJSON永続化形式
type DoubtHumanProfileData struct {
	BluffsByBracket [3]DoubtHumanProfileBracketData `json:"bluffsByBracket"`
	DoubtCorrect    int                             `json:"doubtCorrect"`
	DoubtTotal      int                             `json:"doubtTotal"`
	GamesPlayed     int                             `json:"gamesPlayed"`
	HesitationCount int                             `json:"hesitationCount"`
	HesitationMean  float64                         `json:"hesitationMean"`
	HesitationM2    float64                         `json:"hesitationM2"`
}

// Export プロファイルデータをJSON永続化形式でエクスポートする
func (p *DoubtHumanProfile) Export() DoubtHumanProfileData {
	var brackets [3]DoubtHumanProfileBracketData
	for i := 0; i < 3; i++ {
		brackets[i] = DoubtHumanProfileBracketData{
			Bluffs: p.BluffsByBracket[i].Bluffs,
			Total:  p.BluffsByBracket[i].Total,
		}
	}
	return DoubtHumanProfileData{
		BluffsByBracket: brackets,
		DoubtCorrect:    p.DoubtCorrect,
		DoubtTotal:      p.DoubtTotal,
		GamesPlayed:     p.GamesPlayed,
		HesitationCount: p.HesitationCount,
		HesitationMean:  p.HesitationMean,
		HesitationM2:    p.HesitationM2,
	}
}

// Import JSON永続化形式のデータからプロファイルを復元する
func (p *DoubtHumanProfile) Import(data DoubtHumanProfileData) {
	for i := 0; i < 3; i++ {
		p.BluffsByBracket[i].Bluffs = data.BluffsByBracket[i].Bluffs
		p.BluffsByBracket[i].Total = data.BluffsByBracket[i].Total
	}
	p.DoubtCorrect = data.DoubtCorrect
	p.DoubtTotal = data.DoubtTotal
	p.GamesPlayed = data.GamesPlayed
	p.HesitationCount = data.HesitationCount
	p.HesitationMean = data.HesitationMean
	p.HesitationM2 = data.HesitationM2
}

// ImportDoubtHumanProfileJSON JSONバイトからDoubtHumanProfileDataをデコードする
func ImportDoubtHumanProfileJSON(data []byte) (DoubtHumanProfileData, error) {
	var d DoubtHumanProfileData
	err := json.Unmarshal(data, &d)
	return d, err
}

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

// AdaptStrength 適応強度を返す (0.0 ~ 0.2)
func (p *DoubtHumanProfile) AdaptStrength() float64 {
	return computeAdaptStrength(p.GamesPlayed)
}

// RecordHesitation 迷い時間(ms)を記録する (Welford's online algorithm)
// ms <= 0 の場合は何もしない (CUI等で計測不可の場合)
func (p *DoubtHumanProfile) RecordHesitation(ms int) {
	welfordUpdate(&p.HesitationCount, &p.HesitationMean, &p.HesitationM2, ms)
}

// HesitationStdDev 迷い時間の標準偏差を返す (データ不足の場合0)
func (p *DoubtHumanProfile) HesitationStdDev() float64 {
	return welfordStdDev(p.HesitationCount, p.HesitationM2)
}

// HesitationZScore 指定msの迷い時間のz-scoreを返す (データ不足の場合0)
func (p *DoubtHumanProfile) HesitationZScore(ms int) float64 {
	return welfordZScore(p.HesitationCount, p.HesitationMean, p.HesitationM2, ms)
}

// HesitationBoost 指定msの迷い時間に対するダウト確率ブーストを返す
func (p *DoubtHumanProfile) HesitationBoost(ms int) float64 {
	return hesitationBoost(p.HesitationCount, p.HesitationMean, p.HesitationM2, ms)
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
