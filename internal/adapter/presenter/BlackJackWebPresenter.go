package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BlackJackWebPresenter ブラックジャックWebプレゼンタークラス
type BlackJackWebPresenter struct {
}

// NewBlackJackWebPresenter コンストラクタ
func NewBlackJackWebPresenter() *BlackJackWebPresenter {
	return &BlackJackWebPresenter{}
}

// Output ゲーム状態を出力
func (bjp *BlackJackWebPresenter) Output(bj interfaces.BlackJackGame, lastErr error) string {
	resObj := new(controller.BlackJackWebOutput)
	// dealer
	dealer := bj.GetDealer()
	resObj.Dealer = new(controller.BlackJackWebOutputPlayer)
	resObj.Dealer.Cards = make([]*controller.WebOutputCard, 0)
	resObj.Dealer.Chips = dealer.GetChips()
	if bj.GetGameEndFlag() {
		resObj.Dealer.Score = dealer.GetScore()
		for i := 0; i < dealer.GetCardsSize(); i++ {
			resObj.Dealer.Cards = append(resObj.Dealer.Cards, cardToOutput(dealer.GetCard(i)))
		}
	} else if dealer.GetCardsSize() > 0 {
		resObj.Dealer.Cards = append(resObj.Dealer.Cards, cardToOutput(dealer.GetCard(0)))
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
	resObj.HintEnabled = bj.IsHintEnabled()
	resObj.SuggestedAction = int(bj.GetBasicStrategySuggestion())
	resObj.DeckCount = bj.GetDeckCount()

	// hands
	resObj.Hands = make([]*controller.BlackJackWebOutputHand, len(hands))
	for i, hand := range hands {
		h := new(controller.BlackJackWebOutputHand)
		h.Score = hand.GetScore()
		h.Cards = make([]*controller.WebOutputCard, 0)
		for j := 0; j < hand.GetCardsSize(); j++ {
			h.Cards = append(h.Cards, cardToOutput(hand.GetCard(j)))
		}
		h.Bet = hand.GetBet()
		h.Stood = hand.IsStood()
		h.Doubled = hand.IsDoubled()
		h.Busted = hand.IsBusted()
		h.IsBlackJack = hand.IsBlackJack()
		h.CanSplit = hand.CanSplit()
		h.Surrendered = hand.IsSurrendered()
		h.CanSurrender = hand.CanSurrender()
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
	res, err := jsonMarshal(resObj)
	if err != nil {
		return `{"error":"internal server error"}`
	}
	return string(res)
}
