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

// durakHandStr renders the human's indexed hand, emphasizing trump cards so the
// player can spot which cards beat any non-trump in attack and defense.
func durakHandStr(player *domain.DurakPlayer, trumpSuit int) string {
	parts := make([]string, player.GetCardsSize())
	for i := range parts {
		card := player.GetCard(i)
		s := fmt.Sprintf("[%d]%s", i, cuiCardStr(card))
		if card != nil && card.GetDesign() == trumpSuit {
			s = color.BoldYellow(s)
		}
		parts[i] = s
	}
	return strings.Join(parts, "  ")
}

// durakPlayerStr returns the display string for a single Durak player.
func durakPlayerStr(player *domain.DurakPlayer, i int, isAttacker, isDefender bool, trumpSuit int) string {
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
		b.WriteString(durakHandStr(player, trumpSuit) + "\n")
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
				i == dg.GetAttackerIdx(), i == dg.GetDefenderIdx(), dg.GetTrumpSuit()))
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
	return actionLogOutputTextForSeats[*domain.DurakPlayer](dg)
}

// HintOutput emits the current Durak hint.
//
// **他のトリック系はサーバー計算の理由付きヒントを持つのに、Durak は CUI に
// hint コマンドすら無かった (#4740)。**
func (p *DurakCuiPresenter) HintOutput(g interfaces.DurakGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("durak.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, durakHintReasonKeys)
	if hint.TakeCards {
		return color.Yellow(i18n.Tf("durak.hintTake", "reason", reason)) + "\n"
	}
	if hint.CardIndex == nil {
		return color.Yellow(i18n.Tf("durak.hintPass", "reason", reason)) + "\n"
	}
	card := g.GetPlayer(g.GetCurrentTurn()).GetCard(*hint.CardIndex)
	if hint.AttackIdx != nil {
		return color.Yellow(i18n.Tf("durak.hintDefend",
			"idx", strconv.Itoa(*hint.CardIndex),
			"card", cuiCardStr(card),
			"atk", strconv.Itoa(*hint.AttackIdx),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("durak.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", reason)) + "\n"
}

// durakHintReasonKeys maps hint-reason identifiers to their i18n keys.
var durakHintReasonKeys = map[string]string{
	"attack_weakest":    "durak.hintReasonAttackWeakest",
	"attack_additional": "durak.hintReasonAttackAdditional",
	"defend_beat":       "durak.hintReasonDefendBeat",
	"take_cannot_beat":  "durak.hintReasonTakeCannotBeat",
	"pass_no_addition":  "durak.hintReasonPassNoAddition",
}
