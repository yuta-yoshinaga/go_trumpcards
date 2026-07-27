//go:build !js || !wasm || casino

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BlackJackSwitchWebPresenter ブラックジャック・スイッチWebプレゼンター
type BlackJackSwitchWebPresenter struct{}

// Output ゲーム状態を出力
func (bp *BlackJackSwitchWebPresenter) Output(g interfaces.BlackJackSwitchGame, lastErr error) string {
	resObj := new(controller.BlackJackSwitchWebOutput)

	dealer := g.GetDealer()
	dealerCards := make([]*controller.WebOutputCard, 0, dealer.GetCardsSize())
	revealHole := g.GetGameEndFlag()
	for i := 0; i < dealer.GetCardsSize(); i++ {
		card := dealer.GetCard(i)
		if i == 1 && !revealHole && dealer.GetCardsSize() >= 2 {
			// ホールカードは伏せる
			dealerCards = append(dealerCards, nil)
			continue
		}
		dealerCards = append(dealerCards, cardToOutput(card))
	}
	resObj.DealerCards = dealerCards
	if revealHole {
		resObj.DealerScore = dealer.GetScore()
	} else if dealer.GetCardsSize() > 0 {
		resObj.DealerScore = domain.CalculateBlackJackScore([]*domain.Card{dealer.GetCard(0)})
	}

	hands := g.GetHands()
	results := g.GetHandResults()
	payouts := g.GetHandPayouts()
	outHands := make([]*controller.BlackJackSwitchWebOutputHand, len(hands))
	for i, h := range hands {
		hCards := make([]*controller.WebOutputCard, 0, h.GetCardsSize())
		for j := 0; j < h.GetCardsSize(); j++ {
			hCards = append(hCards, cardToOutput(h.GetCard(j)))
		}
		oh := &controller.BlackJackSwitchWebOutputHand{
			Cards:   hCards,
			Score:   h.GetScore(),
			Bet:     h.GetBet(),
			Stood:   h.IsStood(),
			Doubled: h.IsDoubled(),
			Busted:  h.IsBusted(),
			IsBJ:    h.IsBlackJack(),
		}
		if i < len(results) {
			oh.Result = int(results[i])
		}
		if i < len(payouts) {
			oh.Payout = payouts[i]
		}
		outHands[i] = oh
	}
	resObj.Hands = outHands
	resObj.Phase = g.GetPhase()
	resObj.CurrentHandIdx = g.GetCurrentHandIdx()
	resObj.Chips = g.GetPlayer().GetChips()
	resObj.Switched = g.IsSwitched()
	resObj.DealerPushed22 = g.IsDealerPushed22()
	resObj.OverallResult = int(g.GetOverallResult())
	resObj.TotalPayout = g.GetTotalPayout()

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if g.GetGameEndFlag() {
		resObj.Message, resObj.MessageCode = blackJackSwitchEndMessage(g)
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜JSON出力
func (bp *BlackJackSwitchWebPresenter) ActionLogOutput(g interfaces.BlackJackSwitchGame) string {
	return actionLogOutputJSON(g)
}

// blackJackSwitchEndMessage は終了時の表示メッセージと i18n キーを返す。
// dealerPushed22 はバナーで別途表示されるため、トップレベルメッセージは
// 常にプレイヤーから見た総合結果を反映する（ナチュラル21がディーラー22に勝つ
// ようなケースで「全プッシュ」と誤認されないようにする）。
func blackJackSwitchEndMessage(g interfaces.BlackJackSwitchGame) (string, string) {
	switch g.GetOverallResult() {
	case domain.GameResultWin:
		return "", "blackjackswitch.result.overallWin"
	case domain.GameResultLose:
		return "", "blackjackswitch.result.overallLose"
	}
	if g.IsDealerPushed22() {
		return "", "blackjackswitch.result.dealer22Push"
	}
	return "", "blackjackswitch.result.overallDraw"
}
