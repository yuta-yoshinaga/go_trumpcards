package domain

// CrazyEightsPlayer クレイジーエイトプレイヤークラス
type CrazyEightsPlayer struct {
	*GamePlayer
	roundScore      int // 現在のラウンドのスコア
	cumulativeScore int // 累積スコア
}

// NewCrazyEightsPlayer コンストラクタ
func NewCrazyEightsPlayer(isHuman bool) *CrazyEightsPlayer {
	return &CrazyEightsPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// GetRoundScore 現在のラウンドスコアを取得
func (p *CrazyEightsPlayer) GetRoundScore() int { return p.roundScore }

// SetRoundScore 現在のラウンドスコアを設定
func (p *CrazyEightsPlayer) SetRoundScore(score int) { p.roundScore = score }

// GetCumulativeScore 累積スコアを取得
func (p *CrazyEightsPlayer) GetCumulativeScore() int { return p.cumulativeScore }

// SetCumulativeScore 累積スコアを設定
func (p *CrazyEightsPlayer) SetCumulativeScore(score int) { p.cumulativeScore = score }

// CommitRoundScore ラウンドスコアを累積スコアに加算
func (p *CrazyEightsPlayer) CommitRoundScore() {
	p.cumulativeScore += p.roundScore
}

// ResetRound ラウンドをリセット（手札・スコア・終了状態を初期化）
func (p *CrazyEightsPlayer) ResetRound() {
	p.roundScore = 0
	p.Reset()
	p.SetIsFinished(false)
}
