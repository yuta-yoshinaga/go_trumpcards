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

	// チップ情報
	res += "chips: player=" + strconv.Itoa(player.GetChips()) + " dealer=" + strconv.Itoa(dealer.GetChips()) + "\n"

	// フェーズ情報
	res += "phase: " + bjp.phaseStr(bj.GetPhase()) + "\n"

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
		if dealer.GetCardsSize() > 0 {
			res += bjp.GetCardStr(dealer.GetCard(0)) + ",\n"
		}
	}
	res += "----------\n"

	// player hands
	hands := bj.GetPlayerHands()
	for i, hand := range hands {
		prefix := "player"
		if len(hands) > 1 {
			prefix = "hand " + strconv.Itoa(i+1)
		}
		if i == bj.GetCurrentHandIdx() && !bj.GetGameEndFlag() {
			prefix += " (*)"
		}
		res += prefix + " score " + strconv.Itoa(hand.GetScore()) + " bet=" + strconv.Itoa(hand.GetBet())
		if hand.IsDoubled() {
			res += " [DD]"
		}
		if hand.IsBusted() {
			res += " [BUST]"
		}
		if hand.IsStood() {
			res += " [STAND]"
		}
		if hand.IsBlackJack() {
			res += " [BJ]"
		}
		res += "\n"
		for j := 0; j < hand.GetCardsSize(); j++ {
			if j != 0 {
				res += ","
			}
			res += bjp.GetCardStr(hand.GetCard(j))
		}
		res += "\n"
	}

	res += "----------\n"

	// インシュランス情報
	if bj.GetInsuranceBet() > 0 {
		res += "insurance bet: " + strconv.Itoa(bj.GetInsuranceBet()) + "\n"
	}
	if bj.IsInsuranceAvailable() && bj.GetPhase() == entities.BJPhaseInsurance {
		res += "Insurance available!\n"
	}

	// エラーメッセージ（ベット失敗等）
	if errMsg := bj.GetLastError(); errMsg != "" {
		res += errMsg + "\n"
	}

	if bj.GetGameEndFlag() {
		if len(hands) > 1 {
			for i, hand := range hands {
				_ = hand
				result := bj.GameJudgmentForHand(i)
				res += "hand " + strconv.Itoa(i+1) + ": "
				switch result {
				case entities.GameResultDraw:
					res += "It is a draw."
				case entities.GameResultWin:
					res += "You are the winner."
				case entities.GameResultLose:
					res += "It is your loss."
				}
				res += "\n"
			}
		} else {
			switch bj.GameJudgment() {
			case entities.GameResultDraw:
				res += "It is a draw.\n"
			case entities.GameResultWin:
				res += "You are the winner.\n"
			case entities.GameResultLose:
				res += "It is your loss.\n"
			}
		}
		res += "\n----------\n"
	}
	return res
}

// phaseStr フェーズ文字列
func (bjp *BlackJackCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case entities.BJPhaseBet:
		return "BET"
	case entities.BJPhaseDeal:
		return "DEAL"
	case entities.BJPhaseInsurance:
		return "INSURANCE"
	case entities.BJPhaseAction:
		return "ACTION"
	case entities.BJPhaseEnd:
		return "END"
	default:
		return "UNKNOWN"
	}
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
