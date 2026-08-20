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

// killeReveal は全員の手札を公開する局面かを返す。
func killeReveal(g interfaces.KilleGame) bool {
	phase := g.GetPhase()
	return phase == domain.KillePhaseShowdown || phase == domain.KillePhaseGameEnd
}

// killeCardStr は専用デッキの札を名前で描く。
//
// **標準52枚の記法は使えない。**単一スートなので ♠A のような表記に意味が無い。
func killeCardStr(card *domain.Card) string {
	r := domain.KilleRankOf(card)
	if r == 0 {
		return i18n.T("kille.hiddenCard")
	}
	return domain.KilleRankName(r)
}

// killePlayerStr returns the display string for a single Kille player.
func killePlayerStr(g interfaces.KilleGame, player *domain.KillePlayer, i int) string {
	card := i18n.T("kille.hiddenCard")
	switch {
	case player.GetIsFinished():
		card = i18n.T("kille.eliminated")
	case player.GetIsHuman() || killeReveal(g):
		if player.GetCardsSize() > 0 {
			card = killeCardStr(player.GetCard(0))
		}
	}
	role := ""
	if i == g.GetDealerIdx() {
		role = " " + i18n.T("kille.dealerTag")
	}
	state := ""
	switch {
	case player.IsOut() && player.GetKnockedBy() == domain.KilleKnockHussar:
		state = " " + i18n.T("kille.outHussar")
	case player.IsOut() && player.GetKnockedBy() == domain.KilleKnockPig:
		state = " " + i18n.T("kille.outPig")
	case player.IsOut():
		state = " " + i18n.T("kille.out")
	case player.IsSatisfied():
		state = " " + i18n.T("kille.satisfiedTag")
	}
	// 買い戻しは上限つき (KilleMaxReentries)。**残り回数が判断材料**なので、
	// 使った回数だけでなく上限も添える (Web の reentriesUsed と同じ)。
	if used := player.GetReentries(); used > 0 {
		state += " " + i18n.Tf("kille.reentriesUsed",
			"used", strconv.Itoa(used), "max", strconv.Itoa(domain.KilleMaxReentries))
	}
	return i18n.Tf("kille.playerLine",
		"name", cuiPlayerName(player, i),
		"chips", strconv.Itoa(player.GetChips()),
		"card", card,
		"role", role,
		"state", state) + "\n"
}

// killeWriteEvents appends this round's exchange log (nothing when the round is quiet).
func killeWriteEvents(b *strings.Builder, g interfaces.KilleGame) {
	events := g.GetEvents()
	if len(events) == 0 {
		return
	}
	b.WriteString(i18n.T("kille.eventsTitle") + "\n")
	for _, e := range events {
		if e == nil {
			continue
		}
		target := i18n.T("kille.stockLabel")
		if e.Target >= 0 {
			target = cuiPlayerName(g.GetPlayer(e.Target), e.Target)
		}
		b.WriteString("  " + i18n.Tf("kille.event."+e.Kind,
			"actor", cuiPlayerName(g.GetPlayer(e.Actor), e.Actor),
			"target", target) + "\n")
	}
}

// KilleCuiPresenter renders the Kille CUI view.
type KilleCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *KilleCuiPresenter) Output(g interfaces.KilleGame, lastErr error) string {
	return buildCuiOutput(i18n.T("kille.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("kille.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"pot", strconv.Itoa(g.GetPot()),
			"stock", strconv.Itoa(g.GetStockCount())) + "\n")

		for i := range g.GetPlayers() {
			b.WriteString(killePlayerStr(g, g.GetPlayer(i), i))
		}

		// 誰が誰と交換したかはこのゲームの読みの中心 (カッコウ・軽騎兵・豚の
		// 判明状況がここにしか出ない)。Web は kille-events で出している。
		killeWriteEvents(b, g)

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			banner := i18n.Tf("kille.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(g.GetWinnerIdx()), g.GetWinnerIdx()))
			b.WriteString(color.Green(banner) + "\n")
			return
		}

		switch g.GetPhase() {
		case domain.KillePhaseExchange:
			idx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("kille.promptTurn", "name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
			if idx == g.GetDealerIdx() {
				b.WriteString(i18n.T("kille.promptTurnHelpDealer") + "\n")
			} else {
				b.WriteString(i18n.T("kille.promptTurnHelp") + "\n")
			}
		case domain.KillePhaseShowdown:
			b.WriteString(i18n.Tf("kille.promptShowdown",
				"count", strconv.Itoa(len(g.GetLoserIdxs()))) + "\n")
			b.WriteString(i18n.T("kille.promptShowdownHelp") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *KilleCuiPresenter) ActionLogOutput(g interfaces.KilleGame) string {
	return actionLogOutputTextForSeats[*domain.KillePlayer](g)
}
