package domain

// SpadesPlayer スペードプレイヤークラス
type SpadesPlayer struct {
	*GamePlayer
	bid             int       // 宣言したトリック数 (-1 = 未ビッド)
	tricksTaken     [][]*Card // 獲得したトリック
	roundScore      int       // 現在のラウンドのスコア
	cumulativeScore int       // 累積スコア
	bags            int       // 累積バッグ数 (オーバートリック)
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

// GetTricksTaken 獲得したトリック一覧を取得
func (p *SpadesPlayer) GetTricksTaken() [][]*Card { return p.tricksTaken }

// GetTrickCount 獲得したトリック数を取得
func (p *SpadesPlayer) GetTrickCount() int { return len(p.tricksTaken) }

// AddTrick トリックを追加
func (p *SpadesPlayer) AddTrick(cards []*Card) {
	p.tricksTaken = append(p.tricksTaken, cards)
}

// GetRoundScore 現在のラウンドスコアを取得
func (p *SpadesPlayer) GetRoundScore() int { return p.roundScore }

// SetRoundScore 現在のラウンドスコアを設定
func (p *SpadesPlayer) SetRoundScore(score int) { p.roundScore = score }

// GetCumulativeScore 累積スコアを取得
func (p *SpadesPlayer) GetCumulativeScore() int { return p.cumulativeScore }

// SetCumulativeScore 累積スコアを設定
func (p *SpadesPlayer) SetCumulativeScore(score int) { p.cumulativeScore = score }

// GetBags 累積バッグ数を取得
func (p *SpadesPlayer) GetBags() int { return p.bags }

// SetBags バッグ数を設定
func (p *SpadesPlayer) SetBags(bags int) { p.bags = bags }

// CommitRoundScore ラウンドスコアを累積スコアに加算
func (p *SpadesPlayer) CommitRoundScore() {
	p.cumulativeScore += p.roundScore
}

// ResetRound ラウンドをリセット（ビッド・トリック・手札・終了状態を初期化）
func (p *SpadesPlayer) ResetRound() {
	p.bid = -1
	p.roundScore = 0
	p.tricksTaken = nil
	p.Reset()
	p.SetIsFinished(false)
}
