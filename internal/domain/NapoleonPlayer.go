package domain

// NapoleonPlayer ナポレオンプレイヤークラス
type NapoleonPlayer struct {
	*GamePlayer
	bid              int       // 宣言した絵札数 (-1 = 未ビッド, 0 = パス)
	isNapoleon       bool      // ナポレオンかどうか
	isAdjutant       bool      // 副官かどうか
	adjutantRevealed bool      // 副官が公開されたかどうか
	pictureCards     int       // 獲得した絵札数 (今ラウンド)
	tricksTaken      [][]*Card // 獲得したトリック
	roundScore       int       // 現在のラウンドのスコア
	cumulativeScore  int       // 累積スコア
}

// NewNapoleonPlayer コンストラクタ
func NewNapoleonPlayer(isHuman bool) *NapoleonPlayer {
	return &NapoleonPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		bid:        -1,
	}
}

// GetBid ビッド取得 (-1 = 未ビッド, 0 = パス)
func (p *NapoleonPlayer) GetBid() int { return p.bid }

// SetBid ビッド設定
func (p *NapoleonPlayer) SetBid(bid int) { p.bid = bid }

// GetIsNapoleon ナポレオンかどうか
func (p *NapoleonPlayer) GetIsNapoleon() bool { return p.isNapoleon }

// SetIsNapoleon ナポレオン設定
func (p *NapoleonPlayer) SetIsNapoleon(v bool) { p.isNapoleon = v }

// GetIsAdjutant 副官かどうか
func (p *NapoleonPlayer) GetIsAdjutant() bool { return p.isAdjutant }

// SetIsAdjutant 副官設定
func (p *NapoleonPlayer) SetIsAdjutant(v bool) { p.isAdjutant = v }

// GetAdjutantRevealed 副官公開状態
func (p *NapoleonPlayer) GetAdjutantRevealed() bool { return p.adjutantRevealed }

// SetAdjutantRevealed 副官公開状態設定
func (p *NapoleonPlayer) SetAdjutantRevealed(v bool) { p.adjutantRevealed = v }

// GetPictureCards 獲得した絵札数
func (p *NapoleonPlayer) GetPictureCards() int { return p.pictureCards }

// SetPictureCards 絵札数設定
func (p *NapoleonPlayer) SetPictureCards(n int) { p.pictureCards = n }

// GetTricksTaken 獲得したトリック一覧を取得
func (p *NapoleonPlayer) GetTricksTaken() [][]*Card { return p.tricksTaken }

// GetTrickCount 獲得したトリック数を取得
func (p *NapoleonPlayer) GetTrickCount() int { return len(p.tricksTaken) }

// AddTrick トリックを追加
func (p *NapoleonPlayer) AddTrick(cards []*Card) {
	p.tricksTaken = append(p.tricksTaken, cards)
}

// GetRoundScore 現在のラウンドスコアを取得
func (p *NapoleonPlayer) GetRoundScore() int { return p.roundScore }

// SetRoundScore 現在のラウンドスコアを設定
func (p *NapoleonPlayer) SetRoundScore(score int) { p.roundScore = score }

// GetCumulativeScore 累積スコアを取得
func (p *NapoleonPlayer) GetCumulativeScore() int { return p.cumulativeScore }

// SetCumulativeScore 累積スコアを設定
func (p *NapoleonPlayer) SetCumulativeScore(score int) { p.cumulativeScore = score }

// CommitRoundScore ラウンドスコアを累積スコアに加算
func (p *NapoleonPlayer) CommitRoundScore() {
	p.cumulativeScore += p.roundScore
}

// ResetRound ラウンドをリセット（ビッド・トリック・手札・チーム状態を初期化）
func (p *NapoleonPlayer) ResetRound() {
	p.bid = -1
	p.isNapoleon = false
	p.isAdjutant = false
	p.adjutantRevealed = false
	p.pictureCards = 0
	p.roundScore = 0
	p.tricksTaken = nil
	p.Reset()
	p.SetIsFinished(false)
}
