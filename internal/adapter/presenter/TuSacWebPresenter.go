//go:build !js || !wasm || solo

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// tuSacFace は四色牌の札を手続き描画の記述子にする。
//
// **専用の PNG アートを持たない札。** ADR-0033 の手続き描画に乗せて、色 (4 色)
// と駒の漢字を札そのものに書かせる ── 共有のカード絵はトランプのスートを
// 描くので、四色牌では何も意味しない。
func tuSacFace(card *domain.Card) *CardFace {
	if card == nil {
		return nil
	}
	glyph := tuSacPieceGlyph(card.GetValue())
	return &CardFace{
		Glyph: glyph,
		Label: glyph,
		Color: tuSacFaceColor(card.GetDesign()),
		Deck:  "tusac",
	}
}

// tuSacPieceGlyph は駒の漢字を返す。
func tuSacPieceGlyph(value int) string {
	switch value {
	case domain.TuSacPieceGeneral:
		return "將"
	case domain.TuSacPieceAdvisor:
		return "士"
	case domain.TuSacPieceElephant:
		return "象"
	case domain.TuSacPieceChariot:
		return "車"
	case domain.TuSacPieceHorse:
		return "馬"
	case domain.TuSacPieceCannon:
		return "砲"
	case domain.TuSacPieceSoldier:
		return "卒"
	default:
		return "?"
	}
}

// tuSacFaceColor は色調トークンを返す。**スートではなく色そのもの。**
func tuSacFaceColor(design int) string {
	switch design {
	case domain.TuSacColorYellow:
		return "gold"
	case domain.TuSacColorRed:
		return "red"
	case domain.TuSacColorGreen:
		return "green"
	default:
		return "black"
	}
}

// tuSacCardsOutput は札の並びを手続き描画つきで変換する。
func tuSacCardsOutput(cards []*domain.Card) []*controller.WebOutputCard {
	out := make([]*controller.WebOutputCard, 0, len(cards))
	for _, c := range cards {
		out = append(out, cardToOutputWithFace(c, tuSacFace))
	}
	return out
}

// TuSacWebPresenter 四色牌Webプレゼンタークラス
type TuSacWebPresenter struct{}

// Output ゲーム状態を出力
//
// **配列は必ず配列で返します。**
func (cp *TuSacWebPresenter) Output(c interfaces.TuSacGame, lastErr error) string {
	resObj := new(controller.TuSacWebOutput)

	resObj.Phase = int(c.GetPhase())
	resObj.Seats = tuSacSeatsToOutput(c)
	if top := c.GetDiscardTop(); top != nil {
		resObj.DiscardTop = cardToOutputWithFace(top, tuSacFace)
	}
	resObj.DiscardCount = c.GetDiscardCount()
	resObj.StockCount = c.GetStockCount()
	resObj.TurnSeat = c.GetTurnSeat()
	resObj.HumanSeat = c.HumanSeat()
	resObj.IsHumanTurn = c.IsHumanTurn()
	resObj.RoundNumber = c.GetRoundNumber()
	resObj.Rounds = c.GetConfig().Rounds
	resObj.WentOutSeat = c.GetWentOutSeat()
	// **配る枚数・札の総数・種別ごとの得点はサーバが載せる。**
	resObj.HandSize = domain.TuSacHandSize
	resObj.DeckSize = domain.TuSacDeckSize
	resObj.MeldPointsByKind = tuSacMeldPointsByKindOut()
	resObj.WinnerSeat = c.WinnerSeat()
	resObj.GameEndFlag = c.GetGameEndFlag()
	cfg := c.GetConfig()
	resObj.Config = &controller.TuSacWebOutCfg{Seats: cfg.Seats, Rounds: cfg.Rounds}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if c.GetGameEndFlag() {
		resObj.MessageCode = "tusac.result.gameEnd"
	}

	return marshalOrError(resObj)
}

// tuSacMeldPointsByKindOut は種別ごとの得点を添字順に並べる。
func tuSacMeldPointsByKindOut() []int {
	out := make([]int, int(domain.TuSacMeldKindMax)+1)
	for k := domain.TuSacMeldNone; k <= domain.TuSacMeldKindMax; k++ {
		out[int(k)] = domain.TuSacMeldPoints(k)
	}
	return out
}

// tuSacSeatsToOutput は席ごとの状態を組み立てる。
//
// **相手の手札はワイヤに乗せない。** 枚数だけを送る ── 四色牌は最後まで
// 相手の手が見えないゲームで、見えるのは場に出た組み合わせだけ。「画面が
// 出さなければよい」ではなく、**サーバが送らない**ことで守る。
func tuSacSeatsToOutput(c interfaces.TuSacGame) []*controller.TuSacWebOutputSeat {
	players := c.GetPlayers()
	results := c.GetResults()

	out := make([]*controller.TuSacWebOutputSeat, 0, len(players))
	for i, p := range players {
		if p == nil {
			continue
		}
		melds := make([]*controller.TuSacWebOutputMeld, 0, len(p.GetMelds()))
		for _, m := range p.GetMelds() {
			melds = append(melds, &controller.TuSacWebOutputMeld{
				Kind:   int(m.Kind),
				Points: domain.TuSacMeldPoints(m.Kind),
				Cards:  tuSacCardsOutput(m.Cards),
			})
		}
		seat := &controller.TuSacWebOutputSeat{
			Name:       p.GetName(),
			IsHuman:    p.GetIsHuman(),
			Cards:      make([]*controller.WebOutputCard, 0),
			HandCount:  len(p.GetCards()),
			Melds:      melds,
			MeldPoints: p.MeldPoints(),
			Score:      p.GetScore(),
			RoundScore: p.GetRoundScore(),
			IsTurn:     i == c.GetTurnSeat(),
			WentOut:    i == c.GetWentOutSeat(),
		}
		if p.GetIsHuman() {
			seat.Cards = tuSacCardsOutput(p.GetCards())
		}
		if i < len(results) {
			seat.RoundScore = results[i].RoundScore
		}
		out = append(out, seat)
	}
	return out
}

// ActionLogOutput 棋譜をJSON出力
func (cp *TuSacWebPresenter) ActionLogOutput(c interfaces.TuSacGame) string {
	return actionLogOutputJSON(c)
}

// HintOutput ヒントをJSON出力
func (cp *TuSacWebPresenter) HintOutput(c interfaces.TuSacGame) string {
	h := c.GetHint()
	if h == nil {
		return marshalOrError(map[string]any{"hint": nil})
	}
	idx := h.Indexes
	if idx == nil {
		idx = []int{}
	}
	return marshalOrError(map[string]any{
		"action": h.Action, "indexes": idx, "reason": h.Reason,
	})
}
