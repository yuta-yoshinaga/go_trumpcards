//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"errors"
)

// エラー値。復元時の検証で使う。
var (
	errCincinnatiNegativeChips = errors.New("cincinnati: chips must not be negative")
	errCincinnatiNegativeBet   = errors.New("cincinnati: bet must not be negative")
	errCincinnatiHandSize      = errors.New("cincinnati: a hand must hold five cards")
)

// CincinnatiPlayer はシンシナティの 1 席。
//
// **手札は 5 枚。** Holdem の 2 枚と違って、手札だけで役が完成しうる ──
// コミュニティを 1 枚も使わない選択が普通にある。
type CincinnatiPlayer struct {
	Player
	ChipHolder
	bettingPlayerBase

	isHuman  bool
	name     string
	handRank int
	bestHand []*Card
}

// NewCincinnatiPlayer は CincinnatiPlayer を構築する。
func NewCincinnatiPlayer(name string, chips int, isHuman bool) *CincinnatiPlayer {
	p := &CincinnatiPlayer{
		Player:  Player{cards: make([]*Card, 0, CincinnatiHoleCards)},
		isHuman: isHuman,
		name:    name,
	}
	p.SetChips(chips)
	return p
}

// NewCincinnatiPlayersForTable は席数ぶんのプレイヤーを作る。席 0 が人間。
func NewCincinnatiPlayersForTable(seats, chips int) []*CincinnatiPlayer {
	players := make([]*CincinnatiPlayer, 0, seats)
	for i := range seats {
		name := "CPU" + string(rune('0'+i))
		if i == 0 {
			name = "YOU"
		}
		players = append(players, NewCincinnatiPlayer(name, chips, i == 0))
	}
	return players
}

// GetName は席の表示名を返す。
func (p *CincinnatiPlayer) GetName() string { return p.name }

// GetIsHuman は人間の席かを返す。
func (p *CincinnatiPlayer) GetIsHuman() bool { return p.isHuman }

// GetCards は手札を返す。
func (p *CincinnatiPlayer) GetCards() []*Card { return p.cards }

// GetHandRank は役のランクを返す。
func (p *CincinnatiPlayer) GetHandRank() int { return p.handRank }

// GetBestHand は選ばれた最良の 5 枚を返す。
func (p *CincinnatiPlayer) GetBestHand() []*Card { return p.bestHand }

// EvaluateBest は手札 5 枚 + コミュニティ 5 枚から最良の 5 枚を選ぶ。
//
// **C(10,5) = 252 通りを総当たりする。** issue は「組合せ探索は新規実装」と
// していたが、`combinations` も `evalFiveCardHand` も既にあり、Holdem が
// ホールカード 2 枚 + コミュニティ 5 枚に対して同じことをしている。**枚数に
// 依存していないので、そのまま 10 枚に使える** ── 書き足したのは呼び出しだけ。
func (p *CincinnatiPlayer) EvaluateBest(community []*Card) int {
	all := make([]*Card, 0, CincinnatiPoolSize)
	all = append(all, p.cards...)
	all = append(all, community...)
	if len(all) < CincinnatiHandSize {
		p.handRank = PokerHandHighCard
		p.bestHand = nil
		return p.handRank
	}

	bestRank := -1
	var bestCards []*Card
	for _, combo := range combinations(all, CincinnatiHandSize) {
		rank := evalFiveCardHand(combo)
		if rank > bestRank || (rank == bestRank && compareHighCardsSlice(combo, bestCards) > 0) {
			bestRank = rank
			bestCards = make([]*Card, CincinnatiHandSize)
			copy(bestCards, combo)
		}
	}
	p.handRank = bestRank
	p.bestHand = bestCards
	return p.handRank
}

// ResetForHand は次のハンドに向けて席の状態を戻す。
func (p *CincinnatiPlayer) ResetForHand() {
	p.cards = p.cards[:0]
	p.SetFolded(false)
	p.SetAllIn(false)
	p.SetCurrentBet(0)
	p.handRank = 0
	p.bestHand = nil
}

// cincinnatiPlayerJSON は CincinnatiPlayer の JSON 表現。
type cincinnatiPlayerJSON struct {
	Chips      int     `json:"c"`
	CurrentBet int     `json:"b"`
	IsHuman    bool    `json:"h"`
	Name       string  `json:"n"`
	Folded     bool    `json:"f"`
	AllIn      bool    `json:"a"`
	Cards      []*Card `json:"cd"`
	HandRank   int     `json:"hr"`
	BestHand   []*Card `json:"bh"`
}

// MarshalJSON implements json.Marshaler.
func (p *CincinnatiPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(cincinnatiPlayerJSON{
		Chips: p.GetChips(), CurrentBet: p.GetCurrentBet(),
		IsHuman: p.isHuman, Name: p.name,
		Folded: p.GetFolded(), AllIn: p.GetAllIn(),
		Cards: p.cards, HandRank: p.handRank, BestHand: p.bestHand,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **手札は 0 枚か 5 枚のどちらか。** 途中の枚数は「配っている最中」を意味する
// ので、保存に現れたら壊れている ── 通すと役の判定が別の枚数で走る。
func (p *CincinnatiPlayer) UnmarshalJSON(data []byte) error {
	var j cincinnatiPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Chips < 0 {
		return errCincinnatiNegativeChips
	}
	if j.CurrentBet < 0 {
		return errCincinnatiNegativeBet
	}
	if len(j.Cards) != 0 && len(j.Cards) != CincinnatiHoleCards {
		return errCincinnatiHandSize
	}
	p.SetChips(j.Chips)
	p.SetCurrentBet(j.CurrentBet)
	p.isHuman = j.IsHuman
	p.name = j.Name
	p.SetFolded(j.Folded)
	p.SetAllIn(j.AllIn)
	p.cards = j.Cards
	p.handRank = j.HandRank
	p.bestHand = j.BestHand
	return nil
}
