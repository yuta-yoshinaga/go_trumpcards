package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// PokerWebPresenter ポーカーWebプレゼンタークラス
type PokerWebPresenter struct {
}

// NewPokerWebPresenter コンストラクタ
func NewPokerWebPresenter() *PokerWebPresenter {
	return &PokerWebPresenter{}
}

// Output ゲーム状態を出力
func (pwp *PokerWebPresenter) Output(p interfaces.PokerGame, lastErr error) string {
	resObj := new(controller.PokerWebOutput)
	resObj.Phase = p.GetPhase()
	resObj.Pot = p.GetPot()
	resObj.Ante = p.GetAnte()

	// player
	player := p.GetPlayer()
	resObj.Player = new(controller.PokerWebOutputPlayer)
	resObj.Player.Cards = make([]*controller.WebOutputCard, 0)
	resObj.Player.HandRank = player.GetHandRank()
	resObj.Player.HandName = player.GetHandName()
	resObj.Player.Chips = player.GetChips()
	resObj.Player.Bet = p.GetPlayerBet()
	for i := 0; i < player.GetCardsSize(); i++ {
		resObj.Player.Cards = append(resObj.Player.Cards, cardToOutput(player.GetCard(i)))
	}

	// dealer
	dealer := p.GetDealer()
	resObj.Dealer = new(controller.PokerWebOutputPlayer)
	resObj.Dealer.Cards = make([]*controller.WebOutputCard, 0)
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
			resObj.Dealer.Cards = append(resObj.Dealer.Cards, cardToOutput(dealer.GetCard(i)))
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

	res, err := jsonMarshal(resObj)
	if err != nil {
		return `{"error":"internal server error"}`
	}
	return string(res)
}
