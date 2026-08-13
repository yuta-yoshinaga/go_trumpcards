//go:build !js || !wasm || casino

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CincinnatiWebPresenter シンシナティWebプレゼンタークラス
type CincinnatiWebPresenter struct{}

// Output ゲーム状態を出力
//
// **配列は必ず配列で返します。**
func (cp *CincinnatiWebPresenter) Output(c interfaces.CincinnatiGame, lastErr error) string {
	resObj := new(controller.CincinnatiWebOutput)

	resObj.Phase = int(c.GetPhase())
	resObj.Seats = cincinnatiSeatsToOutput(c)
	// **伏せているコミュニティは載せない。** 枚数だけ別に伝える。
	resObj.Community = cardsToOutputOrEmpty(c.GetCommunityCards())
	resObj.RevealedCount = c.GetRevealedCount()
	resObj.CommunityTotal = domain.CincinnatiCommunityCards
	resObj.Pot = c.GetPot()
	resObj.CurrentBet = c.GetCurrentBet()
	resObj.ToCall = c.GetToCall()
	resObj.RaiseCount = c.GetRaiseCount()
	resObj.CanRaise = c.CanRaise()
	resObj.TurnSeat = c.GetTurnSeat()
	resObj.HumanSeat = c.HumanSeat()
	resObj.IsHumanTurn = c.IsHumanTurn()
	resObj.HandNumber = c.GetHandNumber()
	resObj.RemainingCards = c.GetRemainingCards()
	resObj.WinnerSeat = c.WinnerSeat()
	resObj.GameEndFlag = c.GetGameEndFlag()
	cfg := c.GetConfig()
	resObj.Config = &controller.CincinnatiWebOutCfg{
		Seats: cfg.Seats, InitialChips: cfg.InitialChips, Ante: cfg.Ante,
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if c.GetGameEndFlag() {
		resObj.MessageCode = "cincinnati.result.gameEnd"
	}

	return marshalOrError(resObj)
}

// cincinnatiSeatsToOutput は席ごとの状態を組み立てる。
//
// **CPU の手札はショーダウンまで載せない。** 手札が 5 枚もあるゲームなので、
// ワイヤに乗せてしまうと相手の役がほぼ確定して勝負にならない ── 「画面が
// 出さなければよい」ではなく、**サーバが送らない**ことで守る。
func cincinnatiSeatsToOutput(c interfaces.CincinnatiGame) []*controller.CincinnatiWebOutputSeat {
	players := c.GetPlayers()
	results := c.GetResults()
	showdown := c.GetPhase() == domain.CincinnatiPhaseShowdown || c.GetGameEndFlag()

	out := make([]*controller.CincinnatiWebOutputSeat, 0, len(players))
	for i, p := range players {
		if p == nil {
			continue
		}
		seat := &controller.CincinnatiWebOutputSeat{
			Name:     p.GetName(),
			IsHuman:  p.GetIsHuman(),
			Chips:    p.GetChips(),
			Bet:      p.GetCurrentBet(),
			Cards:    make([]*controller.WebOutputCard, 0),
			BestHand: make([]*controller.WebOutputCard, 0),
			Folded:   p.GetFolded(),
			AllIn:    p.GetAllIn(),
			IsTurn:   i == c.GetTurnSeat() && c.GetPhase() == domain.CincinnatiPhaseBetting,
		}
		if p.GetIsHuman() || showdown {
			seat.Cards = cardsToOutputOrEmpty(p.GetCards())
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
func (cp *CincinnatiWebPresenter) ActionLogOutput(c interfaces.CincinnatiGame) string {
	return actionLogOutputJSON(c)
}

// HintOutput ヒントをJSON出力
func (cp *CincinnatiWebPresenter) HintOutput(c interfaces.CincinnatiGame) string {
	h := c.GetHint()
	if h == nil {
		return marshalOrError(map[string]any{"hint": nil})
	}
	return marshalOrError(map[string]any{"action": h.Action, "reason": h.Reason})
}
