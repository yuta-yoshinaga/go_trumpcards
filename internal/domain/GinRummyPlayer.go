package domain

// GinRummyPlayer ジンラミープレイヤークラス
type GinRummyPlayer struct {
	*GamePlayer
	RoundScoreHolder
}

// NewGinRummyPlayer コンストラクタ
func NewGinRummyPlayer(isHuman bool) *GinRummyPlayer {
	return &GinRummyPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// ResetRound ラウンドをリセット（手札・スコア・終了状態を初期化）
func (p *GinRummyPlayer) ResetRound() {
	p.roundScore = 0
	p.Reset()
	p.SetIsFinished(false)
}
