package domain

// EuchrePlayer ユーカープレイヤークラス
type EuchrePlayer struct {
	*GamePlayer
	TrickHolder
	team int // チームインデックス (0 or 1)
}

// NewEuchrePlayer コンストラクタ
func NewEuchrePlayer(isHuman bool, team int) *EuchrePlayer {
	return &EuchrePlayer{
		GamePlayer: NewGamePlayer(isHuman),
		team:       team,
	}
}

// GetTeam チームインデックスを取得
func (p *EuchrePlayer) GetTeam() int { return p.team }

// ResetRound ラウンドをリセット（トリック・手札・終了状態を初期化）
func (p *EuchrePlayer) ResetRound() {
	p.ResetTricks()
	p.Reset()
	p.SetIsFinished(false)
}
