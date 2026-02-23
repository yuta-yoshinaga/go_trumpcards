package presenters

import (
	"encoding/json"

	"github.com/yuta-yoshinaga/go_trumpcards/entities"
	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/controllers"
)

// PokerWebPresenter ポーカーWebプレゼンタークラス
type PokerWebPresenter struct {
}

// NewPokerWebPresenter コンストラクタ
func NewPokerWebPresenter() *PokerWebPresenter {
	return &PokerWebPresenter{}
}

// Output ゲーム状態を出力
func (pwp *PokerWebPresenter) Output(p *entities.Poker) string {
	resObj := new(controllers.PokerWebOutput)
	resObj.Phase = p.GetPhase()
	resObj.Pot = p.GetPot()
	resObj.Ante = p.GetAnte()

	// player
	player := p.GetPlayer()
	resObj.Player = new(controllers.PokerWebOutputPlayer)
	resObj.Player.Cards = make([]*controllers.PokerWebOutputCard, 0)
	resObj.Player.HandRank = player.GetHandRank()
	resObj.Player.HandName = player.GetHandName()
	resObj.Player.Chips = player.GetChips()
	resObj.Player.Bet = p.GetPlayerBet()
	for i := 0; i < player.GetCardsSize(); i++ {
		resObj.Player.Cards = append(resObj.Player.Cards, pwp.GetCardObj(player.GetCard(i)))
	}

	// dealer
	dealer := p.GetDealer()
	resObj.Dealer = new(controllers.PokerWebOutputPlayer)
	resObj.Dealer.Cards = make([]*controllers.PokerWebOutputCard, 0)
	resObj.Dealer.Chips = dealer.GetChips()
	resObj.Dealer.Bet = p.GetDealerBet()
	if p.GetPhase() == entities.PokerPhaseEnd {
		resObj.Dealer.HandRank = dealer.GetHandRank()
		resObj.Dealer.HandName = dealer.GetHandName()
		for i := 0; i < dealer.GetCardsSize(); i++ {
			resObj.Dealer.Cards = append(resObj.Dealer.Cards, pwp.GetCardObj(dealer.GetCard(i)))
		}
		if p.GetFolded() == 1 {
			resObj.Message = "You folded."
		} else if p.GetFolded() == 2 {
			resObj.Message = "Dealer folded. You win!"
		} else {
			switch p.GameJudgment() {
			case 0:
				resObj.Message = "It is a draw."
			case 1:
				resObj.Message = "You are the winner."
			default:
				resObj.Message = "It is your loss."
			}
		}
	}

	res, _ := json.Marshal(resObj)
	return string(res)
}

// GetCardObj カード情報取得
func (pwp *PokerWebPresenter) GetCardObj(card *entities.Card) *controllers.PokerWebOutputCard {
	res := new(controllers.PokerWebOutputCard)
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
