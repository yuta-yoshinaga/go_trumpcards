//go:build !js || !wasm || extra3

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// NiuNiuWebPresenter 闘牛 Web プレゼンタークラス
type NiuNiuWebPresenter struct{}

// niuNiuHandOutput 1 つの手を出力形に落とす。
//
// **reveal が false の手は中身をワイヤに載せない。** 枚数だけを残し、札・役・
// 組み合わせは落とす。
//
// ただし断っておくと、**この分岐は現状の闘牛では通らない**。`deal` が同じ呼び出しで
// `settle` まで進めるので、手が存在するときは必ず精算済みで、配る前は手そのものが
// nil になる。ここに置いてあるのは、プレゼンターが「配ると精算が不可分である」という
// ドメイン側の性質に寄りかからないようにするため。演出のために deal と settle を
// 分ける日が来ても、伏せ札がワイヤに漏れない側から始まる。
//
// 漏れる側から始めた結果がどうなるかは Pontoon (#4485) が示していて、あれは
// 画面で裏面を描くだけで中身を送り続けていた。
//
// 見せる側では格・表示名・倍率、そして牛を作った 3 枚の位置まで渡す。組み合わせ
// 探索をクライアントに再実装させないため。
func niuNiuHandOutput(n interfaces.NiuNiuGame, h *domain.NiuNiuHand, reveal bool) *controller.NiuNiuWebOutputHand {
	if h == nil {
		return nil
	}
	cards := h.GetCards()
	out := &controller.NiuNiuWebOutputHand{
		Cards:    make([]*controller.WebOutputCard, len(cards)),
		Bet:      h.GetBet(),
		ComboIdx: make([]int, 0, domain.NiuNiuComboSize),
		Payout:   h.GetPayout(),
		Hidden:   !reveal,
	}
	if !reveal {
		return out
	}
	out.Rank = int(h.GetRank())
	out.RankLabel = n.GetRankLabel(h.GetRank())
	out.Multiplier = n.GetMultiplier(h.GetRank())
	out.ComboIdx = append(out.ComboIdx, h.GetComboIdx()...)
	for i, c := range cards {
		out.Cards[i] = cardToOutput(c)
	}
	return out
}

// Output ゲーム状態をJSON出力
func (np *NiuNiuWebPresenter) Output(n interfaces.NiuNiuGame, lastErr error) string {
	resObj := new(controller.NiuNiuWebOutput)

	// 配ると同時に精算まで走るゲームなので、開くのは局が終わったときだけ。
	ended := n.GetGameEndFlag()
	seats := n.GetSeats()
	resObj.Seats = make([]*controller.NiuNiuWebOutputSeat, len(seats))
	for i, s := range seats {
		resObj.Seats[i] = &controller.NiuNiuWebOutputSeat{
			Name:  s.GetName(),
			IsCPU: s.IsCPU(),
			Hand:  niuNiuHandOutput(n, s.GetHand(), ended || !s.IsCPU()),
		}
	}

	resObj.BankerHand = niuNiuHandOutput(n, n.GetBankerHand(), ended)
	resObj.BankerIdx = n.GetBankerIdx()
	resObj.Chips = n.GetChips()
	resObj.MaxMultiplier = n.GetMaxMultiplier()
	resObj.LastResult = n.GetLastResult()
	resObj.Phase = n.GetPhase()

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if ended {
		resObj.Message = n.GetLastResult()
		resObj.MessageCode = "niuniu.roundOver"
		resObj.MessageParams = map[string]string{"result": n.GetLastResult()}
	} else {
		resObj.MessageCode = "niuniu.placeBet"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (np *NiuNiuWebPresenter) ActionLogOutput(n interfaces.NiuNiuGame) string {
	return actionLogOutputJSON(n)
}
