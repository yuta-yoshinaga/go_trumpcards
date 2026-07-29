//go:build !js || !wasm || extra2

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SetteEMezzoWebPresenter セッテ・エ・メッツォ Web プレゼンタークラス
type SetteEMezzoWebPresenter struct{}

// setteEMezzoHandOutput 1 つの手を出力形に落とす。
//
// **reveal が false の手は中身をワイヤに載せない。** 画面で裏面を描くだけでは
// レスポンスを覗けば読めてしまう（Pontoon の #4485 で踏んだのと同じ穴）。
// 枚数だけは残す — 引いた枚数は卓上で見えている情報だから。
func setteEMezzoHandOutput(s interfaces.SetteEMezzoGame, h *domain.SetteEMezzoHand, reveal bool) *controller.SetteEMezzoWebOutputHand {
	if h == nil {
		return nil
	}
	cards := h.GetCards()
	out := &controller.SetteEMezzoWebOutputHand{
		Cards:  make([]*controller.WebOutputCard, len(cards)),
		Bet:    h.GetBet(),
		Stood:  h.IsStood(),
		Payout: h.GetPayout(),
		Hidden: !reveal,
	}
	if !reveal {
		return out
	}
	out.TotalHalves = s.GetHandHalves(h)
	out.TotalLabel = s.FormatHalves(out.TotalHalves)
	out.MattaHalves = h.GetMattaHalves()
	out.HasMatta = h.HasMatta()
	for i, c := range cards {
		out.Cards[i] = cardToOutput(c)
	}
	return out
}

// Output ゲーム状態をJSON出力
func (sp *SetteEMezzoWebPresenter) Output(s interfaces.SetteEMezzoGame, lastErr error) string {
	resObj := new(controller.SetteEMezzoWebOutput)

	ended := s.GetGameEndFlag()
	seats := s.GetSeats()
	resObj.Seats = make([]*controller.SetteEMezzoWebOutputSeat, len(seats))
	for i, seat := range seats {
		resObj.Seats[i] = &controller.SetteEMezzoWebOutputSeat{
			Name:  seat.GetName(),
			IsCPU: seat.IsCPU(),
			Hand:  setteEMezzoHandOutput(s, seat.GetHand(), ended || !seat.IsCPU()),
		}
	}

	resObj.BankerHand = setteEMezzoHandOutput(s, s.GetBankerHand(), ended || s.IsHumanBanker())
	resObj.BankerIdx = s.GetBankerIdx()
	resObj.IsHumanBanker = s.IsHumanBanker()
	resObj.Chips = s.GetChips()
	resObj.ActiveSeat = s.GetActiveSeat()
	resObj.NextBanker = s.GetNextBanker()
	resObj.LastResult = s.GetLastResult()
	resObj.Phase = s.GetPhase()
	resObj.TargetHalves = domain.SetteEMezzoTargetHalves
	resObj.CanHit = s.CanHit()
	resObj.CanStand = s.CanStand()
	resObj.CanSetMatta = s.CanSetMatta()

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		switch s.GetPhase() {
		case domain.SetteEMezzoPhaseBet:
			if s.IsHumanBanker() {
				resObj.MessageCode = "settemezzo.dealAsBanker"
			} else {
				resObj.MessageCode = "settemezzo.placeBet"
			}
		case domain.SetteEMezzoPhasePlayerTurn:
			resObj.MessageCode = "settemezzo.yourTurn"
		case domain.SetteEMezzoPhaseBankerTurn:
			resObj.MessageCode = "settemezzo.bankerTurn"
		case domain.SetteEMezzoPhaseEnd:
			resObj.Message = s.GetLastResult()
			resObj.MessageCode = "settemezzo.roundOver"
			resObj.MessageParams = map[string]string{"result": s.GetLastResult()}
			if s.GetNextBanker() >= 0 {
				resObj.MessageCode = "settemezzo.bankPasses"
			}
		}
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (sp *SetteEMezzoWebPresenter) ActionLogOutput(s interfaces.SetteEMezzoGame) string {
	return actionLogOutputJSON(s)
}
