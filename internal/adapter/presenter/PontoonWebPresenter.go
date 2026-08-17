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

// pontoonHandOutput 1 つの手を出力形に落とす。
//
// **reveal が false の手は、中身をワイヤに載せない。** 画面で裏面を描くだけでは
// レスポンスを覗けば読めてしまい、「親の手が読めない」というこのゲームの前提が
// 成立しない。カードは nil に、合計と格も伏せる。
//
// ただし**枚数だけは残す**。Twist と Buy で手札の枚数が局中に変わるので、そこは
// 隠すと盤面が描けなくなるうえ、実際のテーブルでも見えている情報になる。
//
// 見せる側では合計と格をサーバーが計算して渡す。A の 1/11 とファイブカード・
// トリックの判定をクライアントに再実装させないため。
func pontoonHandOutput(p interfaces.PontoonGame, h *domain.PontoonHand, reveal bool) *controller.PontoonWebOutputHand {
	if h == nil {
		return nil
	}
	cards := h.GetCards()
	out := &controller.PontoonWebOutputHand{
		Cards:   make([]*controller.WebOutputCard, len(cards)),
		Bet:     h.GetBet(),
		Twisted: h.IsTwisted(),
		Stuck:   h.IsStuck(),
		Payout:  h.GetPayout(),
		Hidden:  !reveal,
	}
	if !reveal {
		// Cards は長さだけの nil スライス。Total/Rank はゼロ値のまま伏せる。
		return out
	}
	out.Total = p.GetHandTotal(cards)
	out.Rank = int(p.GetHandRank(cards))
	for i, c := range cards {
		out.Cards[i] = cardToOutput(c)
	}
	return out
}

// Output ゲーム状態をJSON出力
func (pp *PontoonWebPresenter) Output(p interfaces.PontoonGame, lastErr error) string {
	resObj := new(controller.PontoonWebOutput)

	// 局が終わるまでは自分の手しか見えない。終われば全員ぶん開く。
	ended := p.GetGameEndFlag()
	seats := p.GetSeats()
	resObj.Seats = make([]*controller.PontoonWebOutputSeat, len(seats))
	for i, s := range seats {
		out := &controller.PontoonWebOutputSeat{
			Name:  s.GetName(),
			IsCPU: s.IsCPU(),
			Hands: make([]*controller.PontoonWebOutputHand, 0, len(s.GetHands())),
		}
		reveal := ended || !s.IsCPU()
		for _, h := range s.GetHands() {
			out.Hands = append(out.Hands, pontoonHandOutput(p, h, reveal))
		}
		resObj.Seats[i] = out
	}

	// 親の手は、自分が親なら自分の手なので見えてよい。
	resObj.BankerHand = pontoonHandOutput(p, p.GetBankerHand(), ended || p.IsHumanBanker())
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
	resObj.StickMin = domain.PontoonStickMin
	resObj.CpuStickMin = domain.PontoonCpuStickMin

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
