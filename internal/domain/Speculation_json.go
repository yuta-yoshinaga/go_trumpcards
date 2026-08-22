//go:build !js || !wasm || extra

package domain

import (
	"encoding/json"
	"fmt"
)

// **フィールドが全て非公開なので marshaller は必須。** 書き忘れると
// `json.Marshal` は 2 バイトの `{}` を黙って返し、Worker はリクエストの
// たびに卓を作り直す (Skat が実際にその状態で出荷されている)。

// speculationPlayerJSON は 1 人分の wire format。
type speculationPlayerJSON struct {
	Name   string  `json:"n"`
	Chips  int     `json:"c"`
	Hidden []*Card `json:"h"`
	Best   *Card   `json:"b"`
}

// speculationJSON は Speculation の wire format。
type speculationJSON struct {
	Deck        *TrumpCards             `json:"dk"`
	Players     []speculationPlayerJSON `json:"pl"`
	Config      SpeculationConfig       `json:"cf"`
	Phase       SpeculationPhase        `json:"ph"`
	TrumpSuit   int                     `json:"ts"`
	TrumpCard   *Card                   `json:"tc"`
	Pot         int                     `json:"pt"`
	TurnSeat    int                     `json:"tn"`
	OfferFrom   int                     `json:"of"`
	OfferTo     int                     `json:"ot"`
	OfferAmount int                     `json:"oa"`
	BestSeat    int                     `json:"bs"`
	RoundNo     int                     `json:"rn"`
	WinnerSeat  int                     `json:"ws"`
	GameEndFlag bool                    `json:"ge"`
	ActionLog   []*ActionLogEntry       `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Speculation) MarshalJSON() ([]byte, error) {
	players := make([]speculationPlayerJSON, len(g.players))
	for i, p := range g.players {
		players[i] = speculationPlayerJSON{
			Name:   p.GetName(),
			Chips:  p.GetChips(),
			Hidden: p.GetHidden(),
			Best:   p.GetBest(),
		}
	}
	return json.Marshal(speculationJSON{
		Deck:        g.deck,
		Players:     players,
		Config:      g.config,
		Phase:       g.phase,
		TrumpSuit:   g.trumpSuit,
		TrumpCard:   g.trumpCard,
		Pot:         g.pot,
		TurnSeat:    g.turnSeat,
		OfferFrom:   g.offerFrom,
		OfferTo:     g.offerTo,
		OfferAmount: g.offerAmount,
		BestSeat:    g.bestSeat,
		RoundNo:     g.roundNo,
		WinnerSeat:  g.winnerSeat,
		GameEndFlag: g.gameEndFlag,
		ActionLog:   g.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **座席番号は復元してから範囲を確かめる。** turnSeat や bestSeat が席数の外を
// 指したまま戻ると、次のめくりで即座に範囲外アクセスになる。
func (g *Speculation) UnmarshalJSON(data []byte) error {
	var j speculationJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > speculationMaxSliceLen || len(j.ActionLog) > speculationMaxSliceLen {
		return fmt.Errorf("speculation: input array exceeds maximum allowed size")
	}
	for _, p := range j.Players {
		if len(p.Hidden) > speculationMaxSliceLen {
			return fmt.Errorf("speculation: input array exceeds maximum allowed size")
		}
	}

	g.deck = j.Deck
	if g.deck == nil {
		g.deck = NewTrumpCards(0)
	}
	g.config = j.Config
	g.config.Normalize()

	g.players = make([]*SpeculationPlayer, len(j.Players))
	for i, pj := range j.Players {
		p := NewSpeculationPlayer(pj.Name, pj.Chips)
		p.SetHidden(pj.Hidden)
		p.SetBest(pj.Best)
		g.players[i] = p
	}
	if len(g.players) == 0 {
		// 席が無い卓は進行できない。既定人数で作り直す。
		g.players = make([]*SpeculationPlayer, g.config.Players)
		for i := range g.players {
			name := fmt.Sprintf("CPU%d", i)
			if i == 0 {
				name = "You"
			}
			g.players[i] = NewSpeculationPlayer(name, g.config.InitialChips)
		}
	}

	g.phase = j.Phase
	if g.phase < SpeculationPhaseFlip || g.phase > SpeculationPhaseMax {
		g.phase = SpeculationPhaseFlip
	}
	g.trumpSuit = j.TrumpSuit
	g.trumpCard = j.TrumpCard
	g.pot = j.Pot
	g.turnSeat = speculationClampSeat(j.TurnSeat, len(g.players), 0)
	g.offerFrom = speculationClampSeat(j.OfferFrom, len(g.players), -1)
	g.offerTo = speculationClampSeat(j.OfferTo, len(g.players), -1)
	g.offerAmount = j.OfferAmount
	g.bestSeat = speculationClampSeat(j.BestSeat, len(g.players), -1)
	g.roundNo = j.RoundNo
	g.winnerSeat = speculationClampSeat(j.WinnerSeat, len(g.players), -1)
	g.gameEndFlag = j.GameEndFlag
	g.actionLog = j.ActionLog
	return nil
}

// speculationClampSeat は席番号を [0, n) に収める。範囲外なら fallback を返す。
func speculationClampSeat(seat, n, fallback int) int {
	if seat < 0 || seat >= n {
		return fallback
	}
	return seat
}
