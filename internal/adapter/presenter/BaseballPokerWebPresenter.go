//go:build !js || !wasm || casino

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BaseballPokerWebPresenter ベースボールポーカーWebプレゼンタークラス
type BaseballPokerWebPresenter struct{}

// Output ゲーム状態を出力
//
// **配列は必ず配列で返します。**
func (cp *BaseballPokerWebPresenter) Output(c interfaces.BaseballPokerGame, lastErr error) string {
	resObj := new(controller.BaseballPokerWebOutput)

	resObj.Phase = int(c.GetPhase())
	resObj.Seats = baseballSeatsToOutput(c)
	resObj.Street = c.GetStreet()
	resObj.StreetTotal = domain.BaseballUpCards
	// **ワイルドの値はサーバが載せる。** 画面に 3 と 9 を書き写させると、
	// 役の判定と画面の印が別々に育って食い違う。
	resObj.WildValues = []int{domain.BaseballWildThree, domain.BaseballWildNine}
	resObj.BonusValue = domain.BaseballBonusFour
	resObj.BuyInValue = domain.BaseballWildThree
	resObj.Pot = c.GetPot()
	resObj.CurrentBet = c.GetCurrentBet()
	resObj.ToCall = c.GetToCall()
	resObj.RaiseCount = c.GetRaiseCount()
	resObj.CanRaise = c.CanRaise()
	resObj.TurnSeat = c.GetTurnSeat()
	resObj.HumanSeat = c.HumanSeat()
	resObj.IsHumanTurn = c.IsHumanTurn()
	resObj.BuyerSeat = c.GetBuyerSeat()
	resObj.BuyCost = c.GetBuyCost()
	resObj.IsBuying = c.IsHumanBuying()
	resObj.HandNumber = c.GetHandNumber()
	resObj.RemainingCards = c.GetRemainingCards()
	resObj.WinnerSeat = c.WinnerSeat()
	resObj.GameEndFlag = c.GetGameEndFlag()
	cfg := c.GetConfig()
	resObj.Config = &controller.BaseballPokerWebOutCfg{
		Seats: cfg.Seats, InitialChips: cfg.InitialChips, Ante: cfg.Ante,
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if c.GetGameEndFlag() {
		resObj.MessageCode = "baseballpoker.result.gameEnd"
	}

	return marshalOrError(resObj)
}

// baseballSeatsToOutput は席ごとの状態を組み立てる。
//
// **表札は全員に載せ、伏せ札だけを落とす。** 表札はスタッドの読み合いの
// 材料そのもので、隠すとゲームが運になる。伏せ札は位置を保ったまま null に
// する ── 詰めると表札の並びが崩れ、どの札がいつ公開されたのかが画面から
// 復元できない。
func baseballSeatsToOutput(c interfaces.BaseballPokerGame) []*controller.BaseballPokerWebOutputSeat {
	players := c.GetPlayers()
	results := c.GetResults()
	showdown := c.GetPhase() == domain.BaseballPhaseShowdown || c.GetGameEndFlag()

	out := make([]*controller.BaseballPokerWebOutputSeat, 0, len(players))
	for i, p := range players {
		if p == nil {
			continue
		}
		cards, faceUp := p.GetCards(), p.GetFaceUp()
		showAll := p.GetIsHuman() || showdown

		wire := make([]*controller.WebOutputCard, len(cards))
		flags := make([]bool, len(cards))
		for k, card := range cards {
			up := k < len(faceUp) && faceUp[k]
			flags[k] = up
			if showAll || up {
				wire[k] = cardToOutput(card)
			}
		}

		seat := &controller.BaseballPokerWebOutputSeat{
			Name:       p.GetName(),
			IsHuman:    p.GetIsHuman(),
			Chips:      p.GetChips(),
			Bet:        p.GetCurrentBet(),
			Cards:      wire,
			FaceUp:     flags,
			BonusCards: p.GetBonusCards(),
			BestHand:   make([]*controller.WebOutputCard, 0),
			Folded:     p.GetFolded(),
			AllIn:      p.GetAllIn(),
			IsTurn:     i == c.GetTurnSeat() && c.GetPhase() == domain.BaseballPhaseBetting,
			IsBuying:   i == c.GetBuyerSeat(),
		}
		if showdown {
			seat.HandRank = p.GetHandRank()
			seat.UsedWild = p.GetUsedWild()
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
func (cp *BaseballPokerWebPresenter) ActionLogOutput(c interfaces.BaseballPokerGame) string {
	return actionLogOutputJSON(c)
}

// HintOutput ヒントをJSON出力
func (cp *BaseballPokerWebPresenter) HintOutput(c interfaces.BaseballPokerGame) string {
	h := c.GetHint()
	if h == nil {
		return marshalOrError(map[string]any{"hint": nil})
	}
	return marshalOrError(map[string]any{"action": h.Action, "reason": h.Reason})
}
