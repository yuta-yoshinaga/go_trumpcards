//go:build !js || !wasm || casino

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// IronCrossWebPresenter アイアンクロスWebプレゼンタークラス
type IronCrossWebPresenter struct{}

// Output ゲーム状態を出力
//
// **配列は必ず配列で返します。**
func (cp *IronCrossWebPresenter) Output(c interfaces.IronCrossGame, lastErr error) string {
	resObj := new(controller.IronCrossWebOutput)

	resObj.Phase = int(c.GetPhase())
	resObj.Seats = ironCrossSeatsToOutput(c)
	resObj.Cross = ironCrossCrossToOutput(c.GetCross())
	resObj.RevealedCount = c.GetRevealedCount()
	resObj.CrossTotal = domain.IronCrossCommunityCards
	resObj.VerticalIndexes = domain.IronCrossLineIndexes(domain.IronCrossLineVertical)
	resObj.HorizontalIndexes = domain.IronCrossLineIndexes(domain.IronCrossLineHorizontal)
	resObj.Pot = c.GetPot()
	resObj.CurrentBet = c.GetCurrentBet()
	resObj.ToCall = c.GetToCall()
	resObj.RaiseCount = c.GetRaiseCount()
	resObj.CanRaise = c.CanRaise()
	resObj.TurnSeat = c.GetTurnSeat()
	resObj.HumanSeat = c.HumanSeat()
	resObj.IsHumanTurn = c.IsHumanTurn()
	resObj.IsChoosing = c.IsChoosing()
	resObj.HandNumber = c.GetHandNumber()
	resObj.RemainingCards = c.GetRemainingCards()
	resObj.WinnerSeat = c.WinnerSeat()
	resObj.GameEndFlag = c.GetGameEndFlag()
	cfg := c.GetConfig()
	resObj.Config = &controller.IronCrossWebOutCfg{
		Seats: cfg.Seats, InitialChips: cfg.InitialChips, Ante: cfg.Ante,
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if c.GetGameEndFlag() {
		resObj.MessageCode = "ironcross.result.gameEnd"
	}

	return marshalOrError(resObj)
}

// ironCrossCrossToOutput は十字を添字を保ったまま並べる。
//
// **伏せている位置は null で埋める。** 詰めてしまうと 3 枚目が中央なのか
// 左なのか画面には分からず、縦横の選択が成り立たなくなる。
func ironCrossCrossToOutput(cross []*domain.Card) []*controller.WebOutputCard {
	out := make([]*controller.WebOutputCard, domain.IronCrossCommunityCards)
	for i := range out {
		if i < len(cross) && cross[i] != nil {
			out[i] = cardToOutput(cross[i])
		}
	}
	return out
}

// ironCrossSeatsToOutput は席ごとの状態を組み立てる。
//
// **CPU の手札と選んだ列はショーダウンまで載せない。** 「画面が出さなければ
// よい」ではなく、**サーバが送らない**ことで守る。
func ironCrossSeatsToOutput(c interfaces.IronCrossGame) []*controller.IronCrossWebOutputSeat {
	players := c.GetPlayers()
	results := c.GetResults()
	showdown := c.GetPhase() == domain.IronCrossPhaseShowdown || c.GetGameEndFlag()

	out := make([]*controller.IronCrossWebOutputSeat, 0, len(players))
	for i, p := range players {
		if p == nil {
			continue
		}
		seat := &controller.IronCrossWebOutputSeat{
			Name:     p.GetName(),
			IsHuman:  p.GetIsHuman(),
			Chips:    p.GetChips(),
			Bet:      p.GetCurrentBet(),
			Cards:    make([]*controller.WebOutputCard, 0),
			BestHand: make([]*controller.WebOutputCard, 0),
			Folded:   p.GetFolded(),
			AllIn:    p.GetAllIn(),
			IsTurn:   i == c.GetTurnSeat() && c.GetPhase() == domain.IronCrossPhaseBetting,
		}
		if p.GetIsHuman() || showdown {
			seat.Cards = cardsToOutputOrEmpty(p.GetCards())
			seat.Line = int(p.GetLine())
		}
		if showdown {
			seat.HandRank = p.GetHandRank()
			seat.BestHand = cardsToOutputOrEmpty(p.GetBestHand())
		}
		if i < len(results) {
			seat.WonAmount = results[i].WonAmount
		}
		out = append(out, seat)
	}
	return out
}

// ActionLogOutput 棋譜をJSON出力
func (cp *IronCrossWebPresenter) ActionLogOutput(c interfaces.IronCrossGame) string {
	return actionLogOutputJSON(c)
}

// HintOutput ヒントをJSON出力
func (cp *IronCrossWebPresenter) HintOutput(c interfaces.IronCrossGame) string {
	h := c.GetHint()
	if h == nil {
		return marshalOrError(map[string]any{"hint": nil})
	}
	return marshalOrError(map[string]any{"action": h.Action, "line": int(h.Line), "reason": h.Reason})
}
