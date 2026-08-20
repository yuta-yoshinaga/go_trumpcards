//go:build !js || !wasm || extra3

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// maoPlayerStr returns the display string for a single Mao player.
func maoPlayerStr(player *domain.MaoPlayer, i int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("mao.playerLine",
		"name", cuiPlayerName(player, i),
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize())) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// maoDirectionName returns the human-readable play-direction marker.
func maoDirectionName(direction int) string {
	if direction < 0 {
		return "←"
	}
	return "→"
}

// MaoCuiPresenter renders the Mao CUI view. The secret rule itself is never
// rendered; only the unlocked half-hint and penalty notices are shown.
type MaoCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *MaoCuiPresenter) Output(g interfaces.MaoGame, lastErr error) string {
	return buildCuiOutput(i18n.T("mao.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("mao.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"stock", strconv.Itoa(g.GetDrawPileCount()),
			"dir", maoDirectionName(g.GetDirection())) + "\n")

		// Top of discard pile
		if top := g.GetDiscardTop(); top != nil {
			b.WriteString(i18n.Tf("mao.discardLine", "card", cuiCardStr(top)))
			if g.GetChosenSuit() > 0 {
				b.WriteString(i18n.Tf("mao.chosenSuit",
					"suit", suitDisplayName(g.GetChosenSuit())))
			}
			b.WriteString("\n")
		}

		if g.GetPenaltyDrawCount() > 0 {
			b.WriteString(i18n.Tf("mao.penaltyLine",
				"count", strconv.Itoa(g.GetPenaltyDrawCount())) + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(maoPlayerStr(g.GetPlayer(i), i))
		}

		// 隠しルールのハーフヒント (3回正解で解放)。解放前は正解の進捗を表示する。
		if g.GetHintUnlocked() {
			if key := g.GetRuleHintKey(); key != "" {
				// **キーの翻訳はここで行う。**ドメインは文言を持たない (#4917)。
				b.WriteString(color.Yellow(i18n.Tf("mao.ruleHint", "hint", i18n.T("mao."+key))) + "\n")
			}
		} else {
			b.WriteString(i18n.Tf("mao.compliance",
				"count", strconv.Itoa(g.GetPlayerCorrectCount()),
				"total", strconv.Itoa(domain.MaoHintThreshold)) + "\n")
		}
		if g.GetRulePenaltyFlag() {
			b.WriteString(color.Red(i18n.T("mao.rulePenalty")) + "\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerIdx := g.GetWinnerIdx()
			banner := i18n.Tf("mao.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		if g.GetAwaitingWord() {
			b.WriteString(i18n.T("mao.promptAwaitingWord") + "\n")
			b.WriteString(i18n.T("mao.promptAwaitingWordHelp") + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.MaoPhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("mao.promptCurrentPlayer",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("mao.promptPlayHelp") + "\n")
			b.WriteString(i18n.T("mao.promptDrawHelp") + "\n")
		case domain.MaoPhaseChooseSuit:
			b.WriteString(i18n.T("mao.promptChooseSuit") + "\n")
			b.WriteString(i18n.T("mao.promptChooseSuitHelp") + "\n")
		case domain.MaoPhaseMustDeclare:
			b.WriteString(i18n.T("mao.promptMustDeclare") + "\n")
			b.WriteString(i18n.T("mao.promptMustDeclareHelp") + "\n")
		case domain.MaoPhaseRoundEnd:
			b.WriteString(i18n.T("mao.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("mao.promptRoundEndHelp") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *MaoCuiPresenter) ActionLogOutput(g interfaces.MaoGame) string {
	return actionLogOutputTextForSeats[*domain.MaoPlayer](g)
}
