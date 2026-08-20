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

// cuckooReveal ラウンド終了/ゲーム終了時に全員の手札を公開するか
func cuckooReveal(g interfaces.CuckooGame) bool {
	phase := g.GetPhase()
	return phase == domain.CuckooPhaseRoundEnd || phase == domain.CuckooPhaseGameEnd
}

// cuckooPlayerStr returns the display string for a single Cuckoo player.
func cuckooPlayerStr(g interfaces.CuckooGame, player *domain.CuckooPlayer, i int) string {
	reveal := cuckooReveal(g)
	card := i18n.T("cuckoo.hiddenCard")
	switch {
	case player.IsEliminated():
		card = i18n.T("cuckoo.out")
	case player.GetIsHuman() || reveal || g.IsKingRevealed(i):
		card = cuiCardStr(player.Card())
	}
	role := ""
	if i == g.GetDealerIdx() {
		role = " " + i18n.T("cuckoo.dealerTag")
	}
	return i18n.Tf("cuckoo.playerLine",
		"name", cuiPlayerName(player, i),
		"lives", strconv.Itoa(maxInt0(player.GetLives())),
		"card", card,
		"role", role) + "\n"
}

// maxInt0 clamps negative lives to 0 for display.
func maxInt0(n int) int {
	if n < 0 {
		return 0
	}
	return n
}

// CuckooCuiPresenter renders the Cuckoo CUI view.
type CuckooCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *CuckooCuiPresenter) Output(g interfaces.CuckooGame, lastErr error) string {
	return buildCuiOutput(i18n.T("cuckoo.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("cuckoo.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"stock", strconv.Itoa(g.GetStockCount())) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(cuckooPlayerStr(g, g.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			banner := i18n.Tf("cuckoo.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(g.GetWinnerIdx()), g.GetWinnerIdx()))
			b.WriteString(color.Green(banner) + "\n")
			return
		}

		switch g.GetPhase() {
		case domain.CuckooPhaseTurn:
			idx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("cuckoo.promptTurn", "name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
			if idx == g.GetDealerIdx() {
				b.WriteString(i18n.T("cuckoo.promptTurnHelpDealer") + "\n")
			} else {
				b.WriteString(i18n.T("cuckoo.promptTurnHelp") + "\n")
			}
		case domain.CuckooPhaseRefuse:
			to := g.GetPendingSwapTo()
			b.WriteString(i18n.Tf("cuckoo.promptRefuse", "name", cuiPlayerName(g.GetPlayer(to), to)) + "\n")
			b.WriteString(i18n.T("cuckoo.promptRefuseHelp") + "\n")
		case domain.CuckooPhaseRoundEnd:
			b.WriteString(i18n.Tf("cuckoo.promptRoundEnd", "lowest", strconv.Itoa(g.GetRoundLowest())) + "\n")
			b.WriteString(i18n.T("cuckoo.promptRoundEndHelp") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *CuckooCuiPresenter) ActionLogOutput(g interfaces.CuckooGame) string {
	return actionLogOutputTextForSeats[*domain.CuckooPlayer](g)
}
