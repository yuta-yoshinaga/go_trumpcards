package domain

// CrazyEightsPlayer クレイジーエイトプレイヤークラス
type CrazyEightsPlayer struct {
	*GamePlayer
	RoundScoreHolder
}

// NewCrazyEightsPlayer コンストラクタ
func NewCrazyEightsPlayer(isHuman bool) *CrazyEightsPlayer {
	return &CrazyEightsPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// ResetRound ラウンドをリセット（手札・スコア・終了状態を初期化）
func (p *CrazyEightsPlayer) ResetRound() {
	p.SetRoundScore(0)
	p.Reset()
	p.SetIsFinished(false)
}
