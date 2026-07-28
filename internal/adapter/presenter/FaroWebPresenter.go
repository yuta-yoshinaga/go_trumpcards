//go:build !js || !wasm || extra2

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// FaroWebPresenter はファロのWebプレゼンター。
type FaroWebPresenter struct{}

// Output ゲーム状態を出力する。
func (fp *FaroWebPresenter) Output(f interfaces.FaroGame, lastErr error) string {
	resObj := new(controller.FaroWebOutput)

	resObj.Phase = f.GetPhase()
	resObj.Chips = f.GetChips()
	resObj.TurnsPlayed = f.GetTurnsPlayed()
	resObj.TurnsTotal = f.GetTurnsTotal()
	resObj.Remaining = f.GetRemainingCount()
	resObj.TotalPayout = f.GetTotalPayout()
	resObj.GameEndFlag = f.GetGameEndFlag()
	resObj.CallWon = f.GetCallWon()

	bets := f.GetBets()
	resObj.Bets = make([]*controller.FaroWebBet, 0, len(bets))
	for _, r := range f.GetBetRanks() {
		b := bets[r]
		resObj.Bets = append(resObj.Bets, &controller.FaroWebBet{Rank: r, Amount: b.Amount, Copper: b.Copper})
	}

	if soda := f.GetSoda(); soda != nil {
		resObj.Soda = cardToOutput(soda)
	}
	if lt := f.GetLastTurn(); lt != nil {
		resObj.LosingCard = cardToOutput(lt.LosingCard)
		resObj.WinningCard = cardToOutput(lt.WinningCard)
		resObj.Split = lt.Split
	}
	resObj.CallCards = cardsToOutputOrEmpty(f.GetCallCards())
	resObj.CallOrder = f.GetCallOrder()
	if resObj.CallOrder == nil {
		resObj.CallOrder = make([]int, 0)
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if f.GetPhase() == domain.FaroPhaseRoundEnd && f.GetCallOrder() != nil {
		if f.GetCallWon() {
			resObj.Message = "Call won!"
			resObj.MessageCode = "faro.result.callWon"
		} else {
			resObj.Message = "Call lost."
			resObj.MessageCode = "faro.result.callLost"
		}
	} else if f.GetGameEndFlag() {
		resObj.Message = "Game over."
		resObj.MessageCode = "faro.result.gameEnd"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力する。
func (fp *FaroWebPresenter) ActionLogOutput(f interfaces.FaroGame) string {
	return actionLogOutputJSON(f)
}
