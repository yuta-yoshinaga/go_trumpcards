package presenter

import (
	"encoding/json"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
)

// PokerWebPresenter ポーカーWebプレゼンタークラス
type PokerWebPresenter struct {
}

// NewPokerWebPresenter コンストラクタ
func NewPokerWebPresenter() *PokerWebPresenter {
	return &PokerWebPresenter{}
}

// Output ゲーム状態を出力
func (pwp *PokerWebPresenter) Output(p *domain.Poker, lastErr error) string {
	resObj := new(controller.PokerWebOutput)
	resObj.Phase = p.GetPhase()
	resObj.Pot = p.GetPot()
	resObj.Ante = p.GetAnte()

	// player
	player := p.GetPlayer()
	resObj.Player = new(controller.PokerWebOutputPlayer)
	resObj.Player.Cards = make([]*controller.PokerWebOutputCard, 0)
	resObj.Player.HandRank = player.GetHandRank()
	resObj.Player.HandName = player.GetHandName()
	resObj.Player.Chips = player.GetChips()
	resObj.Player.Bet = p.GetPlayerBet()
	for i := 0; i < player.GetCardsSize(); i++ {
		resObj.Player.Cards = append(resObj.Player.Cards, pwp.GetCardObj(player.GetCard(i)))
	}

	// dealer
	dealer := p.GetDealer()
	resObj.Dealer = new(controller.PokerWebOutputPlayer)
	resObj.Dealer.Cards = make([]*controller.PokerWebOutputCard, 0)
	resObj.Dealer.Chips = dealer.GetChips()
	resObj.Dealer.Bet = p.GetDealerBet()
	// エラーメッセージ
	if lastErr != nil {
		resObj.Message = lastErr.Error()
	}

	if p.GetPhase() == domain.PokerPhaseEnd {
		resObj.Dealer.HandRank = dealer.GetHandRank()
		resObj.Dealer.HandName = dealer.GetHandName()
		for i := 0; i < dealer.GetCardsSize(); i++ {
			resObj.Dealer.Cards = append(resObj.Dealer.Cards, pwp.GetCardObj(dealer.GetCard(i)))
		}
		if p.GetFolded() == domain.PokerFoldByPlayer {
			resObj.Message = "You folded."
		} else if p.GetFolded() == domain.PokerFoldByDealer {
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

	res, err := json.Marshal(resObj)
	if err != nil {
		return `{"error":"internal server error"}`
	}
	return string(res)
}

// GetCardObj カード情報取得
func (pwp *PokerWebPresenter) GetCardObj(card *domain.Card) *controller.PokerWebOutputCard {
	res := new(controller.PokerWebOutputCard)
	res.Design = cardDesignToString(card.GetDesign())
	res.Value = card.GetValue()
	return res
}
