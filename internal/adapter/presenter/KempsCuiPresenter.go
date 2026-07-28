//go:build !js || !wasm || extra2

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// KempsCuiPresenter renders the Kemps CUI view.
type KempsCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *KempsCuiPresenter) Output(g interfaces.KempsGame, lastErr error) string {
	return buildCuiOutput(i18n.T("kemps.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("kemps.roundLine",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"teamA", strconv.Itoa(g.GetTeamScore(domain.KempsTeamOf(0))),
			"teamB", strconv.Itoa(g.GetTeamScore(1-domain.KempsTeamOf(0)))) + "\n")
		b.WriteString(i18n.Tf("kemps.signalLine",
			"signal", kempsSignalName(g.GetSignalType())) + "\n")
		b.WriteString("----------\n")

		// フィールド (場のカード) を表示する。
		field := make([]*domain.Card, 0, g.GetFieldSize())
		for i := 0; i < g.GetFieldSize(); i++ {
			field = append(field, g.GetFieldCard(i))
		}
		b.WriteString(i18n.T("kemps.fieldLabel") + " " + cuiIndexedCardListStr(kempsCardSlice(field)) + "\n")
		b.WriteString("----------\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			player := g.GetPlayer(i)
			if player == nil {
				continue
			}
			name := cuiPlayerName(player, i)
			team := i18n.Tf("kemps.teamLabel", "team", kempsTeamName(domain.KempsTeamOf(i)))
			if i == 0 {
				// 人間の手札のみ公開表示する。
				b.WriteString(name + " " + team + "\n")
				b.WriteString("  " + cuiIndexedCardListStr(player) + "\n")
			} else {
				b.WriteString(i18n.Tf("kemps.cpuHandLine",
					"name", name,
					"count", strconv.Itoa(player.GetCardsSize())) + " " + team + "\n")
			}
		}
		b.WriteString("----------\n")

		switch g.GetPhase() {
		case domain.KempsPhaseDeclare:
			if g.IsPartnerSignaling() {
				b.WriteString(color.Green(i18n.T("kemps.promptPartnerSignal")) + "\n")
			} else if g.IsOpponentSignaling() {
				b.WriteString(color.Yellow(i18n.T("kemps.promptOpponentSignal")) + "\n")
			}
			b.WriteString(i18n.T("kemps.promptDeclare") + "\n")
		case domain.KempsPhaseExchange:
			if g.IsHumanTurn() {
				b.WriteString(i18n.T("kemps.promptExchange") + "\n")
			} else {
				b.WriteString(i18n.T("kemps.promptCpuTurn") + "\n")
			}
		case domain.KempsPhaseRoundEnd:
			b.WriteString(i18n.T("kemps.promptNextRound") + "\n")
		}

		if g.GetGameEndFlag() {
			if g.GetWinnerTeam() == domain.KempsTeamOf(0) {
				b.WriteString(color.Green(i18n.T("kemps.winHuman")) + "\n")
			} else {
				b.WriteString(color.Red(i18n.T("kemps.winCpu")) + "\n")
			}
		}

		cuiErrorBlock(b, lastErr)
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *KempsCuiPresenter) ActionLogOutput(g interfaces.KempsGame) string {
	return actionLogOutputText(g)
}

// kempsCardSlice wraps []*Card to satisfy the cuiCardList interface.
type kempsCardSlice []*domain.Card

func (s kempsCardSlice) GetCardsSize() int          { return len(s) }
func (s kempsCardSlice) GetCard(i int) *domain.Card { return s[i] }

// kempsSignalName はシグナル種別の表示名を返す。
func kempsSignalName(st domain.SignalType) string {
	if st == domain.SignalBlink {
		return i18n.T("kemps.signalBlink")
	}
	return i18n.T("kemps.signalSound")
}

// kempsTeamName はチーム番号の表示名 (A / B) を返す。
func kempsTeamName(team int) string {
	if team == domain.KempsTeamOf(0) {
		return "A"
	}
	return "B"
}
