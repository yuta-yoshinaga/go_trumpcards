package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// durakPlayerStr returns the display string for a single Durak player.
func durakPlayerStr(player *domain.DurakPlayer, i int, isAttacker, isDefender bool) string {
	var b strings.Builder
	b.WriteString(cuiPlayerName(player, i))
	if isAttacker {
		b.WriteString(color.BoldYellow(i18n.T("durak.labelAttacker")))
	}
	if isDefender {
		b.WriteString(color.Bold(i18n.T("durak.labelDefender")))
	}
	if player.GetIsFinished() {
		b.WriteString(i18n.T("durak.playerFinished") + "\n")
		return b.String()
	}
	b.WriteString(i18n.Tf("durak.playerHand",
		"count", strconv.Itoa(player.GetCardsSize())) + "\n")
	if player.GetIsHuman() {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// DurakCuiPresenter renders the Durak CUI view.
type DurakCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *DurakCuiPresenter) Output(dg interfaces.DurakGame, lastErr error) string {
	return buildCuiOutput(i18n.T("durak.helpTitle"), func(b *strings.Builder) {
		// Trump info
		b.WriteString(i18n.Tf("durak.trumpLine",
			"suit", cuiSuitName(dg.GetTrumpSuit())))
		if dg.GetTrumpCard() != nil {
			b.WriteString(i18n.Tf("durak.trumpBottom",
				"card", cuiCardStr(dg.GetTrumpCard())))
		}
		b.WriteString(i18n.Tf("durak.stockLine",
			"stock", strconv.Itoa(dg.GetStockCount())) + "\n")

		b.WriteString("----------\n")

		// Players
		for i := 0; i < dg.GetPlayerCnt(); i++ {
			b.WriteString(durakPlayerStr(dg.GetPlayer(i), i,
				i == dg.GetAttackerIdx(), i == dg.GetDefenderIdx()))
		}

		b.WriteString("----------\n")

		// Table
		pairs := dg.GetTablePairs()
		if len(pairs) > 0 {
			b.WriteString(i18n.T("durak.tableHeader") + "\n")
			for i, pair := range pairs {
				if pair.Defense != nil {
					b.WriteString(i18n.Tf("durak.tablePairDefended",
						"idx", strconv.Itoa(i),
						"attack", cuiCardStr(pair.Attack),
						"defense", cuiCardStr(pair.Defense)) + "\n")
				} else {
					b.WriteString(i18n.Tf("durak.tablePairOpen",
						"idx", strconv.Itoa(i),
						"attack", cuiCardStr(pair.Attack)) + "\n")
				}
			}
		} else {
			b.WriteString(i18n.T("durak.tableEmpty") + "\n")
		}

		// Phase
		switch dg.GetPhase() {
		case domain.DurakPhaseAttack:
			b.WriteString(i18n.T("durak.phaseAttack") + "\n")
		case domain.DurakPhaseDefend:
			b.WriteString(i18n.T("durak.phaseDefend") + "\n")
		case domain.DurakPhaseGameEnd:
			b.WriteString(i18n.T("durak.phaseGameEnd") + "\n")
		}

		// Game over
		if dg.GetGameEndFlag() {
			loserIdx := dg.GetLoserIdx()
			if loserIdx < 0 {
				b.WriteString(color.Green(i18n.T("durak.gameEndDraw")) + "\n")
			} else {
				player := dg.GetPlayer(loserIdx)
				if player.GetIsHuman() {
					b.WriteString(color.Red(i18n.T("durak.gameEndHumanLost")) + "\n")
				} else {
					b.WriteString(color.Green(i18n.Tf("durak.gameEndCpuLost",
						"idx", strconv.Itoa(loserIdx))) + "\n")
				}
			}
		}

		cuiErrorBlock(b, lastErr)
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *DurakCuiPresenter) ActionLogOutput(dg interfaces.DurakGame) string {
	return actionLogOutputText(dg)
}
