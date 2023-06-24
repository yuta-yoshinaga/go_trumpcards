package presenters

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/entities"
)

// BlackJackCuiPresenter ブラックジャックCUIプレゼンタークラス
type BlackJackCuiPresenter struct {
}

// NewBlackJackCuiPresenter コンストラクタ
func NewBlackJackCuiPresenter() *BlackJackCuiPresenter {
	return &BlackJackCuiPresenter{}
}

// Output ゲーム状態を出力
func (bjp *BlackJackCuiPresenter) Output(bj *entities.BlackJack) string {
	player := bj.GetPlayer()
	dealer := bj.GetDealer()
	res := "----------\n"
	// dealer
	res += "dealer score "
	if bj.GetGameEndFlag() {
		res += strconv.Itoa(dealer.GetScore()) + "\n"
		for i := 0; i < dealer.GetCardsSize(); i++ {
			if i != 0 {
				res += ","
			}
			res += bjp.GetCardStr(dealer.GetCard(i))
		}
		res += "\n"
	} else {
		res += "\n"
		res += bjp.GetCardStr(dealer.GetCard(0)) + ",\n"
	}
	res += "----------\n"
	// player
	res += "player score " + strconv.Itoa(player.GetScore()) + "\n"
	for i := 0; i < player.GetCardsSize(); i++ {
		if i != 0 {
			res += ","
		}
		res += bjp.GetCardStr(player.GetCard(i))
	}
	res += "\n----------\n"
	if bj.GetGameEndFlag() {
		switch bj.GameJudgment() {
		case 0:
			res += "It is a draw.\n"
		case 1:
			res += "You are the winner.\n"
		default:
			res += "It is your loss.\n"
		}
		res += "\n----------\n"
	}
	return res
}

// GetCardStr カード情報文字列取得
func (bjp *BlackJackCuiPresenter) GetCardStr(card *entities.Card) string {
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
