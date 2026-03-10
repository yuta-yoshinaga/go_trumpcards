package domain

// HeartsPlayer ハーツプレイヤークラス
type HeartsPlayer struct {
	*GamePlayer
	roundScore      int       // 現在のラウンドのスコア
	cumulativeScore int       // 累積スコア
	tricksTaken     [][]*Card // 獲得したトリック
}

// NewHeartsPlayer コンストラクタ
func NewHeartsPlayer(isHuman bool) *HeartsPlayer {
	return &HeartsPlayer{
		GamePlayer:      NewGamePlayer(isHuman),
		roundScore:      0,
		cumulativeScore: 0,
		tricksTaken:     nil,
	}
}

// GetRoundScore 現在のラウンドスコアを取得
func (p *HeartsPlayer) GetRoundScore() int { return p.roundScore }

// SetRoundScore 現在のラウンドスコアを設定
func (p *HeartsPlayer) SetRoundScore(score int) { p.roundScore = score }

// GetCumulativeScore 累積スコアを取得
func (p *HeartsPlayer) GetCumulativeScore() int { return p.cumulativeScore }

// SetCumulativeScore 累積スコアを設定
func (p *HeartsPlayer) SetCumulativeScore(score int) { p.cumulativeScore = score }

// GetTricksTaken 獲得したトリック一覧を取得
func (p *HeartsPlayer) GetTricksTaken() [][]*Card { return p.tricksTaken }

// GetTrickCount 獲得したトリック数を取得
func (p *HeartsPlayer) GetTrickCount() int { return len(p.tricksTaken) }

// AddTrick トリックを追加
func (p *HeartsPlayer) AddTrick(cards []*Card) {
	p.tricksTaken = append(p.tricksTaken, cards)
}

// CommitRoundScore ラウンドスコアを累積スコアに加算
func (p *HeartsPlayer) CommitRoundScore() {
	p.cumulativeScore += p.roundScore
}

// ResetRound ラウンドをリセット（スコア・トリック・手札・終了状態を初期化）
func (p *HeartsPlayer) ResetRound() {
	p.roundScore = 0
	p.tricksTaken = nil
	p.Reset()
	p.SetIsFinished(false)
}
