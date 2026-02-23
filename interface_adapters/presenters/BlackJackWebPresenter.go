package presenters

import (
	"encoding/json"

	"github.com/yuta-yoshinaga/go_trumpcards/entities"
	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/controllers"
)

// BlackJackWebPresenter ブラックジャックWebプレゼンタークラス
type BlackJackWebPresenter struct {
}

// NewBlackJackWebPresenter コンストラクタ
func NewBlackJackWebPresenter() *BlackJackWebPresenter {
	return &BlackJackWebPresenter{}
}

// Output ゲーム状態を出力
func (bjp *BlackJackWebPresenter) Output(bj *entities.BlackJack) string {
	resObj := new(controllers.BlackJackWebOutput)
	// dealer
	dealer := bj.GetDealer()
	resObj.Dealer = new(controllers.BlackJackWebOutputPlayer)
	resObj.Dealer.Cards = make([]*controllers.BlackJackWebOutputCard, 0)
	resObj.Dealer.Chips = dealer.GetChips()
	if bj.GetGameEndFlag() {
		resObj.Dealer.Score = dealer.GetScore()
		for i := 0; i < dealer.GetCardsSize(); i++ {
			resObj.Dealer.Cards = append(resObj.Dealer.Cards, bjp.GetCardObj(dealer.GetCard(i)))
		}
	} else {
		resObj.Dealer.Cards = append(resObj.Dealer.Cards, bjp.GetCardObj(dealer.GetCard(0)))
	}
	// player — derive score/cards from hand 0 (single source of truth)
	player := bj.GetPlayer()
	resObj.Player = new(controllers.BlackJackWebOutputPlayer)
	resObj.Player.Cards = make([]*controllers.BlackJackWebOutputCard, 0)
	resObj.Player.Chips = player.GetChips()
	hands := bj.GetPlayerHands()
	if len(hands) > 0 {
		hand0 := hands[0]
		resObj.Player.Score = hand0.GetScore()
		for j := 0; j < hand0.GetCardsSize(); j++ {
			resObj.Player.Cards = append(resObj.Player.Cards, bjp.GetCardObj(hand0.GetCard(j)))
		}
	}

	// phase info
	resObj.Phase = bj.GetPhase()
	resObj.CurrentHandIdx = bj.GetCurrentHandIdx()
	resObj.InsuranceBet = bj.GetInsuranceBet()
	resObj.InsuranceAvailable = bj.IsInsuranceAvailable()

	// hands
	resObj.Hands = make([]*controllers.BlackJackWebOutputHand, len(hands))
	for i, hand := range hands {
		h := new(controllers.BlackJackWebOutputHand)
		h.Score = hand.GetScore()
		h.Cards = make([]*controllers.BlackJackWebOutputCard, 0)
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

	if bj.GetGameEndFlag() {
		switch bj.GameJudgment() {
		case entities.GameResultDraw:
			resObj.Message = "It is a draw."
		case entities.GameResultWin:
			resObj.Message = "You are the winner."
		case entities.GameResultLose:
			resObj.Message = "It is your loss."
		}
	}
	res, _ := json.Marshal(resObj)
	return string(res)
}

// GetCardObj カード情報取得
func (bjp *BlackJackWebPresenter) GetCardObj(card *entities.Card) *controllers.BlackJackWebOutputCard {
	res := new(controllers.BlackJackWebOutputCard)
	switch card.GetDesign() {
	case entities.CardDesignSpade:
		res.Design = "SPADE"
	case entities.CardDesignClover:
		res.Design = "CLOVER"
	case entities.CardDesignHeart:
		res.Design = "HEART"
	case entities.CardDesignDiamond:
		res.Design = "DIAMOND"
	default:
		res.Design = "Unsupported card"
	}
	res.Value = card.GetValue()
	return res
}
