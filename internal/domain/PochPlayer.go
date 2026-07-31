//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// PochPlayer はポッホのプレイヤークラス。
type PochPlayer struct {
	*GamePlayer
	// chips は通算のチップ残高。**マイナスに落ちても止めない。**
	//
	// Anaconda や Bouillotte のような「所持が足りなければ降ろす」判定は入れて
	// いない。この game の長さは TargetDeals で決まっていて脱落の概念が無く、
	// 原典にもテーブルステークスの規定が無いため、資金切れで座を外す自然な
	// 区切りが存在しない。収支そのものが順位になる。
	chips int
	// betThisRound はこの pochen ラウンドで出した額。
	betThisRound int
	// folded はこの pochen ラウンドで降りたか。
	folded bool
}

// NewPochPlayer コンストラクタ
func NewPochPlayer(isHuman bool) *PochPlayer {
	return &PochPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// GetChips はチップ残高を返す。
func (p *PochPlayer) GetChips() int { return p.chips }

// AddChips はチップを増減する。
func (p *PochPlayer) AddChips(n int) { p.chips += n }

// GetBet はこの pochen ラウンドで出した額を返す。
func (p *PochPlayer) GetBet() int { return p.betThisRound }

// PlaceBet は追加で n チップ出す。残高からも引く。
func (p *PochPlayer) PlaceBet(n int) {
	if n <= 0 {
		return
	}
	p.betThisRound += n
	p.chips -= n
}

// IsFolded は降りているかを返す。
func (p *PochPlayer) IsFolded() bool { return p.folded }

// Fold は降りる。
func (p *PochPlayer) Fold() { p.folded = true }

// ResetBetting は pochen ラウンドの賭け状態だけを初期化する。
func (p *PochPlayer) ResetBetting() {
	p.betThisRound = 0
	p.folded = false
}

// ResetDeal はディール開始時に手札と賭け状態を初期化する。チップは残す。
func (p *PochPlayer) ResetDeal() {
	p.Reset()
	p.SetIsFinished(false)
	p.ResetBetting()
}

// pochPlayerJSON is the JSON wire format for PochPlayer.
type pochPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
	Chips      int         `json:"ch"`
	Bet        int         `json:"bt"`
	Folded     bool        `json:"fd"`
}

// MarshalJSON implements json.Marshaler.
func (p *PochPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(pochPlayerJSON{
		GamePlayer: p.GamePlayer, Chips: p.chips, Bet: p.betThisRound, Folded: p.folded,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *PochPlayer) UnmarshalJSON(data []byte) error {
	var j pochPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	p.chips = j.Chips
	p.betThisRound = j.Bet
	if p.betThisRound < 0 {
		p.betThisRound = 0
	}
	p.folded = j.Folded
	return nil
}
