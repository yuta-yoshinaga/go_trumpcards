package domain

// CribbagePlayer クリベッジプレイヤークラス
type CribbagePlayer struct {
	*GamePlayer
	RoundScoreHolder
}

// NewCribbagePlayer コンストラクタ
func NewCribbagePlayer(isHuman bool) *CribbagePlayer {
	return &CribbagePlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// ResetRound ラウンドをリセット（手札・スコア・終了状態を初期化）
func (p *CribbagePlayer) ResetRound() {
	p.SetRoundScore(0)
	p.Reset()
	p.SetIsFinished(false)
}
