//go:build !js || !wasm || extra2

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// PontoonWebPresenter ポンツーン Web プレゼンタークラス
type PontoonWebPresenter struct{}

// pontoonHandOutput 1 つの手を出力形に落とす。合計と格をサーバー側で計算して
// 渡すのは、A の 1/11 とファイブカード・トリックの判定をクライアントに
// 再実装させないため。
func pontoonHandOutput(p interfaces.PontoonGame, h *domain.PontoonHand) *controller.PontoonWebOutputHand {
	if h == nil {
		return nil
	}
	cards := h.GetCards()
	out := &controller.PontoonWebOutputHand{
		Cards:   make([]*controller.WebOutputCard, len(cards)),
		Bet:     h.GetBet(),
		Total:   p.GetHandTotal(cards),
		Rank:    int(p.GetHandRank(cards)),
		Twisted: h.IsTwisted(),
		Stuck:   h.IsStuck(),
		Payout:  h.GetPayout(),
	}
	for i, c := range cards {
		out.Cards[i] = cardToOutput(c)
	}
	return out
}

// Output ゲーム状態をJSON出力
func (pp *PontoonWebPresenter) Output(p interfaces.PontoonGame, lastErr error) string {
	resObj := new(controller.PontoonWebOutput)

	seats := p.GetSeats()
	resObj.Seats = make([]*controller.PontoonWebOutputSeat, len(seats))
	for i, s := range seats {
		out := &controller.PontoonWebOutputSeat{
			Name:  s.GetName(),
			IsCPU: s.IsCPU(),
			Hands: make([]*controller.PontoonWebOutputHand, 0, len(s.GetHands())),
		}
		for _, h := range s.GetHands() {
			out.Hands = append(out.Hands, pontoonHandOutput(p, h))
		}
		resObj.Seats[i] = out
	}

	resObj.BankerHand = pontoonHandOutput(p, p.GetBankerHand())
	resObj.BankerIdx = p.GetBankerIdx()
	resObj.IsHumanBanker = p.IsHumanBanker()
	resObj.Chips = p.GetChips()
	resObj.ActiveSeat = p.GetActiveSeat()
	resObj.ActiveHand = p.GetActiveHand()
	resObj.NextBanker = p.GetNextBanker()
	resObj.LastResult = p.GetLastResult()
	resObj.Phase = p.GetPhase()
	resObj.CanStick = p.CanStick()
	resObj.CanTwist = p.CanTwist()
	resObj.CanBuy = p.CanBuy()
	resObj.CanSplit = p.CanSplit()

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		switch p.GetPhase() {
		case domain.PontoonPhaseBet:
			if p.IsHumanBanker() {
				resObj.MessageCode = "pontoon.dealAsBanker"
			} else {
				resObj.MessageCode = "pontoon.placeBet"
			}
		case domain.PontoonPhasePlayerTurn:
			resObj.MessageCode = "pontoon.yourTurn"
		case domain.PontoonPhaseBankerTurn:
			resObj.MessageCode = "pontoon.bankerTurn"
		case domain.PontoonPhaseEnd:
			resObj.Message = p.GetLastResult()
			resObj.MessageCode = "pontoon.roundOver"
			resObj.MessageParams = map[string]string{"result": p.GetLastResult()}
			if p.GetNextBanker() >= 0 {
				resObj.MessageCode = "pontoon.bankPasses"
				resObj.MessageParams["seat"] = fmt.Sprintf("%d", p.GetNextBanker())
			}
		}
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (pp *PontoonWebPresenter) ActionLogOutput(p interfaces.PontoonGame) string {
	return actionLogOutputJSON(p)
}
