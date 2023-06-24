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
	if bj.GetGameEndFlag() {
		resObj.Dealer.Score = dealer.GetScore()
		for i := 0; i < dealer.GetCardsSize(); i++ {
			resObj.Dealer.Cards = append(resObj.Dealer.Cards, bjp.GetCardObj(dealer.GetCard(i)))
		}
	} else {
		resObj.Dealer.Cards = append(resObj.Dealer.Cards, bjp.GetCardObj(dealer.GetCard(0)))
	}
	// player
	player := bj.GetPlayer()
	resObj.Player = new(controllers.BlackJackWebOutputPlayer)
	resObj.Player.Cards = make([]*controllers.BlackJackWebOutputCard, 0)
	resObj.Player.Score = player.GetScore()
	for i := 0; i < player.GetCardsSize(); i++ {
		resObj.Player.Cards = append(resObj.Player.Cards, bjp.GetCardObj(player.GetCard(i)))
	}
	if bj.GetGameEndFlag() {
		switch bj.GameJudgment() {
		case 0:
			resObj.Message = "It is a draw."
		case 1:
			resObj.Message = "You are the winner."
		default:
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
