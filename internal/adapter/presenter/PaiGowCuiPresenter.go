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

// PaiGowCuiPresenter パイガオポーカーCUIプレゼンタークラス
type PaiGowCuiPresenter struct {
}

// Output ゲーム状態を出力
func (pp *PaiGowCuiPresenter) Output(pg interfaces.PaiGowGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	sb.WriteString(i18n.Tf("paigow.chipsLine", "chips", strconv.Itoa(pg.GetChips())) + "\n")
	sb.WriteString(i18n.Tf("paigow.phaseLine", "phase", pp.phaseStr(pg.GetPhase())) + "\n")

	playerCards := pg.GetPlayerCards()
	if len(playerCards) > 0 {
		sb.WriteString("--- " + color.Bold(i18n.T("paigow.playerHeader")) + " ---\n")
		sb.WriteString(i18n.T("paigow.cardsLabel"))
		parts := make([]string, len(playerCards))
		for i, c := range playerCards {
			parts[i] = fmt.Sprintf("[%d]%s", i, cuiCardStr(c))
		}
		sb.WriteString(strings.Join(parts, " "))
		sb.WriteString("\n")

		if pg.GetPhase() == domain.PaiGowPhaseSetHands && len(playerCards) == domain.PaiGowHandSize {
			var foulSplits []string
			for i := range playerCards {
				for j := i + 1; j < len(playerCards); j++ {
					if pg.IsFoulSplit(i, j) {
						foulSplits = append(foulSplits, fmt.Sprintf("[%d,%d]", i, j))
					}
				}
			}
			if len(foulSplits) > 0 {
				sb.WriteString(i18n.Tf("paigow.foulSplits", "splits", strings.Join(foulSplits, " ")) + "\n")
			} else {
				sb.WriteString(i18n.T("paigow.foulSplitsNone") + "\n")
			}
		}
	}

	if pg.GetPhase() == domain.PaiGowPhaseEnd {
		highHand := pg.GetPlayerHighHand()
		if len(highHand) > 0 {
			sb.WriteString(i18n.Tf("paigow.highLine",
				"cards", cuiCardSliceStr(highHand),
				"rank", pp.highHandRankStr(pg.GetPlayerHighRank()),
			) + "\n")
		}
		lowHand := pg.GetPlayerLowHand()
		if len(lowHand) > 0 {
			sb.WriteString(i18n.Tf("paigow.lowLine",
				"cards", cuiCardSliceStr(lowHand),
				"rank", pp.lowHandRankStr(pg.GetPlayerLowRank()),
			) + "\n")
		}

		sb.WriteString("--- " + color.Bold(i18n.T("paigow.dealerHeader")) + " ---\n")
		dealerHigh := pg.GetDealerHighHand()
		if len(dealerHigh) > 0 {
			sb.WriteString(i18n.Tf("paigow.highLine",
				"cards", cuiCardSliceStr(dealerHigh),
				"rank", pp.highHandRankStr(pg.GetDealerHighRank()),
			) + "\n")
		}
		dealerLow := pg.GetDealerLowHand()
		if len(dealerLow) > 0 {
			sb.WriteString(i18n.Tf("paigow.lowLine",
				"cards", cuiCardSliceStr(dealerLow),
				"rank", pp.lowHandRankStr(pg.GetDealerLowRank()),
			) + "\n")
		}
	}

	sb.WriteString("----------\n")

	// 共通ヘルパに寄せる。ここで lastErr.Error() を直に書いていたので、
	// i18n キーを名乗るエラーはキーがそのまま画面に出ていた (#5526)。
	cuiErrorBlock(&sb, lastErr)

	if pg.GetGameEndFlag() {
		sb.WriteString(i18n.Tf("paigow.betLine", "bet", strconv.Itoa(pg.GetBet())) + "\n")
		switch pg.GetResult() {
		case domain.GameResultWin:
			sb.WriteString(color.Green(i18n.T("paigow.playerWins")) + "\n")
			sb.WriteString(i18n.Tf("paigow.payoutWithCommissionLine",
				"payout", strconv.Itoa(pg.GetPayout()),
				"commission", strconv.Itoa(pg.GetCommission()),
			) + "\n")
		case domain.GameResultLose:
			sb.WriteString(color.Red(i18n.T("paigow.dealerWins")) + "\n")
		case domain.GameResultDraw:
			sb.WriteString(color.Yellow(i18n.T("paigow.push")) + "\n")
			sb.WriteString(i18n.Tf("paigow.payoutLine", "payout", strconv.Itoa(pg.GetPayout())) + "\n")
		default:
		}
		sb.WriteString("----------\n")
	}

	return sb.String()
}

// HintOutput はセットハンドフェーズでの推奨分割を出力する。
//
// **どの2枚をローに置くかは、ディーラーのハウスウェイと同じ規則で選ぶ** (#4696)。
// Web は「自動設定」ボタンと反則チェックを常時出しているのに、CUI は7枚から
// 反則にならない分割を完全に手作業・無警告で探すしかなかった。
func (pp *PaiGowCuiPresenter) HintOutput(pg interfaces.PaiGowGame) string {
	hint := pg.GetHint()
	if hint == nil {
		return i18n.T("paigow.hintNone") + "\n"
	}
	cards := pg.GetPlayerCards()
	low := ""
	if hint.LowIdx0 < len(cards) && hint.LowIdx1 < len(cards) {
		low = cuiCardStr(cards[hint.LowIdx0]) + " " + cuiCardStr(cards[hint.LowIdx1])
	}
	return color.Yellow(i18n.Tf("paigow.hintSplit",
		"idx0", strconv.Itoa(hint.LowIdx0),
		"idx1", strconv.Itoa(hint.LowIdx1),
		"cards", low,
		"reason", hintReasonStr(hint.Reason, paiGowHintReasonKeys),
	)) + "\n"
}

// paiGowHintReasonKeys はパイガオ固有のヒント理由キー。
var paiGowHintReasonKeys = map[string]string{
	"house_way_pair": "paigow.hintReasonPair",
	"house_way_high": "paigow.hintReasonHigh",
}

// ActionLogOutput 棋譜をテキスト出力
func (pp *PaiGowCuiPresenter) ActionLogOutput(pg interfaces.PaiGowGame) string {
	return actionLogOutputText(pg)
}

// phaseStr フェーズ文字列
func (pp *PaiGowCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.PaiGowPhaseBet:
		return i18n.T("paigow.phaseBet")
	case domain.PaiGowPhaseSetHands:
		return i18n.T("paigow.phaseSetHands")
	case domain.PaiGowPhaseEnd:
		return i18n.T("paigow.phaseEnd")
	default:
		return i18n.T("paigow.phaseUnknown")
	}
}

// highHandRankStr ハイハンドランク文字列
func (pp *PaiGowCuiPresenter) highHandRankStr(rank int) string {
	if rank >= 0 && rank < len(domain.PokerHandNames) {
		return domain.PokerHandNames[rank]
	}
	return i18n.T("paigow.rankUnknown")
}

// lowHandRankStr ローハンドランク文字列
func (pp *PaiGowCuiPresenter) lowHandRankStr(rank int) string {
	if rank >= 0 && rank < len(domain.PaiGowLowHandNames) {
		return domain.PaiGowLowHandNames[rank]
	}
	return i18n.T("paigow.rankUnknown")
}
