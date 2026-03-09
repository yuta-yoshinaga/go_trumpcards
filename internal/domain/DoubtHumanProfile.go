package domain

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

// AdjustedDoubtChance ブラフ率に基づいて調整済みダウト確率を返す
func (p *DoubtHumanProfile) AdjustedDoubtChance(base float64, bracket int) float64 {
	return base + (p.BluffRate(bracket)-0.5)*p.AdaptStrength()
}

// AdjustedBluffChance 人間のダウト正解率に基づいて調整済みブラフ確率を返す
func (p *DoubtHumanProfile) AdjustedBluffChance(base float64) float64 {
	return base * (1.0 - p.DoubtAccuracy()*p.AdaptStrength())
}
