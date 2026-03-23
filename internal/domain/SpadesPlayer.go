package domain

// SpadesPlayer スペードプレイヤークラス
type SpadesPlayer struct {
	*GamePlayer
	RoundScoreHolder
	TrickHolder
	bid  int // 宣言したトリック数 (-1 = 未ビッド)
	bags int // 累積バッグ数 (オーバートリック)
}

// NewSpadesPlayer コンストラクタ
func NewSpadesPlayer(isHuman bool) *SpadesPlayer {
	return &SpadesPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		bid:        -1,
	}
}

// GetBid ビッド取得 (-1 = 未ビッド)
func (p *SpadesPlayer) GetBid() int { return p.bid }

// SetBid ビッド設定
func (p *SpadesPlayer) SetBid(bid int) { p.bid = bid }

// GetBags 累積バッグ数を取得
func (p *SpadesPlayer) GetBags() int { return p.bags }

// SetBags バッグ数を設定
func (p *SpadesPlayer) SetBags(bags int) { p.bags = bags }

// ResetRound ラウンドをリセット（ビッド・トリック・手札・終了状態を初期化）
func (p *SpadesPlayer) ResetRound() {
	p.bid = -1
	p.roundScore = 0
	p.tricksTaken = nil
	p.Reset()
	p.SetIsFinished(false)
}
