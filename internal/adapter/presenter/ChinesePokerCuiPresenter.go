//go:build !js || !wasm || casino

package presenter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// ChinesePokerCuiPresenter チャイニーズポーカーCUIプレゼンタークラス
type ChinesePokerCuiPresenter struct {
}

// Output ゲーム状態を出力
func (pp *ChinesePokerCuiPresenter) Output(cp interfaces.ChinesePokerGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	sb.WriteString(i18n.Tf("chinesepoker.chipsLine", "chips", strconv.Itoa(cp.GetChips())) + "\n")
	sb.WriteString(i18n.Tf("chinesepoker.phaseLine", "phase", pp.phaseStr(cp.GetPhase())) + "\n")

	playerCards := cp.GetPlayerCards()
	if len(playerCards) > 0 {
		sb.WriteString("--- " + color.Bold(i18n.T("chinesepoker.playerHeader")) + " ---\n")
		sb.WriteString(i18n.T("chinesepoker.cardsLabel"))
		parts := make([]string, len(playerCards))
		for idx, c := range playerCards {
			parts[idx] = fmt.Sprintf("[%d]%s", idx, cuiCardStr(c))
		}
		sb.WriteString(strings.Join(parts, " "))
		sb.WriteString("\n")
	}

	if cp.GetPhase() == domain.ChinesePokerPhaseEnd {
		front := cp.GetPlayerFront()
		if len(front) > 0 {
			sb.WriteString(i18n.Tf("chinesepoker.frontLine",
				"cards", cuiCardSliceStr(front),
				"rank", pp.frontRankStr(cp.GetPlayerFrontRank()),
			) + "\n")
		}
		middle := cp.GetPlayerMiddle()
		if len(middle) > 0 {
			sb.WriteString(i18n.Tf("chinesepoker.middleLine",
				"cards", cuiCardSliceStr(middle),
				"rank", pp.fiveCardRankStr(cp.GetPlayerMiddleRank()),
			) + "\n")
		}
		back := cp.GetPlayerBack()
		if len(back) > 0 {
			sb.WriteString(i18n.Tf("chinesepoker.backLine",
				"cards", cuiCardSliceStr(back),
				"rank", pp.fiveCardRankStr(cp.GetPlayerBackRank()),
			) + "\n")
		}

		sb.WriteString("--- " + color.Bold(i18n.T("chinesepoker.dealerHeader")) + " ---\n")
		dealerFront := cp.GetDealerFront()
		if len(dealerFront) > 0 {
			sb.WriteString(i18n.Tf("chinesepoker.frontLine",
				"cards", cuiCardSliceStr(dealerFront),
				"rank", pp.frontRankStr(cp.GetDealerFrontRank()),
			) + "\n")
		}
		dealerMiddle := cp.GetDealerMiddle()
		if len(dealerMiddle) > 0 {
			sb.WriteString(i18n.Tf("chinesepoker.middleLine",
				"cards", cuiCardSliceStr(dealerMiddle),
				"rank", pp.fiveCardRankStr(cp.GetDealerMiddleRank()),
			) + "\n")
		}
		dealerBack := cp.GetDealerBack()
		if len(dealerBack) > 0 {
			sb.WriteString(i18n.Tf("chinesepoker.backLine",
				"cards", cuiCardSliceStr(dealerBack),
				"rank", pp.fiveCardRankStr(cp.GetDealerBackRank()),
			) + "\n")
		}
	}

	sb.WriteString("----------\n")

	if lastErr != nil {
		sb.WriteString(i18n.MarkErrorLine(color.Red(lastErr.Error())) + "\n")
	}

	if cp.GetGameEndFlag() {
		sb.WriteString(i18n.Tf("chinesepoker.betLine", "bet", strconv.Itoa(cp.GetBet())) + "\n")
		switch cp.GetResult() {
		case domain.GameResultWin:
			if cp.GetScoop() {
				sb.WriteString(color.Green(i18n.T("chinesepoker.playerScoop")) + "\n")
			} else {
				sb.WriteString(color.Green(i18n.T("chinesepoker.playerWins")) + "\n")
			}
			sb.WriteString(i18n.Tf("chinesepoker.payoutLine", "payout", strconv.Itoa(cp.GetPayout())) + "\n")
		case domain.GameResultLose:
			if cp.GetScoop() {
				sb.WriteString(color.Red(i18n.T("chinesepoker.dealerScoop")) + "\n")
			} else {
				sb.WriteString(color.Red(i18n.T("chinesepoker.dealerWins")) + "\n")
			}
		default:
		}
		if cp.GetPlayerRoyalty() > 0 || cp.GetDealerRoyalty() > 0 {
			sb.WriteString(i18n.Tf("chinesepoker.royaltyLine",
				"player", strconv.Itoa(cp.GetPlayerRoyalty()),
				"dealer", strconv.Itoa(cp.GetDealerRoyalty()),
			) + "\n")
		}
		sb.WriteString("----------\n")
	}

	return sb.String()
}

