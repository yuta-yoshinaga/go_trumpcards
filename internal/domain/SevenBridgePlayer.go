//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// SevenBridgePlayer セブンブリッジのプレイヤー
type SevenBridgePlayer struct {
	*GamePlayer
	RoundScoreHolder
	ScoreBreakdown SevenBridgeScoreBreakdown
	melds          [][]*Card // 場に出したメルド
}

// SevenBridgeScoreBreakdown records the penalty points awarded to the round winner.
type SevenBridgeScoreBreakdown struct {
	Ace    int `json:"ace"`
	Number int `json:"number"`
	Face   int `json:"face"`
	Seven  int `json:"seven"`
}

// Total returns the total points represented by this breakdown.
func (b SevenBridgeScoreBreakdown) Total() int { return b.Ace + b.Number + b.Face + b.Seven }

// GetScoreBreakdown returns this player's latest round score breakdown.
func (p *SevenBridgePlayer) GetScoreBreakdown() SevenBridgeScoreBreakdown { return p.ScoreBreakdown }

// NewSevenBridgePlayer コンストラクタ
func NewSevenBridgePlayer(isHuman bool) *SevenBridgePlayer {
	return &SevenBridgePlayer{
		GamePlayer: NewGamePlayer(isHuman),
		melds:      nil,
	}
}

// GetMelds 場に出したメルド一覧を返す
func (p *SevenBridgePlayer) GetMelds() [][]*Card { return p.melds }

// SetMelds メルドを差し替える（テスト用）
func (p *SevenBridgePlayer) SetMelds(m [][]*Card) { p.melds = m }

// GetMeldCount メルド数を取得
func (p *SevenBridgePlayer) GetMeldCount() int { return len(p.melds) }

// GetMeld 指定インデックスのメルドを取得
func (p *SevenBridgePlayer) GetMeld(i int) []*Card {
	if i < 0 || i >= len(p.melds) {
		return nil
	}
	return p.melds[i]
}

// AppendMeld 新しいメルドを追加
func (p *SevenBridgePlayer) AppendMeld(meld []*Card) {
	copyMeld := make([]*Card, len(meld))
	copy(copyMeld, meld)
	p.melds = append(p.melds, copyMeld)
}

// AddCardToMeld 指定インデックスのメルドに 1 枚追加
func (p *SevenBridgePlayer) AddCardToMeld(i int, card *Card) bool {
	if i < 0 || i >= len(p.melds) {
		return false
	}
	p.melds[i] = append(p.melds[i], card)
	return true
}

// ClearMelds メルドを空にする
func (p *SevenBridgePlayer) ClearMelds() { p.melds = nil }

// ResetRound ラウンドをリセット（手札・スコア・メルド・終了状態を初期化）
func (p *SevenBridgePlayer) ResetRound() {
	resetRoundScored(p)
	p.ScoreBreakdown = SevenBridgeScoreBreakdown{}
	p.ClearMelds()
}

// sevenBridgePlayerJSON is the JSON wire format for SevenBridgePlayer.
type sevenBridgePlayerJSON struct {
	GamePlayer       *GamePlayer               `json:"gp"`
	RoundScoreHolder *RoundScoreHolder         `json:"rh"`
	ScoreBreakdown   SevenBridgeScoreBreakdown `json:"sb,omitempty"`
	Melds            [][]*Card                 `json:"md"`
}

// MarshalJSON implements json.Marshaler.
func (p *SevenBridgePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(sevenBridgePlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
		ScoreBreakdown:   p.ScoreBreakdown,
		Melds:            p.melds,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *SevenBridgePlayer) UnmarshalJSON(data []byte) error {
	var j sevenBridgePlayerJSON
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
	p.ScoreBreakdown = j.ScoreBreakdown
	p.melds = j.Melds
	return nil
}
