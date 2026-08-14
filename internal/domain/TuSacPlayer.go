//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
)

// エラー値。復元時の検証で使う。
var (
	errTuSacHandTooLarge  = errors.New("tusac: a hand holds more cards than the deal allows")
	errTuSacBadMeld       = errors.New("tusac: a laid meld is not a valid combination")
	errTuSacNegativeScore = errors.New("tusac: score must not be negative")
)

// TuSacPlayer は四色牌の 1 席。
type TuSacPlayer struct {
	Player

	isHuman bool
	name    string
	// melds は場に出した組み合わせ。**出したら戻せない。**
	melds []TuSacMeld
	// score は通算の得点。
	score int
	// roundScore はこのラウンドの得点 (メルド - 手残り)。
	roundScore int
}

// NewTuSacPlayer は TuSacPlayer を構築する。
func NewTuSacPlayer(name string, isHuman bool) *TuSacPlayer {
	return &TuSacPlayer{
		Player:  Player{cards: make([]*Card, 0, TuSacHandSize+1)},
		melds:   make([]TuSacMeld, 0),
		isHuman: isHuman,
		name:    name,
	}
}

// NewTuSacPlayersForTable は席数ぶんのプレイヤーを作る。席 0 が人間。
func NewTuSacPlayersForTable(seats int) []*TuSacPlayer {
	players := make([]*TuSacPlayer, 0, seats)
	for i := range seats {
		name := "CPU" + string(rune('0'+i))
		if i == 0 {
			name = "YOU"
		}
		players = append(players, NewTuSacPlayer(name, i == 0))
	}
	return players
}

// GetName は席の表示名を返す。
func (p *TuSacPlayer) GetName() string { return p.name }

// GetIsHuman は人間の席かを返す。
func (p *TuSacPlayer) GetIsHuman() bool { return p.isHuman }

// GetCards は手札を返す。
func (p *TuSacPlayer) GetCards() []*Card { return p.cards }

// GetMelds は場に出した組み合わせを返す。
func (p *TuSacPlayer) GetMelds() []TuSacMeld { return p.melds }

// GetScore は通算得点を返す。
func (p *TuSacPlayer) GetScore() int { return p.score }

// AddScore は通算得点を足す。
func (p *TuSacPlayer) AddScore(v int) { p.score += v }

// GetRoundScore はこのラウンドの得点を返す。
func (p *TuSacPlayer) GetRoundScore() int { return p.roundScore }

// SetRoundScore はこのラウンドの得点を設定する。
func (p *TuSacPlayer) SetRoundScore(v int) { p.roundScore = v }

// AddMeld は組み合わせを場に出す。
func (p *TuSacPlayer) AddMeld(kind TuSacMeldKind, cards []*Card) {
	held := make([]*Card, len(cards))
	copy(held, cards)
	p.melds = append(p.melds, TuSacMeld{Kind: kind, Cards: held})
}

// MeldPoints は出した組み合わせの合計得点を返す。
func (p *TuSacPlayer) MeldPoints() int {
	total := 0
	for _, m := range p.melds {
		total += TuSacMeldPoints(m.Kind)
	}
	return total
}

// RemoveCardsAt は指定の添字の札を手札から取り除く。
//
// **大きい添字から消す。** 小さいほうから消すと、残りの添字が 1 つずつ
// ずれて別の札が落ちる ([[feedback_stale_offsets_after_mutation]] と同じ罠)。
func (p *TuSacPlayer) RemoveCardsAt(indexes []int) {
	sorted := make([]int, len(indexes))
	copy(sorted, indexes)
	for i := range sorted {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j] > sorted[i] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	for _, i := range sorted {
		if i >= 0 && i < len(p.cards) {
			p.cards = append(p.cards[:i], p.cards[i+1:]...)
		}
	}
}

// ResetForRound は次のラウンドに向けて席の状態を戻す。
func (p *TuSacPlayer) ResetForRound() {
	p.cards = p.cards[:0]
	p.melds = p.melds[:0]
	p.roundScore = 0
}

// tuSacMeldJSON は TuSacMeld の JSON 表現。
type tuSacMeldJSON struct {
	Kind  int     `json:"k"`
	Cards []*Card `json:"c"`
}

// tuSacPlayerJSON は TuSacPlayer の JSON 表現。
type tuSacPlayerJSON struct {
	IsHuman    bool            `json:"h"`
	Name       string          `json:"n"`
	Cards      []*Card         `json:"cd"`
	Melds      []tuSacMeldJSON `json:"ml"`
	Score      int             `json:"s"`
	RoundScore int             `json:"rs"`
}

// MarshalJSON implements json.Marshaler.
func (p *TuSacPlayer) MarshalJSON() ([]byte, error) {
	melds := make([]tuSacMeldJSON, 0, len(p.melds))
	for _, m := range p.melds {
		melds = append(melds, tuSacMeldJSON{Kind: int(m.Kind), Cards: m.Cards})
	}
	return json.Marshal(tuSacPlayerJSON{
		IsHuman: p.isHuman, Name: p.name,
		Cards: p.cards, Melds: melds,
		Score: p.score, RoundScore: p.roundScore,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **出した組み合わせを復元時に組み直して検証する。** メルドの種別は保存に
// 入っているが、札のほうを書き換えれば「卒 2 枚で 5 点」のような保存が作れる
// ── 種別の範囲検査だけでは通ってしまい、得点だけが静かに増える。
func (p *TuSacPlayer) UnmarshalJSON(data []byte) error {
	var j tuSacPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Cards) > TuSacHandSize+1 {
		return errTuSacHandTooLarge
	}
	if j.Score < 0 {
		return errTuSacNegativeScore
	}

	melds := make([]TuSacMeld, 0, len(j.Melds))
	for _, m := range j.Melds {
		kind := TuSacMeldKind(m.Kind)
		if kind <= TuSacMeldNone || kind > TuSacMeldKindMax {
			return errTuSacBadMeld
		}
		// **札から組み直した種別と一致しなければならない。**
		if TuSacClassifyMeld(m.Cards) != kind {
			return errTuSacBadMeld
		}
		melds = append(melds, TuSacMeld{Kind: kind, Cards: m.Cards})
	}

	p.isHuman = j.IsHuman
	p.name = j.Name
	p.cards = j.Cards
	p.melds = melds
	p.score = j.Score
	p.roundScore = j.RoundScore
	return nil
}
