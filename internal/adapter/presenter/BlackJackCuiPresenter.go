package presenter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BlackJackCuiPresenter ブラックジャックCUIプレゼンタークラス
type BlackJackCuiPresenter struct {
}

// NewBlackJackCuiPresenter コンストラクタ
func NewBlackJackCuiPresenter() *BlackJackCuiPresenter {
	return &BlackJackCuiPresenter{}
}

// Output ゲーム状態を出力
func (bjp *BlackJackCuiPresenter) Output(bj interfaces.BlackJackGame, lastErr error) string {
	player := bj.GetPlayer()
	dealer := bj.GetDealer()
	var b strings.Builder

	b.WriteString("----------\n")

	// チップ情報
	fmt.Fprintf(&b, "chips: player=%d dealer=%d\n", player.GetChips(), dealer.GetChips())

	// フェーズ情報
	fmt.Fprintf(&b, "phase: %s\n", bjp.phaseStr(bj.GetPhase()))

	// dealer
	b.WriteString("dealer score ")
	if bj.GetGameEndFlag() {
		fmt.Fprintf(&b, "%d\n", dealer.GetScore())
		for i := 0; i < dealer.GetCardsSize(); i++ {
			if i != 0 {
				b.WriteString(",")
			}
			b.WriteString(bjp.GetCardStr(dealer.GetCard(i)))
		}
		b.WriteString("\n")
	} else {
		b.WriteString("\n")
		if dealer.GetCardsSize() > 0 {
			fmt.Fprintf(&b, "%s,\n", bjp.GetCardStr(dealer.GetCard(0)))
		}
	}
	b.WriteString("----------\n")

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
		fmt.Fprintf(&b, "%s score %d bet=%d", prefix, hand.GetScore(), hand.GetBet())
		if hand.IsDoubled() {
			b.WriteString(" [DD]")
		}
		if hand.IsBusted() {
			b.WriteString(" [BUST]")
		}
		if hand.IsStood() {
			b.WriteString(" [STAND]")
		}
		if hand.IsBlackJack() {
			b.WriteString(" [BJ]")
		}
		b.WriteString("\n")
		for j := 0; j < hand.GetCardsSize(); j++ {
			if j != 0 {
				b.WriteString(",")
			}
			b.WriteString(bjp.GetCardStr(hand.GetCard(j)))
		}
		b.WriteString("\n")
	}

	b.WriteString("----------\n")

	// インシュランス情報
	if bj.GetInsuranceBet() > 0 {
		fmt.Fprintf(&b, "insurance bet: %d\n", bj.GetInsuranceBet())
	}
	if bj.IsInsuranceAvailable() && bj.GetPhase() == domain.BJPhaseInsurance {
		b.WriteString("Insurance available!\n")
	}

	// エラーメッセージ（ベット失敗等）
	if lastErr != nil {
		fmt.Fprintf(&b, "%s\n", lastErr.Error())
	}

	if bj.GetGameEndFlag() {
		if len(hands) > 1 {
			for i, hand := range hands {
				_ = hand
				result := bj.GameJudgmentForHand(i)
				fmt.Fprintf(&b, "hand %d: ", i+1)
				switch result {
				case domain.GameResultDraw:
					b.WriteString("It is a draw.")
				case domain.GameResultWin:
					b.WriteString("You are the winner.")
				case domain.GameResultLose:
					b.WriteString("It is your loss.")
				}
				b.WriteString("\n")
			}
		} else {
			switch bj.GameJudgment() {
			case domain.GameResultDraw:
				b.WriteString("It is a draw.\n")
			case domain.GameResultWin:
				b.WriteString("You are the winner.\n")
			case domain.GameResultLose:
				b.WriteString("It is your loss.\n")
			}
		}
		b.WriteString("\n----------\n")
	}
	return b.String()
}

// phaseStr フェーズ文字列
func (bjp *BlackJackCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.BJPhaseBet:
		return "BET"
	case domain.BJPhaseDeal:
		return "DEAL"
	case domain.BJPhaseInsurance:
		return "INSURANCE"
	case domain.BJPhaseAction:
		return "ACTION"
	case domain.BJPhaseEnd:
		return "END"
	default:
		return "UNKNOWN"
	}
}

// GetCardStr カード情報文字列取得
func (bjp *BlackJackCuiPresenter) GetCardStr(card *domain.Card) string {
	res := ""
	switch card.GetDesign() {
	case domain.CardDesignSpade:
		res = "SPADE "
	case domain.CardDesignClover:
		res = "CLOVER "
	case domain.CardDesignHeart:
		res = "HEART "
	case domain.CardDesignDiamond:
		res = "DIAMOND "
	default:
		res = "Unsupported card "
	}
	res += strconv.Itoa(card.GetValue())
	return res
}
