package domain

import "encoding/json"

// Rummy500Player Rummy 500プレイヤークラス
type Rummy500Player struct {
	*GamePlayer
	RoundScoreHolder
	// laidMelds はプレイヤーが場に出したメルド一覧。
	// レイオフによってカードが追加され、ラウンド終了時のスコア計算に使用する。
	laidMelds [][]*Card
}

// NewRummy500Player コンストラクタ
func NewRummy500Player(isHuman bool) *Rummy500Player {
	return &Rummy500Player{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// ResetRound ラウンドをリセット（手札・スコア・メルド・終了状態を初期化）
func (p *Rummy500Player) ResetRound() {
	p.SetRoundScore(0)
	p.Reset()
	p.laidMelds = nil
	p.SetIsFinished(false)
}

// GetLaidMelds 場に出したメルド一覧を取得する
func (p *Rummy500Player) GetLaidMelds() [][]*Card { return p.laidMelds }

// SetLaidMelds メルド一覧を設定する（テスト用）
func (p *Rummy500Player) SetLaidMelds(melds [][]*Card) { p.laidMelds = melds }

// AddLaidMeld 新しいメルドを場に追加する
func (p *Rummy500Player) AddLaidMeld(meld []*Card) {
	cp := make([]*Card, len(meld))
	copy(cp, meld)
	p.laidMelds = append(p.laidMelds, cp)
}

// AppendToLaidMeld 既存メルドにカードを追加する（レイオフ）
func (p *Rummy500Player) AppendToLaidMeld(meldIdx int, card *Card) bool {
	if meldIdx < 0 || meldIdx >= len(p.laidMelds) {
		return false
	}
	p.laidMelds[meldIdx] = append(p.laidMelds[meldIdx], card)
	return true
}

// rummy500PlayerJSON is the JSON wire format for Rummy500Player.
type rummy500PlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
	LaidMelds        [][]*Card         `json:"lm"`
}

// MarshalJSON implements json.Marshaler.
func (p *Rummy500Player) MarshalJSON() ([]byte, error) {
	return json.Marshal(rummy500PlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
		LaidMelds:        p.laidMelds,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *Rummy500Player) UnmarshalJSON(data []byte) error {
	var j rummy500PlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	if j.RoundScoreHolder != nil {
		p.RoundScoreHolder = *j.RoundScoreHolder
	}
	p.laidMelds = j.LaidMelds
	return nil
}
