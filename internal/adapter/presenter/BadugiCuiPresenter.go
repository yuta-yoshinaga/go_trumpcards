package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BadugiCuiPresenter renders Badugi state for the CLI.
type BadugiCuiPresenter struct{}

// Output produces the CUI-rendered state.
func (bcp *BadugiCuiPresenter) Output(g interfaces.BadugiGame, lastErr error) string {
	var b strings.Builder
	players := g.GetPlayers()

	b.WriteString("==========\n")
	b.WriteString("Badugi (4-Card Draw Lowball)\n")
	b.WriteString("==========\n")

	fmt.Fprintf(&b, "ディーラー: Player %d\n", g.GetDealerIdx())
	fmt.Fprintf(&b, "ポット: %d\n", g.GetPot())

	if drawIdx := g.GetDrawIndex(); drawIdx > 0 {
		fmt.Fprintf(&b, "ドロー: %d/%d\n", drawIdx, domain.BadugiMaxDraws)
	} else {
		b.WriteString("ドロー: プリドロー\n")
	}

	cfg := g.GetConfig()
	if int(cfg.BettingLimit) < len(domain.BettingLimitNames) {
		fmt.Fprintf(&b, "リミット: %s\n", domain.BettingLimitNames[cfg.BettingLimit])
	}

	b.WriteString("----------\n")
	isEnd := g.GetPhase() == domain.BadugiPhaseEnd
	for i, pl := range players {
		b.WriteString(cuiPlayerNameWithStyle(pl, i))
		fmt.Fprintf(&b, " チップ:%d", pl.GetChips())
		switch {
		case pl.GetFolded():
			b.WriteString(" " + color.BoldYellow("[フォールド]"))
		case pl.GetAllIn():
			b.WriteString(" " + color.BoldYellow("[オールイン]"))
		}
		if pl.GetCurrentBet() > 0 {
			fmt.Fprintf(&b, " ベット:%d", pl.GetCurrentBet())
		}
		if pl.GetDrawCount() > 0 {
			fmt.Fprintf(&b, " 交換:%d枚", pl.GetDrawCount())
		}
		b.WriteString("\n")

		if pl.GetIsHuman() && !pl.GetFolded() {
			handStr := cuiIndexedCardListStrEmoji(pl)
			if isEnd {
				fmt.Fprintf(&b, "  手札: %s  [%s]\n", handStr, pl.GetHandName())
			} else {
				fmt.Fprintf(&b, "  手札: %s\n", handStr)
			}
		}
		if !pl.GetIsHuman() && isEnd && !pl.GetFolded() {
			fmt.Fprintf(&b, "  手札: %s  [%s]\n", cuiCardListStrEmoji(pl), pl.GetHandName())
		}
	}

	cpuActions := g.GetCpuActions()
	if len(cpuActions) > 0 {
		b.WriteString("----------\n")
		b.WriteString(color.Bold("[CPU行動]") + "\n")
		for _, a := range cpuActions {
			fmt.Fprintf(&b, "  Player %d (%s): %s", a.PlayerIdx, a.RoundLabel, cuiBettingActionName(a.Action))
			if a.Amount > 0 {
				fmt.Fprintf(&b, " (%d)", a.Amount)
			}
			b.WriteString("\n")
		}
	}

	cpuExchanges := g.GetCpuExchanges()
	if len(cpuExchanges) > 0 {
		b.WriteString("----------\n")
		b.WriteString(color.Bold("[CPUドロー]") + "\n")
		for _, e := range cpuExchanges {
			fmt.Fprintf(&b, "  Player %d (draw %d): %d枚交換\n", e.PlayerIdx, e.DrawIndex, e.ExchangeCount)
		}
	}

	results := g.GetRoundResults()
	if len(results) > 0 && isEnd {
		b.WriteString("==========\n")
		b.WriteString(color.Bold("[結果]") + "\n")
		for _, r := range results {
			name := cuiPlayerName(players[r.PlayerIdx], r.PlayerIdx)
			if r.HandName != "" {
				fmt.Fprintf(&b, "  %s: %s", name, r.HandName)
			} else {
				fmt.Fprintf(&b, "  %s", name)
			}
			if r.WonAmount > 0 {
				fmt.Fprintf(&b, " → %dチップ獲得", r.WonAmount)
			}
			b.WriteString("\n")
		}
	}

	if lastErr != nil {
		fmt.Fprintf(&b, "%s\n", color.Red(lastErr.Error()))
	}
	if g.GetGameEndFlag() {
		b.WriteString("ゲーム終了\n")
	}
	return b.String()
}

// ActionLogOutput renders the action log as plain text.
func (bcp *BadugiCuiPresenter) ActionLogOutput(g interfaces.BadugiGame) string {
	return actionLogOutputText(g)
}
