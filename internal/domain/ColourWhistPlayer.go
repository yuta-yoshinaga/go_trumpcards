//go:build !js || !wasm || classic

package domain

import (
	"encoding/json"
	"errors"
)

// ColourWhistPlayer はカラーホイストのプレイヤー。
//
// 手札 (GamePlayer)、獲得したトリック (TrickHolder)、通算得点を持ちます。
// **組は契約ごとに変わる**ので、チームは席ではなく本体側が持ちます。
type ColourWhistPlayer struct {
	*GamePlayer
	TrickHolder
	score int
}

// NewColourWhistPlayer はコンストラクタ。
func NewColourWhistPlayer(isHuman bool) *ColourWhistPlayer {
	return &ColourWhistPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// GetScore は通算得点を返す。
func (p *ColourWhistPlayer) GetScore() int { return p.score }

// AddScore は得点を加える。**負にもなります。**
func (p *ColourWhistPlayer) AddScore(n int) { p.score += n }

// SetScore は得点を設定する (主にテスト/復元用)。
func (p *ColourWhistPlayer) SetScore(n int) { p.score = n }

// CountAces は手札のエースの枚数を返す。**Troel の判定に使います。**
func (p *ColourWhistPlayer) CountAces() int {
	n := 0
	for i := range p.GetCardsSize() {
		if p.GetCard(i).GetValue() == 1 {
			n++
		}
	}
	return n
}

// ResetRound はラウンドの状態に戻す (得点は保持)。
func (p *ColourWhistPlayer) ResetRound() {
	resetPlayerRound(p)
}

// ResetGame はゲーム開始時の状態に戻す。
func (p *ColourWhistPlayer) ResetGame() {
	p.ResetRound()
	p.score = 0
}

// colourWhistPlayerJSON is the JSON wire format for ColourWhistPlayer.
type colourWhistPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	Score       int          `json:"sc"`
}

// MarshalJSON implements json.Marshaler.
func (p *ColourWhistPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(colourWhistPlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: &p.TrickHolder,
		Score:       p.score,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *ColourWhistPlayer) UnmarshalJSON(data []byte) error {
	var j colourWhistPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer == nil {
		return errors.New("colourwhist player is missing its base player")
	}
	p.GamePlayer = j.GamePlayer
	if j.TrickHolder != nil {
		p.TrickHolder = *j.TrickHolder
	}
	p.score = j.Score
	return nil
}
