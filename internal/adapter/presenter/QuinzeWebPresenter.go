//go:build !js || !wasm || extra2

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// QuinzeWebPresenter カーンズ Web プレゼンタークラス
type QuinzeWebPresenter struct{}

// quinzeHandOutput 1 つの手を出力形に落とす。
//
// **reveal が false の手は中身をワイヤに載せない。** 画面で裏面を描くだけでは
// レスポンスを覗けば読めてしまう（Pontoon の #4485 で踏んだのと同じ穴）。
// 枚数だけは残す — 引いた枚数は卓上で見えている情報だから。
func quinzeHandOutput(s interfaces.QuinzeGame, h *domain.QuinzeHand, reveal bool) *controller.QuinzeWebOutputHand {
	if h == nil {
		return nil
	}
	cards := h.GetCards()
	out := &controller.QuinzeWebOutputHand{
		Cards:  make([]*controller.WebOutputCard, len(cards)),
		Bet:    h.GetBet(),
		Stood:  h.IsStood(),
		Payout: h.GetPayout(),
		Hidden: !reveal,
	}
	if !reveal {
		return out
	}
	out.TotalPoints = s.GetHandPoints(h)
	out.TotalLabel = s.FormatPoints(out.TotalPoints)
	for i, c := range cards {
		out.Cards[i] = cardToOutput(c)
	}
	return out
}

// Output ゲーム状態をJSON出力
func (sp *QuinzeWebPresenter) Output(s interfaces.QuinzeGame, lastErr error) string {
	resObj := new(controller.QuinzeWebOutput)

	ended := s.GetGameEndFlag()
	seats := s.GetSeats()
	resObj.Seats = make([]*controller.QuinzeWebOutputSeat, len(seats))
	for i, seat := range seats {
		resObj.Seats[i] = &controller.QuinzeWebOutputSeat{
			Name:  seat.GetName(),
			IsCPU: seat.IsCPU(),
			Hand:  quinzeHandOutput(s, seat.GetHand(), ended || !seat.IsCPU()),
		}
	}

	resObj.BankerHand = quinzeHandOutput(s, s.GetBankerHand(), ended || s.IsHumanBanker())
	resObj.BankerIdx = s.GetBankerIdx()
	resObj.IsHumanBanker = s.IsHumanBanker()
	resObj.Chips = s.GetChips()
	resObj.ActiveSeat = s.GetActiveSeat()
	resObj.NextBanker = s.GetNextBanker()
	resObj.LastResult = s.GetLastResult()
	resObj.Phase = s.GetPhase()
	resObj.TargetPoints = domain.QuinzeTarget
	resObj.CanHit = s.CanHit()
	resObj.CanStand = s.CanStand()
	resObj.CpuStandPoints = domain.QuinzeCpuStandPoints

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		switch s.GetPhase() {
		case domain.QuinzePhaseBet:
			if s.IsHumanBanker() {
				resObj.MessageCode = "quinze.dealAsBanker"
			} else {
				resObj.MessageCode = "quinze.placeBet"
			}
		case domain.QuinzePhasePlayerTurn:
			resObj.MessageCode = "quinze.yourTurn"
		case domain.QuinzePhaseBankerTurn:
			resObj.MessageCode = "quinze.bankerTurn"
		case domain.QuinzePhaseEnd:
			resObj.Message = s.GetLastResult()
			resObj.MessageCode = "quinze.roundOver"
			resObj.MessageParams = map[string]string{"result": s.GetLastResult()}
			if s.GetNextBanker() >= 0 {
				resObj.MessageCode = "quinze.bankPasses"
			}
		}
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (sp *QuinzeWebPresenter) ActionLogOutput(s interfaces.QuinzeGame) string {
	return actionLogOutputJSON(s)
}