// HintOutput は13枚の推奨分割とファウル警告を出力する (#4717)。
//
// **Web にはファウル警告 (cp-foul-warning) も列プレビューもあるのに、CUI は
// 13枚を自力で 3/5/5 に分けるしかなく、ファウルしても無警告だった。**
func (cp *ChinesePokerCuiPresenter) HintOutput(g interfaces.ChinesePokerGame) string {
	arr := g.GetSuggestedArrangement()
	if arr == nil {
		return i18n.T("chinesepoker.hintNone") + "\n"
	}
	cards := g.GetPlayerCards()
	row := func(indices []int) string {
		parts := make([]string, 0, len(indices))
		for _, i := range indices {
			if i < 0 || i >= len(cards) {
				continue
			}
			parts = append(parts, "["+strconv.Itoa(i)+"]"+cuiCardStr(cards[i]))
		}
		return strings.Join(parts, " ")
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("chinesepoker.hintRowFront", "cards", row(arr.Front)) + "\n")
	b.WriteString(i18n.Tf("chinesepoker.hintRowMiddle", "cards", row(arr.Middle)) + "\n")
	b.WriteString(i18n.Tf("chinesepoker.hintRowBack", "cards", row(arr.Back)) + "\n")
	// **ランク順に切るだけでは合法とは限らない。**そのまま置かせない。
	if arr.Foul {
		b.WriteString(color.BoldYellow(i18n.T("chinesepoker.hintFoulRisk")) + "\n")
	} else {
		b.WriteString(color.Yellow(i18n.T("chinesepoker.hintSplitOk")) + "\n")
	}
	return b.String()
}

// ActionLogOutput 棋譜をテキスト出力
func (pp *ChinesePokerCuiPresenter) ActionLogOutput(cp interfaces.ChinesePokerGame) string {
	return actionLogOutputText(cp)
}

// phaseStr フェーズ文字列
func (pp *ChinesePokerCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.ChinesePokerPhaseBet:
		return i18n.T("chinesepoker.phaseBet")
	case domain.ChinesePokerPhaseSetHands:
		return i18n.T("chinesepoker.phaseSetHands")
	case domain.ChinesePokerPhaseEnd:
		return i18n.T("chinesepoker.phaseEnd")
	default:
		return i18n.T("chinesepoker.phaseUnknown")
	}
}

// frontRankStr フロントハンドランク文字列（3枚ポーカー）
func (pp *ChinesePokerCuiPresenter) frontRankStr(rank int) string {
	// 3 枚役の名前はスリーカードポーカーと同じ表を引く。英語の
	// `ThreeCardHandNames` を直接返すと日本語ロケールでも英語で出る。
	if rank >= 0 && rank < len(domain.ThreeCardHandNames) {
		return threeCardHandName(rank)
	}
	return i18n.T("chinesepoker.rankUnknown")
}

// fiveCardRankStr 5枚ハンドランク文字列
func (pp *ChinesePokerCuiPresenter) fiveCardRankStr(rank int) string {
	if rank >= 0 && rank < len(domain.PokerHandNames) {
		return domain.PokerHandNames[rank]
	}
	return i18n.T("chinesepoker.rankUnknown")
}
