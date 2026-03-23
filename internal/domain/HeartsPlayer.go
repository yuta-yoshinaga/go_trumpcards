package domain

// HeartsPlayer ハーツプレイヤークラス
type HeartsPlayer struct {
	*GamePlayer
	RoundScoreHolder
	TrickHolder
}

// NewHeartsPlayer コンストラクタ
func NewHeartsPlayer(isHuman bool) *HeartsPlayer {
	return &HeartsPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// ResetRound ラウンドをリセット（スコア・トリック・手札・終了状態を初期化）
func (p *HeartsPlayer) ResetRound() {
	p.SetRoundScore(0)
	p.ResetTricks()
	p.Reset()
	p.SetIsFinished(false)
}
