//go:build !js || !wasm || extra

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// TrenteEtQuaranteCuiPresenter renders the Trente et Quarante CUI view.
type TrenteEtQuaranteCuiPresenter struct{}

// trenteEtQuaranteRowStr は列の札を "♠5 ♥J" 形式で返す。
func trenteEtQuaranteRowStr(cards []*domain.Card) string {
	if len(cards) == 0 {
		return "-"
	}
	parts := make([]string, len(cards))
	for i, c := range cards {
		parts[i] = cuiCardStr(c)
	}
	return strings.Join(parts, " ")
}

// trenteEtQuaranteBetKeys はベット種別の i18n キー。
var trenteEtQuaranteBetKeys = map[domain.TrenteEtQuaranteBet]string{
	domain.TrenteEtQuaranteBetNoir:    "trenteetquarante.betNoir",
	domain.TrenteEtQuaranteBetRouge:   "trenteetquarante.betRouge",
	domain.TrenteEtQuaranteBetCouleur: "trenteetquarante.betCouleur",
	domain.TrenteEtQuaranteBetInverse: "trenteetquarante.betInverse",
}

func trenteEtQuaranteBetName(bet domain.TrenteEtQuaranteBet) string {
	if key, ok := trenteEtQuaranteBetKeys[bet]; ok {
		return i18n.T(key)
	}
	return "-"
}

func trenteEtQuaranteWinningName(row int) string {
	switch row {
	case domain.TrenteEtQuaranteRowNoir:
		return i18n.T("trenteetquarante.winningNoir")
	case domain.TrenteEtQuaranteRowRouge:
		return i18n.T("trenteetquarante.winningRouge")
	default:
		return i18n.T("trenteetquarante.winningNone")
	}
}

func trenteEtQuaranteColorName(red bool) string {
	if red {
		return i18n.T("trenteetquarante.colorRed")
	}
	return i18n.T("trenteetquarante.colorBlack")
}

// Output renders the current game state for the active locale.
func (p *TrenteEtQuaranteCuiPresenter) Output(g interfaces.TrenteEtQuaranteGame, lastErr error) string {
	return buildCuiOutput(i18n.T("trenteetquarante.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("trenteetquarante.chipsLine",
			"chips", strconv.Itoa(g.GetChips()),
			"round", strconv.Itoa(g.GetRoundNumber()),
			"deck", strconv.Itoa(g.GetRemainingDeck())) + "\n")

		if g.GetPhase() == domain.TrenteEtQuarantePhaseResult {
			b.WriteString(i18n.Tf("trenteetquarante.betLine",
				"bet", trenteEtQuaranteBetName(g.GetCurrentBet()),
				"stake", strconv.Itoa(g.GetStake())) + "\n")
			b.WriteString(i18n.Tf("trenteetquarante.noirLine",
				"cards", trenteEtQuaranteRowStr(g.GetNoirRow()),
				"total", strconv.Itoa(g.GetNoirTotal())) + "\n")
			b.WriteString(i18n.Tf("trenteetquarante.rougeLine",
				"cards", trenteEtQuaranteRowStr(g.GetRougeRow()),
				"total", strconv.Itoa(g.GetRougeTotal())) + "\n")
			b.WriteString(i18n.Tf("trenteetquarante.firstCardLine",
				"color", trenteEtQuaranteColorName(g.GetFirstCardRed())) + "\n")
			b.WriteString(i18n.Tf("trenteetquarante.winningLine",
				"row", trenteEtQuaranteWinningName(g.GetWinningRow())) + "\n")
		}
		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch g.GetPhase() {
		case domain.TrenteEtQuarantePhaseBet:
			b.WriteString(i18n.T("trenteetquarante.promptBet") + "\n")
		case domain.TrenteEtQuarantePhaseResult:
			b.WriteString(p.resultLine(g))
			b.WriteString(i18n.Tf("trenteetquarante.payoutLine",
				"payout", strconv.Itoa(g.GetPayout())) + "\n")
			b.WriteString(i18n.T("trenteetquarante.promptResult") + "\n")
		}
		b.WriteString(i18n.T("trenteetquarante.promptHelp") + "\n")
	})
}

// resultLine はラウンド結果の 1 行 (色付き) を返す。
func (p *TrenteEtQuaranteCuiPresenter) resultLine(g interfaces.TrenteEtQuaranteGame) string {
	if g.GetRefait() {
		// 半額を取られること自体はラベルに出ているが、**なぜ 31 だけ違うのか**は
		// Web の refait explainer にしか無かった (#5696)。
		return color.Yellow(i18n.T("trenteetquarante.result.refait")) + "\n" +
			i18n.T("trenteetquarante.result.refaitWhy") + "\n"
	}
	switch g.GetResult() {
	case domain.TrenteEtQuaranteResultWin:
		return color.Green(i18n.T("trenteetquarante.result.win")) + "\n"
	case domain.TrenteEtQuaranteResultDraw:
		return color.Yellow(i18n.T("trenteetquarante.result.push")) + "\n"
	default:
		return color.Red(i18n.T("trenteetquarante.result.lose")) + "\n"
	}
}

// HintOutput emits the current Trente et Quarante hint.
func (p *TrenteEtQuaranteCuiPresenter) HintOutput(g interfaces.TrenteEtQuaranteGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("trenteetquarante.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, trenteEtQuaranteHintReasonKeys)
	return color.Yellow(i18n.Tf("trenteetquarante.hintBet",
		"bet", trenteEtQuaranteBetName(hint.Bet),
		"reason", reason)) + "\n"
}

// trenteEtQuaranteHintReasonKeys maps hint-reason identifiers to i18n keys.
var trenteEtQuaranteHintReasonKeys = map[string]string{
	"even_odds": "trenteetquarante.hintReasonEvenOdds",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *TrenteEtQuaranteCuiPresenter) ActionLogOutput(g interfaces.TrenteEtQuaranteGame) string {
	return actionLogOutputText(g)
}
