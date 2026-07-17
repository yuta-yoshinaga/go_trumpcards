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

// MichiganCuiPresenter renders the Michigan CUI view.
type MichiganCuiPresenter struct{}

// michiganBoodleStr は 1 つのブードルの表示文字列を返す。
func michiganBoodleStr(g interfaces.MichiganGame, i int) string {
	b := g.GetBoodle(i)
	if b == nil {
		return ""
	}
	status := i18n.T("michigan.boodleOpen")
	if b.GetClaimedBy() >= 0 {
		status = i18n.Tf("michigan.boodleClaimed", "player", strconv.Itoa(b.GetClaimedBy()))
	}
	return i18n.Tf("michigan.boodleLine",
		"card", cuiCardStr(b.GetCard()),
		"chips", strconv.Itoa(b.GetChips()),
		"status", status,
	)
}

// michiganPlayerStr は 1 プレイヤーの表示文字列を返す。
func michiganPlayerStr(g interfaces.MichiganGame, i int) string {
	player := g.GetPlayer(i)
	if player == nil {
		return ""
	}
	reveal := g.GetPhase() == domain.MichiganPhaseResult
	showCards := player.GetIsHuman() || reveal

	var b strings.Builder
	status := i18n.T("michigan.statusWaiting")
	if g.GetPhase() == domain.MichiganPhasePlay && g.GetCurrentPlayerIdx() == i {
		status = i18n.T("michigan.statusTurn")
	}
	if i == g.GetWinnerIdx() {
		status = i18n.T("michigan.statusWinner")
	}
	b.WriteString(i18n.Tf("michigan.playerLine",
		"name", cuiPlayerName(player, i),
		"chips", strconv.Itoa(player.GetChips()),
		"bet", strconv.Itoa(player.GetRoundBet()),
		"count", strconv.Itoa(player.GetCardsSize()),
		"status", status,
	))
	b.WriteString("\n")
	if showCards && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// michiganBetHintStr suggests which boodles to weight when placing bets: a boodle
// is worth backing when the human already holds the exact card that claims it. The
// boodle index matches the positional argument of the `b <♥> <♣> <♦> <♠>` command.
func michiganBetHintStr(g interfaces.MichiganGame) string {
	human := -1
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if pl := g.GetPlayer(i); pl != nil && pl.GetIsHuman() {
			human = i
			break
		}
	}
	if human < 0 {
		return ""
	}
	player := g.GetPlayer(human)
	held := make([]string, 0, g.GetBoodleCnt())
	for i := 0; i < g.GetBoodleCnt(); i++ {
		bd := g.GetBoodle(i)
		if bd == nil || bd.GetCard() == nil {
			continue
		}
		bc := bd.GetCard()
		for j := 0; j < player.GetCardsSize(); j++ {
			hc := player.GetCard(j)
			if hc != nil && hc.GetDesign() == bc.GetDesign() && hc.GetValue() == bc.GetValue() {
				held = append(held, i18n.Tf("michigan.betHintHold",
					"card", cuiCardStr(bc), "idx", strconv.Itoa(i)))
				break
			}
		}
	}
	if len(held) == 0 {
		return i18n.T("michigan.betHintNone")
	}
	return i18n.T("michigan.betHintHeader") + " " + strings.Join(held, ", ")
}

// Output renders the current game state for the active locale.
func (p *MichiganCuiPresenter) Output(g interfaces.MichiganGame, lastErr error) string {
	return buildCuiOutput(i18n.T("michigan.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("michigan.roundLine",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"ante", strconv.Itoa(g.GetAnte()),
		) + "\n")

		b.WriteString(i18n.T("michigan.boodleHeader") + "\n")
		for i := 0; i < g.GetBoodleCnt(); i++ {
			b.WriteString("  " + michiganBoodleStr(g, i) + "\n")
		}

		if g.GetPhase() == domain.MichiganPhasePlay && g.GetSeqSuit() != 0 {
			seqCard := domain.NewCard(g.GetSeqSuit(), g.GetSeqHighValue(), false)
			b.WriteString(i18n.Tf("michigan.seqLine", "card", cuiCardStr(seqCard)) + "\n")
		}
		b.WriteString(i18n.Tf("michigan.deadHandLine", "count", strconv.Itoa(g.GetDeadHandCount())) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(michiganPlayerStr(g, i))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			banner := i18n.Tf("michigan.gameEnd", "player", strconv.Itoa(g.GetMatchWinnerIdx()))
			b.WriteString(color.Green(banner) + "\n")
			return
		}

		switch g.GetPhase() {
		case domain.MichiganPhaseBet:
			b.WriteString(i18n.Tf("michigan.promptBet", "budget", strconv.Itoa(g.GetBetBudget())) + "\n")
			b.WriteString(michiganBetHintStr(g) + "\n")
		case domain.MichiganPhasePlay:
			if g.IsHumanTurn() {
				b.WriteString(i18n.Tf("michigan.promptPlay", "indices", michiganIndicesStr(g.GetPlayableIndices())) + "\n")
			} else {
				b.WriteString(i18n.T("michigan.promptWait") + "\n")
			}
		case domain.MichiganPhaseResult:
			b.WriteString(p.resultLine(g))
			b.WriteString(i18n.T("michigan.promptResult") + "\n")
		}
		b.WriteString(i18n.T("michigan.promptHelp") + "\n")
	})
}

// michiganIndicesStr は手札インデックス列をカンマ区切り文字列に変換する。
func michiganIndicesStr(indices []int) string {
	if len(indices) == 0 {
		return "-"
	}
	parts := make([]string, len(indices))
	for i, v := range indices {
		parts[i] = strconv.Itoa(v)
	}
	return strings.Join(parts, ", ")
}

// resultLine はラウンド結果の 1 行 (色付き) を返す。
func (p *MichiganCuiPresenter) resultLine(g interfaces.MichiganGame) string {
	switch g.GetResult() {
	case domain.MichiganResultWin:
		return color.Green(i18n.T("michigan.result.win")) + "\n"
	case domain.MichiganResultLose:
		return color.Red(i18n.T("michigan.result.lose")) + "\n"
	default:
		return color.Yellow(i18n.T("michigan.result.none")) + "\n"
	}
}

// michiganHintReasonKeys maps hint-reason identifiers to i18n keys.
var michiganHintReasonKeys = map[string]string{
	"forced":       "michigan.hintReasonForced",
	"claim_boodle": "michigan.hintReasonClaimBoodle",
	"lead_low":     "michigan.hintReasonLeadLow",
}

// HintOutput emits the current Michigan hint.
func (p *MichiganCuiPresenter) HintOutput(g interfaces.MichiganGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("michigan.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, michiganHintReasonKeys)
	return color.Yellow(i18n.Tf("michigan.hint", "index", strconv.Itoa(hint.CardIndex), "reason", reason)) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *MichiganCuiPresenter) ActionLogOutput(g interfaces.MichiganGame) string {
	return actionLogOutputText(g)
}
