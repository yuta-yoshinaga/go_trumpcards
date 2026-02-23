package presenters

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/entities"
)

// PokerCuiPresenter ポーカーCUIプレゼンタークラス
type PokerCuiPresenter struct {
}

// NewPokerCuiPresenter コンストラクタ
func NewPokerCuiPresenter() *PokerCuiPresenter {
	return &PokerCuiPresenter{}
}

// Output ゲーム状態を出力
func (pcp *PokerCuiPresenter) Output(p *entities.Poker) string {
	player := p.GetPlayer()
	dealer := p.GetDealer()
	res := "----------\n"

	// chips/pot info
	res += "Pot: " + strconv.Itoa(p.GetPot()) + " | Player Chips: " + strconv.Itoa(player.GetChips()) + " | Dealer Chips: " + strconv.Itoa(dealer.GetChips()) + "\n"
	if p.GetDealerBet() > 0 {
		res += "Dealer Bet: " + strconv.Itoa(p.GetDealerBet()) + "\n"
	}
	res += "----------\n"

	// player
	res += "player hand"
	if p.GetPhase() == entities.PokerPhaseEnd {
		res += " [" + player.GetHandName() + "]"
	}
	res += "\n"
	for i := 0; i < player.GetCardsSize(); i++ {
		if i != 0 {
			res += ","
		}
		res += "[" + strconv.Itoa(i) + "]" + pcp.GetCardStr(player.GetCard(i))
	}
	res += "\n----------\n"

	// dealer
	res += "dealer hand"
	if p.GetPhase() == entities.PokerPhaseEnd {
		res += " [" + dealer.GetHandName() + "]"
		res += "\n"
		for i := 0; i < dealer.GetCardsSize(); i++ {
			if i != 0 {
				res += ","
			}
			res += pcp.GetCardStr(dealer.GetCard(i))
		}
		res += "\n"
	} else {
		res += "\n"
	}
	res += "----------\n"

	if p.GetPhase() == entities.PokerPhaseEnd {
		if p.GetFolded() == 1 {
			res += "You folded.\n"
		} else if p.GetFolded() == 2 {
			res += "Dealer folded. You win!\n"
		} else {
			switch p.GameJudgment() {
			case 0:
				res += "It is a draw.\n"
			case 1:
				res += "You are the winner.\n"
			default:
				res += "It is your loss.\n"
			}
		}
		res += "----------\n"
	}
	return res
}

// GetCardStr カード情報文字列取得
func (pcp *PokerCuiPresenter) GetCardStr(card *entities.Card) string {
	res := ""
	switch card.GetDesign() {
	case entities.CardDesignSpade:
		res = "SPADE "
	case entities.CardDesignClover:
		res = "CLOVER "
	case entities.CardDesignHeart:
		res = "HEART "
	case entities.CardDesignDiamond:
		res = "DIAMOND "
	default:
		res = "Unsupported card "
	}
	res += strconv.Itoa(card.GetValue())
	return res
}
