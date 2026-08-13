//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"errors"
)

// エラー値。復元時の検証で使う。
var (
	errIronCrossNegativeChips = errors.New("ironcross: chips must not be negative")
	errIronCrossNegativeBet   = errors.New("ironcross: bet must not be negative")
	errIronCrossHandSize      = errors.New("ironcross: a hand must hold four cards")
	errIronCrossLineRange     = errors.New("ironcross: line out of range")
)

// IronCrossPlayer はアイアンクロスの 1 席。
//
// **席ごとに使うコミュニティが違う。** 縦を選んだ席と横を選んだ席では、
// 同じ十字を見ていても手に入る 3 枚が別物になる ── 共有の場から全員が同じ
// 5 枚を使う Holdem との一番の違いがここ。
type IronCrossPlayer struct {
	Player
	ChipHolder
	bettingPlayerBase

	isHuman bool
	name    string
	// line は選んだ列。選ぶ前は None。
	line     IronCrossLine
	handRank int
	bestHand []*Card
}

// NewIronCrossPlayer は IronCrossPlayer を構築する。
func NewIronCrossPlayer(name string, chips int, isHuman bool) *IronCrossPlayer {
	p := &IronCrossPlayer{
		Player:  Player{cards: make([]*Card, 0, IronCrossHoleCards)},
		isHuman: isHuman,
		name:    name,
	}
	p.SetChips(chips)
	return p
}

// NewIronCrossPlayersForTable は席数ぶんのプレイヤーを作る。席 0 が人間。
func NewIronCrossPlayersForTable(seats, chips int) []*IronCrossPlayer {
	players := make([]*IronCrossPlayer, 0, seats)
	for i := range seats {
		name := "CPU" + string(rune('0'+i))
		if i == 0 {
			name = "YOU"
		}
		players = append(players, NewIronCrossPlayer(name, chips, i == 0))
	}
	return players
}

// GetName は席の表示名を返す。
func (p *IronCrossPlayer) GetName() string { return p.name }

// GetIsHuman は人間の席かを返す。
func (p *IronCrossPlayer) GetIsHuman() bool { return p.isHuman }

// GetCards は手札を返す。
func (p *IronCrossPlayer) GetCards() []*Card { return p.cards }

// GetLine は選んだ列を返す。
func (p *IronCrossPlayer) GetLine() IronCrossLine { return p.line }

// SetLine は選んだ列を設定する。
func (p *IronCrossPlayer) SetLine(l IronCrossLine) { p.line = l }

// GetHandRank は役のランクを返す。
func (p *IronCrossPlayer) GetHandRank() int { return p.handRank }

// GetBestHand は選ばれた最良の 5 枚を返す。
func (p *IronCrossPlayer) GetBestHand() []*Card { return p.bestHand }

// EvaluateLine は指定の列を使ったときの最良の役を返す (席の状態は変えない)。
//
// **選んだ列の 3 枚だけが使える。** もう一方の列の札は、見えていても手札には
// ならない ── ここを「十字 5 枚から自由に選ぶ」と実装すると、縦横を選ぶという
// このゲームの唯一の判断が消える。
func (p *IronCrossPlayer) EvaluateLine(cross []*Card, l IronCrossLine) (int, []*Card) {
	idx := IronCrossLineIndexes(l)
	if len(idx) == 0 {
		return PokerHandHighCard, nil
	}
	all := make([]*Card, 0, IronCrossPoolSize)
	all = append(all, p.cards...)
	for _, i := range idx {
		if i < len(cross) && cross[i] != nil {
			all = append(all, cross[i])
		}
	}
	if len(all) < IronCrossHandSize {
		return PokerHandHighCard, nil
	}

	bestRank := -1
	var bestCards []*Card
	for _, combo := range combinations(all, IronCrossHandSize) {
		rank := evalFiveCardHand(combo)
		if rank > bestRank || (rank == bestRank && compareHighCardsSlice(combo, bestCards) > 0) {
			bestRank = rank
			bestCards = make([]*Card, IronCrossHandSize)
			copy(bestCards, combo)
		}
	}
	return bestRank, bestCards
}

// EvaluateBest は縦と横の両方を試し、強いほうを席に記録する。
//
// **選ぶのはプレイヤーだが、CPU と決着では最良を採る。** 人間が明示的に選んだ
// 場合はその列を尊重する (`SetLine` 済みならそちらを使う)。
func (p *IronCrossPlayer) EvaluateBest(cross []*Card) int {
	if p.line != IronCrossLineNone {
		rank, best := p.EvaluateLine(cross, p.line)
		p.handRank, p.bestHand = rank, best
		return rank
	}
	vRank, vBest := p.EvaluateLine(cross, IronCrossLineVertical)
	hRank, hBest := p.EvaluateLine(cross, IronCrossLineHorizontal)
	if hRank > vRank || (hRank == vRank && compareHighCardsSlice(hBest, vBest) > 0) {
		p.line, p.handRank, p.bestHand = IronCrossLineHorizontal, hRank, hBest
	} else {
		p.line, p.handRank, p.bestHand = IronCrossLineVertical, vRank, vBest
	}
	return p.handRank
}

// ResetForHand は次のハンドに向けて席の状態を戻す。
func (p *IronCrossPlayer) ResetForHand() {
	p.cards = p.cards[:0]
	p.SetFolded(false)
	p.SetAllIn(false)
	p.SetCurrentBet(0)
	p.line = IronCrossLineNone
	p.handRank = 0
	p.bestHand = nil
}

// ironCrossPlayerJSON は IronCrossPlayer の JSON 表現。
type ironCrossPlayerJSON struct {
	Chips      int     `json:"c"`
	CurrentBet int     `json:"b"`
	IsHuman    bool    `json:"h"`
	Name       string  `json:"n"`
	Folded     bool    `json:"f"`
	AllIn      bool    `json:"a"`
	Cards      []*Card `json:"cd"`
	Line       int     `json:"ln"`
	HandRank   int     `json:"hr"`
	BestHand   []*Card `json:"bh"`
}

// MarshalJSON implements json.Marshaler.
func (p *IronCrossPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(ironCrossPlayerJSON{
		Chips: p.GetChips(), CurrentBet: p.GetCurrentBet(),
		IsHuman: p.isHuman, Name: p.name,
		Folded: p.GetFolded(), AllIn: p.GetAllIn(),
		Cards: p.cards, Line: int(p.line), HandRank: p.handRank, BestHand: p.bestHand,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **手札は 0 枚か 4 枚のどちらか。** 途中の枚数は「配っている最中」を意味する。
func (p *IronCrossPlayer) UnmarshalJSON(data []byte) error {
	var j ironCrossPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Chips < 0 {
		return errIronCrossNegativeChips
	}
	if j.CurrentBet < 0 {
		return errIronCrossNegativeBet
	}
	if len(j.Cards) != 0 && len(j.Cards) != IronCrossHoleCards {
		return errIronCrossHandSize
	}
	if j.Line < int(IronCrossLineNone) || j.Line > int(IronCrossLineMax) {
		return errIronCrossLineRange
	}
	p.SetChips(j.Chips)
	p.SetCurrentBet(j.CurrentBet)
	p.isHuman = j.IsHuman
	p.name = j.Name
	p.SetFolded(j.Folded)
	p.SetAllIn(j.AllIn)
	p.cards = j.Cards
	p.line = IronCrossLine(j.Line)
	p.handRank = j.HandRank
	p.bestHand = j.BestHand
	return nil
}
