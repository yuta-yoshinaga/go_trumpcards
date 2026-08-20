//go:build !js || !wasm || extra3

package presenter

import (
	"regexp"
	"strings"

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
	out.RankKey = domain.NiuNiuRankKey(h.GetRank())
	out.Multiplier = n.GetMultiplier(h.GetRank())
	out.ComboIdx = append(out.ComboIdx, h.GetComboIdx()...)
	for i, c := range cards {
		out.Cards[i] = cardToOutput(c)
	}
	return out
}

// niuNiuNumberedRankKey は数字つきの格 ("n1".."n9") のキーだけに一致する。
var niuNiuNumberedRankKey = regexp.MustCompile(`^n[1-9]$`)

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
	resObj.BankerRankKey = n.GetBankerRankKey()
	resObj.Phase = n.GetPhase()

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if ended {
		// **格ごとに別のコードを送る。** 以前は "親: 牛牛" という完成済みの日本語を
		// params に載せていたので、英語ロケールでもそのまま出ていた (#5567)。
		// 数字つきの格だけ n を渡し、文言はクライアント側の i18n が組み立てる。
		//
		// 認識できないキーでは**何も送らない**。default で roundOverN に流すと
		// n が空のまま "親: 牛" が出る。deal と settle が不可分な今は親の役が
		// 必ず確定しているので到達しないが、Pontoon 風の演出で両者を分ける日に
		// 壊れるのはこの分岐で、しかも画面に出るまで気づけない。CUI と
		// フロントの niuniuRankText も同じく「不明なら描かない」で揃えている。
		key := n.GetBankerRankKey()
		switch {
		case key == "none":
			resObj.MessageCode = "niuniu.roundOverNone"
		case key == "niuniu":
			resObj.MessageCode = "niuniu.roundOverNiuNiu"
		case niuNiuNumberedRankKey.MatchString(key):
			resObj.MessageCode = "niuniu.roundOverN"
			resObj.MessageParams = map[string]string{"n": strings.TrimPrefix(key, "n")}
		}
	} else {
		resObj.MessageCode = "niuniu.placeBet"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (np *NiuNiuWebPresenter) ActionLogOutput(n interfaces.NiuNiuGame) string {
	return actionLogOutputJSON(n)
}
