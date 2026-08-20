//go:build !js || !wasm || extra4

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// RedDogCuiPresenter レッドドッグCUIプレゼンタークラス
type RedDogCuiPresenter struct{}

// redDogRankOf maps a card to Red Dog rank space (Ace high = 14).
func redDogRankOf(c *domain.Card) int {
	if c.GetValue() == 1 {
		return 14
	}
	return c.GetValue()
}

// redDogRankLabel renders a Red Dog rank as a short label (K/Q/J or number).
// Winning ranks are always strictly below the higher card (max 14), so the
// highest possible label here is K (13); Ace never appears.
func redDogRankLabel(rank int) string {
	switch rank {
	case 13:
		return "K"
	case 12:
		return "Q"
	case 11:
		return "J"
	default:
		return strconv.Itoa(rank)
	}
}

// redDogWinningRanksStr lists the ranks strictly between the two initial cards —
// the ranks the third card must hit to win.
func redDogWinningRanksStr(initial []*domain.Card) string {
	if len(initial) != 2 {
		return ""
	}
	lo, hi := redDogRankOf(initial[0]), redDogRankOf(initial[1])
	if lo > hi {
		lo, hi = hi, lo
	}
	labels := make([]string, 0, hi-lo)
	for r := lo + 1; r < hi; r++ {
		labels = append(labels, redDogRankLabel(r))
	}
	return strings.Join(labels, ", ")
}

// Output ゲーム状態を出力
func (rp *RedDogCuiPresenter) Output(rd interfaces.RedDogGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	sb.WriteString(i18n.Tf("reddog.chipsLine", "chips", strconv.Itoa(rd.GetChips())) + "\n")
	sb.WriteString(i18n.Tf("reddog.phaseLine", "phase", rp.phaseStr(rd.GetPhase())) + "\n")
	if rd.GetAnte() > 0 {
		sb.WriteString(i18n.Tf("reddog.anteLine", "ante", strconv.Itoa(rd.GetAnte())))
		if rd.GetRaise() > 0 {
			sb.WriteString(i18n.Tf("reddog.raiseInline", "raise", strconv.Itoa(rd.GetRaise())))
		}
		sb.WriteString("\n")
	}
	if rd.GetPhase() == domain.RedDogPhaseSpreadDecision || rd.GetPhase() == domain.RedDogPhaseEnd {
		sb.WriteString(i18n.Tf("reddog.spreadLine", "spread", strconv.Itoa(rd.GetSpread())) + "\n")
	}
	if rd.GetPhase() == domain.RedDogPhaseSpreadDecision {
		sb.WriteString(i18n.Tf("reddog.cuiSpreadGuide",
			"ranks", redDogWinningRanksStr(rd.GetInitialCards())) + "\n")
	}

	// 配当表。Web はベット前に常設しているのに、CUI はベット額を決める材料が
	// スプレッドの広さだけだった (#5539)。賭け終わった後は出さない。
	if rd.GetPhase() == domain.RedDogPhaseBet {
		sb.WriteString(redDogPaytableStr())
	}

	initial := rd.GetInitialCards()
	if len(initial) > 0 {
		sb.WriteString("--- " + color.Bold(i18n.T("reddog.initialHeader")) + " ---\n")
		parts := make([]string, len(initial))
		for i, c := range initial {
			parts[i] = cuiCardStr(c)
		}
		sb.WriteString(strings.Join(parts, ","))
		sb.WriteString("\n")
	}

	if rd.GetThirdCard() != nil {
		sb.WriteString("--- " + color.Bold(i18n.T("reddog.thirdHeader")) + " ---\n")
		sb.WriteString(cuiCardStr(rd.GetThirdCard()))
		sb.WriteString("\n")
	}

	sb.WriteString("----------\n")

	if lastErr != nil {
		sb.WriteString(i18n.MarkErrorLine(color.Red(lastErr.Error())) + "\n")
	}

	if rd.GetGameEndFlag() {
		switch rd.GetResult() {
		case domain.GameResultWin:
			sb.WriteString(color.Green(i18n.T("reddog.playerWins")) + "\n")
		case domain.GameResultLose:
			sb.WriteString(color.Red(i18n.T("reddog.playerLoses")) + "\n")
		default:
			sb.WriteString(color.Yellow(i18n.T("reddog.push")) + "\n")
		}
		sb.WriteString(i18n.Tf("reddog.totalPayoutLine", "payout", strconv.Itoa(rd.GetTotalPayout())) + "\n")
		sb.WriteString("----------\n")
	}

	return sb.String()
}

// redDogRaiseThreshold はレイズを推奨する最小スプレッド。スプレッド（勝てるランク数）が
// これ以上なら3枚目が的中する確率が5割を超えるため、最大レイズが有利になる。
const redDogRaiseThreshold = 7

// HintOutput スプレッドの広さからステイ/レイズの推奨を出力する
func (rp *RedDogCuiPresenter) HintOutput(rd interfaces.RedDogGame) string {
	switch rd.GetPhase() {
	case domain.RedDogPhaseSpreadDecision:
		spread := rd.GetSpread()
		rec := i18n.T("reddog.hintStayRec")
		if spread >= redDogRaiseThreshold {
			rec = i18n.T("reddog.hintRaiseRec")
		}
		return i18n.Tf("reddog.hintSpread",
			"spread", strconv.Itoa(spread),
			"ranks", redDogWinningRanksStr(rd.GetInitialCards()),
			"rec", rec) + "\n"
	case domain.RedDogPhaseBet, domain.RedDogPhaseInitialDealt:
		return i18n.T("reddog.hintBetFirst") + "\n"
	default:
		return i18n.T("reddog.hintGameOver") + "\n"
	}
}

// ActionLogOutput 棋譜をテキスト出力
func (rp *RedDogCuiPresenter) ActionLogOutput(rd interfaces.RedDogGame) string {
	return actionLogOutputText(rd)
}

// redDogPaytableStr renders the payout table shown before the ante is placed.
//
// **倍率はドメインの定数から読む。**文言に書き写すと、配当を1つ直したときに
// 表だけが古いまま残り、プレイヤーは違う倍率を見て賭ける (#5539)。
func redDogPaytableStr() string {
	rows := []struct {
		key  string
		mult int
	}{
		{"reddog.paySpread1", domain.RedDogPaySpread1},
		{"reddog.paySpread2", domain.RedDogPaySpread2},
		{"reddog.paySpread3", domain.RedDogPaySpread3},
		{"reddog.paySpread4Plus", domain.RedDogPaySpread4},
		{"reddog.payPair", domain.RedDogPayPair},
	}
	parts := make([]string, 0, len(rows)+1)
	for _, r := range rows {
		parts = append(parts, i18n.Tf(r.key, "mult", strconv.Itoa(r.mult)))
	}
	parts = append(parts, i18n.T("reddog.payPush"))
	return i18n.T("reddog.paytableHeader") + strings.Join(parts, " / ") + "\n"
}

// phaseStr フェーズ文字列
func (rp *RedDogCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.RedDogPhaseBet:
		return i18n.T("reddog.phaseBet")
	case domain.RedDogPhaseInitialDealt:
		return i18n.T("reddog.phaseInitialDealt")
	case domain.RedDogPhaseSpreadDecision:
		return i18n.T("reddog.phaseSpreadDecision")
	case domain.RedDogPhaseEnd:
		return i18n.T("reddog.phaseEnd")
	default:
		return i18n.T("reddog.phaseUnknown")
	}
}
