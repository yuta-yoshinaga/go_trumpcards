package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// speedHumanIdx is the human seat. Speed is strictly two-handed: seat 0 is the
// player, seat 1 the CPU.
const speedHumanIdx = 0

// speedIndexedHandStr renders the human's hand, marking the cards that can go
// onto one of the two centre piles right now. The Web GUI rings exactly these
// cards for the whole PLAY phase, independent of the hint feature; the CUI
// printed a bare indexed list and left the ±1 (and K↔A) comparison against both
// piles to the player, every turn, under time pressure (#4861).
//
// 索引と区切りは他ゲームと同じ formatCardList に任せ、印だけを足す
// (OmbreCuiPresenter / SpoonsCuiPresenter と同じ形)。
func speedIndexedHandStr(s interfaces.SpeedGame, human cuiCardList) string {
	idx := -1
	return formatCardList(human, func(c *domain.Card) string {
		idx++
		out := cuiCardStr(c)
		// 判定はドメインの CanPlay をそのまま呼ぶ。ここで ±1 を書き直すと
		// K↔A のラップが片方だけずれる。
		for pile := range domain.SpeedCenterPileCnt {
			if s.CanPlay(speedHumanIdx, idx, pile) {
				return out + i18n.T("speed.playableMark")
			}
		}
		return out
	}, "  ", true)
}

// SpeedCuiPresenter renders the Speed CUI view.
type SpeedCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *SpeedCuiPresenter) Output(s interfaces.SpeedGame, lastErr error) string {
	return buildCuiOutput(i18n.T("speed.helpTitle"), func(b *strings.Builder) {
		// CPU info
		cpu := s.GetPlayer(1)
		b.WriteString(i18n.Tf("speed.cpuStats",
			"hand", strconv.Itoa(cpu.GetCardsSize()),
			"draw", strconv.Itoa(cpu.GetDrawPileSize())) + "\n")

		// Center piles
		b.WriteString("----------\n")
		b.WriteString(color.Bold(i18n.T("speed.centerLabel")) + " ")
		for i := range 2 {
			c := s.GetCenterPile(i)
			if c != nil {
				b.WriteString(i18n.Tf("speed.centerCard",
					"idx", strconv.Itoa(i),
					"card", cuiCardStr(c)))
			}
		}
		b.WriteString("\n")
		b.WriteString("----------\n")

		// Human player info
		human := s.GetPlayer(0)
		b.WriteString(i18n.Tf("speed.humanStats",
			"hand", strconv.Itoa(human.GetCardsSize()),
			"draw", strconv.Itoa(human.GetDrawPileSize())) + "\n")
		b.WriteString(speedIndexedHandStr(s, human) + "\n")

		// Hint
		ci, pi, found := s.GetHint()
		if found {
			b.WriteString(i18n.Tf("speed.hintLine",
				"ci", strconv.Itoa(ci),
				"pi", strconv.Itoa(pi)) + "\n")
		}

		// Phase state
		switch s.GetPhase() {
		case domain.SpeedPhaseStuck:
			b.WriteString(color.Yellow(i18n.T("speed.promptStuck")) + "\n")
			b.WriteString(i18n.T("speed.promptStuckHelp") + "\n")
		}

		// Outcome
		if s.GetGameEndFlag() {
			if s.GetWinnerIdx() == 0 {
				b.WriteString(color.Green(i18n.T("speed.winHuman")) + "\n")
			} else {
				b.WriteString(color.Red(i18n.T("speed.winCpu")) + "\n")
			}
		}

		cuiErrorBlock(b, lastErr)
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *SpeedCuiPresenter) ActionLogOutput(s interfaces.SpeedGame) string {
	return actionLogOutputText(s)
}
