package presenter

import (
	"encoding/json"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
)

// BlackJackWebPresenter ブラックジャックWebプレゼンタークラス
type BlackJackWebPresenter struct {
}

// NewBlackJackWebPresenter コンストラクタ
func NewBlackJackWebPresenter() *BlackJackWebPresenter {
	return &BlackJackWebPresenter{}
}

// Output ゲーム状態を出力
func (bjp *BlackJackWebPresenter) Output(bj *domain.BlackJack, lastErr error) string {
	resObj := new(controller.BlackJackWebOutput)
	// dealer
	dealer := bj.GetDealer()
	resObj.Dealer = new(controller.BlackJackWebOutputPlayer)
	resObj.Dealer.Cards = make([]*controller.BlackJackWebOutputCard, 0)
	resObj.Dealer.Chips = dealer.GetChips()
	if bj.GetGameEndFlag() {
		resObj.Dealer.Score = dealer.GetScore()
		for i := 0; i < dealer.GetCardsSize(); i++ {
			resObj.Dealer.Cards = append(resObj.Dealer.Cards, bjp.GetCardObj(dealer.GetCard(i)))
		}
	} else if dealer.GetCardsSize() > 0 {
		resObj.Dealer.Cards = append(resObj.Dealer.Cards, bjp.GetCardObj(dealer.GetCard(0)))
	}
	// player — chips only; score/cards are in hands (single source of truth)
	player := bj.GetPlayer()
	resObj.Player = new(controller.BlackJackWebOutputPlayer)
	resObj.Player.Chips = player.GetChips()
	hands := bj.GetPlayerHands()

	// phase info
	resObj.Phase = bj.GetPhase()
	resObj.CurrentHandIdx = bj.GetCurrentHandIdx()
	resObj.InsuranceBet = bj.GetInsuranceBet()
	resObj.InsuranceAvailable = bj.IsInsuranceAvailable()

	// hands
	resObj.Hands = make([]*controller.BlackJackWebOutputHand, len(hands))
	for i, hand := range hands {
		h := new(controller.BlackJackWebOutputHand)
		h.Score = hand.GetScore()
		h.Cards = make([]*controller.BlackJackWebOutputCard, 0)
		for j := 0; j < hand.GetCardsSize(); j++ {
			h.Cards = append(h.Cards, bjp.GetCardObj(hand.GetCard(j)))
		}
		h.Bet = hand.GetBet()
		h.Stood = hand.IsStood()
		h.Doubled = hand.IsDoubled()
		h.Busted = hand.IsBusted()
		h.IsBlackJack = hand.IsBlackJack()
		h.CanSplit = hand.CanSplit()
		resObj.Hands[i] = h
	}

	// エラーメッセージ（ベット失敗等）
	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if bj.GetGameEndFlag() {
		switch bj.GameJudgment() {
		case domain.GameResultDraw:
			resObj.Message = "It is a draw."
		case domain.GameResultWin:
			resObj.Message = "You are the winner."
		case domain.GameResultLose:
			resObj.Message = "It is your loss."
		}
	}
	res, err := json.Marshal(resObj)
	if err != nil {
		return `{"error":"internal server error"}`
	}
	return string(res)
}

// GetCardObj カード情報取得
func (bjp *BlackJackWebPresenter) GetCardObj(card *domain.Card) *controller.BlackJackWebOutputCard {
	res := new(controller.BlackJackWebOutputCard)
	res.Design = cardDesignToString(card.GetDesign())
	res.Value = card.GetValue()
	return res
}
