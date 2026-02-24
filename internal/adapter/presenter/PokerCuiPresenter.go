package presenter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// PokerCuiPresenter ポーカーCUIプレゼンタークラス
type PokerCuiPresenter struct {
}

// NewPokerCuiPresenter コンストラクタ
func NewPokerCuiPresenter() *PokerCuiPresenter {
	return &PokerCuiPresenter{}
}

// Output ゲーム状態を出力
func (pcp *PokerCuiPresenter) Output(p interfaces.PokerGame, lastErr error) string {
	player := p.GetPlayer()
	dealer := p.GetDealer()
	var b strings.Builder

	b.WriteString("----------\n")

	// chips/pot info
	fmt.Fprintf(&b, "Pot: %d | Player Chips: %d | Dealer Chips: %d\n", p.GetPot(), player.GetChips(), dealer.GetChips())
	if p.GetDealerBet() > 0 {
		fmt.Fprintf(&b, "Dealer Bet: %d\n", p.GetDealerBet())
	}
	b.WriteString("----------\n")

	// player
	b.WriteString("player hand")
	if p.GetPhase() == domain.PokerPhaseEnd {
		fmt.Fprintf(&b, " [%s]", player.GetHandName())
	}
	b.WriteString("\n")
	for i := 0; i < player.GetCardsSize(); i++ {
		if i != 0 {
			b.WriteString(",")
		}
		fmt.Fprintf(&b, "[%d]%s", i, pcp.GetCardStr(player.GetCard(i)))
	}
	b.WriteString("\n----------\n")

	// dealer
	b.WriteString("dealer hand")
	if p.GetPhase() == domain.PokerPhaseEnd {
		fmt.Fprintf(&b, " [%s]", dealer.GetHandName())
		b.WriteString("\n")
		for i := 0; i < dealer.GetCardsSize(); i++ {
			if i != 0 {
				b.WriteString(",")
			}
			b.WriteString(pcp.GetCardStr(dealer.GetCard(i)))
		}
		b.WriteString("\n")
	} else {
		b.WriteString("\n")
	}
	b.WriteString("----------\n")

	// エラーメッセージ
	if lastErr != nil {
		fmt.Fprintf(&b, "%s\n", lastErr.Error())
	}

	if p.GetPhase() == domain.PokerPhaseEnd {
		if p.GetFolded() == domain.PokerFoldByPlayer {
			b.WriteString("You folded.\n")
		} else if p.GetFolded() == domain.PokerFoldByDealer {
			b.WriteString("Dealer folded. You win!\n")
		} else {
			switch p.GameJudgment() {
			case 0:
				b.WriteString("It is a draw.\n")
			case 1:
				b.WriteString("You are the winner.\n")
			default:
				b.WriteString("It is your loss.\n")
			}
		}
		b.WriteString("----------\n")
	}
	return b.String()
}

// GetCardStr カード情報文字列取得
func (pcp *PokerCuiPresenter) GetCardStr(card *domain.Card) string {
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
