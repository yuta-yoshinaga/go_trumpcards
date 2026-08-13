//go:build !js || !wasm || extra3

package domain

import (
	"encoding/json"
	"errors"
)

// エラー値。復元時の検証で使う。
var (
	errSakuraHandTooLarge = errors.New("sakura: a hand holds more cards than the deal allows")
	errSakuraNegativeWins = errors.New("sakura: round wins must not be negative")
	errSakuraTakenTooMany = errors.New("sakura: captured more cards than the deck holds")
)

// SakuraPlayer はさくらの 1 席。
type SakuraPlayer struct {
	Player

	isHuman bool
	name    string
	// taken は獲得した札。**点数はここから毎回数える。**
	taken []*Card
	// score は通算の得点。
	score int
	// roundScore はこのラウンドの得点。
	roundScore int
	// roundWins は勝ったラウンド数。
	roundWins int
}

// NewSakuraPlayer は SakuraPlayer を構築する。
func NewSakuraPlayer(name string, isHuman bool) *SakuraPlayer {
	return &SakuraPlayer{
		Player:  Player{cards: make([]*Card, 0, SakuraHandSize)},
		taken:   make([]*Card, 0, SakuraDeckSize),
		isHuman: isHuman,
		name:    name,
	}
}

// NewSakuraPlayersForTable は席数ぶんのプレイヤーを作る。席 0 が人間。
func NewSakuraPlayersForTable(seats int) []*SakuraPlayer {
	players := make([]*SakuraPlayer, 0, seats)
	for i := range seats {
		name := "CPU" + string(rune('0'+i))
		if i == 0 {
			name = "YOU"
		}
		players = append(players, NewSakuraPlayer(name, i == 0))
	}
	return players
}

// GetName は席の表示名を返す。
func (p *SakuraPlayer) GetName() string { return p.name }

// GetIsHuman は人間の席かを返す。
func (p *SakuraPlayer) GetIsHuman() bool { return p.isHuman }

// GetCards は手札を返す。
func (p *SakuraPlayer) GetCards() []*Card { return p.cards }

// GetTaken は獲得した札を返す。
func (p *SakuraPlayer) GetTaken() []*Card { return p.taken }

// AddTaken は獲得した札を加える。
func (p *SakuraPlayer) AddTaken(cards ...*Card) {
	for _, c := range cards {
		if c != nil {
			p.taken = append(p.taken, c)
		}
	}
}

// GetScore は通算得点を返す。
func (p *SakuraPlayer) GetScore() int { return p.score }

// AddScore は通算得点を足す。
func (p *SakuraPlayer) AddScore(v int) { p.score += v }

// GetRoundScore はこのラウンドの得点を返す。
func (p *SakuraPlayer) GetRoundScore() int { return p.roundScore }

// SetRoundScore はこのラウンドの得点を設定する。
func (p *SakuraPlayer) SetRoundScore(v int) { p.roundScore = v }

// GetRoundWins は勝ったラウンド数を返す。
func (p *SakuraPlayer) GetRoundWins() int { return p.roundWins }

// AddRoundWin は勝ったラウンド数を 1 増やす。
func (p *SakuraPlayer) AddRoundWin() { p.roundWins++ }

// CardPoints は獲得した札の素点を返す (追加役は含まない)。
//
// **点数は札から毎回数える。** 別に持たせて獲得のたびに足すと、足し忘れが
// 1 か所あるだけで画面の点と勝敗が食い違う。
func (p *SakuraPlayer) CardPoints() int {
	total := 0
	for _, c := range p.taken {
		total += SakuraCardPoints(c)
	}
	return total
}

// CountByCategory は獲得札を種類ごとに数える (光/タネ/短冊/カス)。
func (p *SakuraPlayer) CountByCategory() map[KoiKoiCategory]int {
	counts := map[KoiKoiCategory]int{}
	for _, c := range p.taken {
		if c != nil {
			counts[KoiKoiCardCategory(c)]++
		}
	}
	return counts
}

// Bonuses は成立した追加役を返す。
func (p *SakuraPlayer) Bonuses() []SakuraBonus {
	out := make([]SakuraBonus, 0, 2)
	if p.CountByCategory()[KoiKoiBright] >= SakuraAllBrightsCount {
		out = append(out, SakuraBonusAllBrights)
	}
	if p.hasCard(SakuraCurtainMonth, SakuraCurtainIndex) &&
		p.hasCard(SakuraMoonMonth, SakuraMoonIndex) {
		out = append(out, SakuraBonusSakuraSake)
	}
	return out
}

// BonusPoints は追加役の合計点を返す。
func (p *SakuraPlayer) BonusPoints() int {
	total := 0
	for _, b := range p.Bonuses() {
		total += SakuraBonusPoints(b)
	}
	return total
}

// TotalPoints は素点と追加役を合わせた点数を返す。
func (p *SakuraPlayer) TotalPoints() int { return p.CardPoints() + p.BonusPoints() }

// hasCard は指定の (月, インデックス) の札を持っているかを返す。
func (p *SakuraPlayer) hasCard(month, index int) bool {
	for _, c := range p.taken {
		if c != nil && c.GetDesign() == month && c.GetValue() == index {
			return true
		}
	}
	return false
}

// ResetForRound は次のラウンドに向けて席の状態を戻す。
func (p *SakuraPlayer) ResetForRound() {
	p.cards = p.cards[:0]
	p.taken = p.taken[:0]
	p.roundScore = 0
}

// sakuraPlayerJSON は SakuraPlayer の JSON 表現。
type sakuraPlayerJSON struct {
	IsHuman    bool    `json:"h"`
	Name       string  `json:"n"`
	Cards      []*Card `json:"cd"`
	Taken      []*Card `json:"tk"`
	Score      int     `json:"s"`
	RoundScore int     `json:"rs"`
	RoundWins  int     `json:"rw"`
}

// MarshalJSON implements json.Marshaler.
func (p *SakuraPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(sakuraPlayerJSON{
		IsHuman: p.isHuman, Name: p.name,
		Cards: p.cards, Taken: p.taken,
		Score: p.score, RoundScore: p.roundScore, RoundWins: p.roundWins,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **獲得札の枚数は山を超えられない。** 点数は獲得札から数えるので、札を
// 水増しした保存は素点をいくらでも押し上げられる ── 得点そのものを検査しても
// 意味がなく、枚数のほうを縛る必要がある。
func (p *SakuraPlayer) UnmarshalJSON(data []byte) error {
	var j sakuraPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Cards) > SakuraHandSize {
		return errSakuraHandTooLarge
	}
	if len(j.Taken) > SakuraDeckSize {
		return errSakuraTakenTooMany
	}
	if j.RoundWins < 0 {
		return errSakuraNegativeWins
	}
	p.isHuman = j.IsHuman
	p.name = j.Name
	p.cards = j.Cards
	p.taken = j.Taken
	p.score = j.Score
	p.roundScore = j.RoundScore
	p.roundWins = j.RoundWins
	return nil
}
