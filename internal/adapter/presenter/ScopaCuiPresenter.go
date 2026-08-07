//go:build !js || !wasm || classic

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// ScopaCuiPresenter renders the Scopa CUI view.
type ScopaCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *ScopaCuiPresenter) Output(sg interfaces.ScopaGame, lastErr error) string {
	return buildCuiOutput(i18n.T("scopa.helpTitle"), func(b *strings.Builder) {
		for i := 0; i < sg.GetPlayerCnt(); i++ {
			b.WriteString(scopaPlayerStr(sg.GetPlayer(i), i))
		}
		b.WriteString("----------\n")

		if tableCards := sg.GetTableCards(); len(tableCards) > 0 {
			b.WriteString(i18n.Tf("scopa.tableLine", "cards", cuiCardSliceStr(tableCards)) + "\n")
		} else {
			b.WriteString(i18n.T("scopa.tableEmpty") + "\n")
		}

		if ha := sg.GetHumanAction(); ha != nil {
			b.WriteString(i18n.Tf("scopa.humanActionLine", "text", scopaActionStr(ha)) + "\n")
		}
		if cpu := sg.GetCpuActions(); len(cpu) > 0 {
			b.WriteString(color.Bold(i18n.T("scopa.cpuActionsHeader")) + "\n")
			for _, a := range cpu {
				b.WriteString(i18n.Tf("scopa.cpuActionLine",
					"name", cuiPlayerName(sg.GetPlayer(a.PlayerIdx), a.PlayerIdx),
					"text", scopaActionStr(a)) + "\n")
			}
		}

		// **なぜその点数になったのかを CUI は一切出していなかった (#4756)。**
		// Web はカルテ/デナリ/プリミエラ/セッテベッロごとに誰が取ったかを
		// 出している。CUI は最終合計しか見えなかった。
		writeScopaBreakdown(b, sg)

		cuiErrorBlock(b, lastErr)

		if sg.GetGameEndFlag() {
			b.WriteString(i18n.T("scopa.gameEnd") + "\n")
			for i := 0; i < sg.GetPlayerCnt(); i++ {
				pl := sg.GetPlayer(i)
				if pl == nil {
					continue
				}
				b.WriteString(i18n.Tf("scopa.scoreEntry",
					"name", cuiPlayerName(pl, i),
					"score", strconv.Itoa(pl.GetTotalScore())) + "\n")
			}
			return
		}
		currentTurn := sg.GetCurrentTurn()
		b.WriteString(i18n.Tf("scopa.promptCurrentTurn",
			"name", cuiPlayerName(sg.GetPlayer(currentTurn), currentTurn)) + "\n")
		b.WriteString(i18n.T("scopa.promptHelp") + "\n")
	})
}

// scopaPlayerStr returns the display string for a single Scopa player.
func scopaPlayerStr(player *domain.ScopaPlayer, i int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("scopa.playerLine",
		"name", cuiPlayerName(player, i),
		"hand", strconv.Itoa(player.GetCardsSize()),
		"captured", strconv.Itoa(player.CapturedCount()),
		"scopa", strconv.Itoa(player.GetScopaCount()),
		"total", strconv.Itoa(player.GetTotalScore())) + "\n")
	if player.GetIsHuman() {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// scopaActionStr renders an action as a short readable line.
func scopaActionStr(a *domain.ScopaAction) string {
	if a == nil {
		return ""
	}
	if len(a.CapturedCards) > 0 {
		suffix := ""
		if a.IsScopa {
			suffix = i18n.T("scopa.actionScopaSuffix")
		}
		return i18n.Tf("scopa.actionCapture",
			"played", scopaCardShort(a.PlayedCard),
			"count", strconv.Itoa(len(a.CapturedCards)),
			"suffix", suffix)
	}
	return i18n.Tf("scopa.actionLay", "played", scopaCardShort(a.PlayedCard))
}

// scopaCardShort renders a single card as a short text representation.
func scopaCardShort(c *domain.Card) string {
	if c == nil {
		return "-"
	}
	return cuiCardSliceStr([]*domain.Card{c})
}

// HintOutput emits a capture recommendation for the human's turn: the hand
// card and the table cards it captures (flagging a scopa when it clears the
// table), reusing the domain's EnumerateScopaCaptures.
func (p *ScopaCuiPresenter) HintOutput(sg interfaces.ScopaGame) string {
	if sg.GetPhase() != domain.ScopaPhasePlayerTurn {
		return i18n.T("scopa.hintNone") + "\n"
	}
	turn := sg.GetCurrentTurn()
	player := sg.GetPlayer(turn)
	if player == nil || !player.GetIsHuman() {
		return i18n.T("scopa.hintNone") + "\n"
	}
	table := sg.GetTableCards()
	bestHand := -1
	var bestCap []int
	bestScopa := false
	for i := 0; i < player.GetCardsSize(); i++ {
		for _, cap := range domain.EnumerateScopaCaptures(player.GetCard(i), table) {
			isScopa := len(table) > 0 && len(cap) == len(table)
			switch {
			case bestHand == -1:
			case isScopa && !bestScopa:
			case isScopa == bestScopa && len(cap) > len(bestCap):
			default:
				continue
			}
			bestHand, bestCap, bestScopa = i, cap, isScopa
		}
	}
	if bestHand == -1 {
		return color.Yellow(i18n.T("scopa.hintNoCapture")) + "\n"
	}
	capCards := make([]*domain.Card, 0, len(bestCap))
	for _, idx := range bestCap {
		capCards = append(capCards, table[idx])
	}
	key := "scopa.hintCapture"
	if bestScopa {
		key = "scopa.hintScopa"
	}
	return color.Yellow(i18n.Tf(key,
		"played", scopaCardShort(player.GetCard(bestHand)),
		"captured", cuiCardSliceStr(capCards))) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *ScopaCuiPresenter) ActionLogOutput(sg interfaces.ScopaGame) string {
	return actionLogOutputText(sg)
}

// writeScopaBreakdown は直前ラウンドの得点内訳を書く。まだ1ラウンドも
// 終わっていなければ何も書かない。
func writeScopaBreakdown(b *strings.Builder, sg interfaces.ScopaGame) {
	det := sg.GetLastRoundDetail()
	if det == nil {
		return
	}
	rows := domain.ScopaCategoryWinners(det)
	if len(rows) == 0 {
		return
	}
	b.WriteString("----------\n")
	b.WriteString(i18n.T("scopa.breakdownHeader") + "\n")
	for _, row := range rows {
		// **同点・該当なしも書く。**行が消えると「誰かが取った」と読める。
		name := i18n.T("scopa.breakdownNobody")
		if row.Winner >= 0 {
			if pl := sg.GetPlayer(row.Winner); pl != nil {
				name = cuiPlayerName(pl, row.Winner)
			}
		}
		b.WriteString(i18n.Tf("scopa.breakdownRow",
			"category", i18n.T("scopa.category."+row.Key),
			"name", name,
			"points", strconv.Itoa(row.Points)) + "\n")
	}
	// スコパ回数は単独勝者ではなく人数ぶんの回数なので別に出す。
	for i := 0; i < sg.GetPlayerCnt(); i++ {
		n := det.Scopas[i]
		if n == 0 {
			continue
		}
		pl := sg.GetPlayer(i)
		if pl == nil {
			continue
		}
		b.WriteString(i18n.Tf("scopa.breakdownScopa",
			"name", cuiPlayerName(pl, i),
			"count", strconv.Itoa(n)) + "\n")
	}
}
