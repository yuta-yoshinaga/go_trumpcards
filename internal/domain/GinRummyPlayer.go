package domain

// GinRummyPlayer ジンラミープレイヤークラス
type GinRummyPlayer struct {
	*GamePlayer
	roundScore      int // 現在のラウンドのスコア
	cumulativeScore int // 累積スコア
}

// NewGinRummyPlayer コンストラクタ
func NewGinRummyPlayer(isHuman bool) *GinRummyPlayer {
	return &GinRummyPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// GetRoundScore 現在のラウンドスコアを取得
func (p *GinRummyPlayer) GetRoundScore() int { return p.roundScore }

// SetRoundScore 現在のラウンドスコアを設定
func (p *GinRummyPlayer) SetRoundScore(score int) { p.roundScore = score }

// GetCumulativeScore 累積スコアを取得
func (p *GinRummyPlayer) GetCumulativeScore() int { return p.cumulativeScore }

// SetCumulativeScore 累積スコアを設定
func (p *GinRummyPlayer) SetCumulativeScore(score int) { p.cumulativeScore = score }

// CommitRoundScore ラウンドスコアを累積スコアに加算
func (p *GinRummyPlayer) CommitRoundScore() {
	p.cumulativeScore += p.roundScore
}

// ResetRound ラウンドをリセット（手札・スコア・終了状態を初期化）
func (p *GinRummyPlayer) ResetRound() {
	p.roundScore = 0
	p.Reset()
	p.SetIsFinished(false)
}
