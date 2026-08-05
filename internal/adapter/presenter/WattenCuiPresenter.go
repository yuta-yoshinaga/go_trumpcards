//go:build !js || !wasm || extra

package presenter

import (
	"sort"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// wattenPlayerStr returns the display string for a single Watten player.
func wattenPlayerStr(player *domain.WattenPlayer, i int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("watten.playerLine",
		"name", cuiPlayerName(player, i),
		"team", strconv.Itoa(player.GetTeam()),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// wattenDeclarePreview renders how many of the human's cards each declaration
// would turn into trumps. The Web GUI previews this live while the dealer picks
// (#4848); the CUI prompt showed only the command syntax, and the declaration
// decides who holds the initiative for the whole deal.
func wattenDeclarePreview(g interfaces.WattenGame) string {
	var human *domain.WattenPlayer
	for i := range g.GetPlayerCnt() {
		if p := g.GetPlayer(i); p != nil && p.GetIsHuman() {
			human = p
			break
		}
	}
	if human == nil || human.GetCardsSize() == 0 {
		return ""
	}
	cards := make([]*domain.Card, 0, human.GetCardsSize())
	for i := range human.GetCardsSize() {
		cards = append(cards, human.GetCard(i))
	}
	// 分類はドメインに置いた。ここで書き直すと Max/Belli/Spitz の二重計上や
	// Web との食い違いが起きる。
	pv := domain.WattenPreviewTrumps(cards)

	suitParts := make([]string, 0, domain.CardDesignDiamond)
	for suit := domain.CardDesignSpade; suit <= domain.CardDesignDiamond; suit++ {
		suitParts = append(suitParts, cuiSuitName(suit)+" "+strconv.Itoa(pv.BySuit[suit]))
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("watten.declareSuitPreview",
		"suits", strings.Join(suitParts, "  "),
		"permanent", strconv.Itoa(pv.Permanent)) + "\n")

	// Schlag は手札に無いランクを並べても仕方がないので、持っているものだけ。
	ranks := make([]int, 0, len(pv.ByRank))
	for r := range pv.ByRank {
		ranks = append(ranks, r)
	}
	sort.Ints(ranks)
	rankParts := make([]string, 0, len(ranks))
	for _, r := range ranks {
		rankParts = append(rankParts, cuiRankLabel(r)+" "+strconv.Itoa(pv.ByRank[r]))
	}
	if len(rankParts) > 0 {
		b.WriteString(i18n.Tf("watten.declareRankPreview", "ranks", strings.Join(rankParts, "  ")) + "\n")
	}
	return b.String()
}

// WattenCuiPresenter renders the Watten CUI view.
type WattenCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *WattenCuiPresenter) Output(g interfaces.WattenGame, lastErr error) string {
	return buildCuiOutput(i18n.T("watten.helpTitle"), func(out *strings.Builder) {
		out.WriteString(i18n.Tf("watten.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber())) + "\n")
		dealerIdx := g.GetDealerIdx()
		out.WriteString(i18n.Tf("watten.dealer",
			"name", cuiPlayerName(g.GetPlayer(dealerIdx), dealerIdx)) + "\n")

		if g.GetCriticalSuit() > 0 {
			out.WriteString(i18n.Tf("watten.declaredLine",
				"rank", cuiRankLabel(g.GetSchlagRank()),
				"suit", cuiSuitName(g.GetCriticalSuit())) + "\n")
		} else {
			out.WriteString(i18n.T("watten.undeclared") + "\n")
		}

		out.WriteString(i18n.Tf("watten.stakeLine",
			"stake", strconv.Itoa(g.GetStake())) + "\n")
		out.WriteString(i18n.Tf("watten.teamScoreLine",
			"t0", strconv.Itoa(g.GetTeamScore(0)),
			"t1", strconv.Itoa(g.GetTeamScore(1))) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			out.WriteString(wattenPlayerStr(g.GetPlayer(i), i))
		}

		out.WriteString("----------\n")

		trick := g.GetCurrentTrick()
		cuiTrickBlock(out, trick,
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(out, lastErr)

		if g.GetGameEndFlag() {
			banner := i18n.Tf("watten.gameEnd", "team", strconv.Itoa(g.GetWinnerTeam()))
			out.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.WattenPhaseDeclare:
			out.WriteString(i18n.Tf("watten.promptDeclare",
				"name", cuiPlayerName(g.GetPlayer(dealerIdx), dealerIdx)) + "\n")
			out.WriteString(wattenDeclarePreview(g))
			out.WriteString(i18n.T("watten.promptDeclareHelp") + "\n")
		case domain.WattenPhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			out.WriteString(i18n.Tf("watten.promptCurrentPlayer",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			if g.CanHumanRaise() {
				out.WriteString(i18n.T("watten.promptRaiseHelp") + "\n")
			}
			out.WriteString(i18n.T("watten.promptPlayHelp") + "\n")
		case domain.WattenPhaseRespond:
			respIdx := g.GetResponderIdx()
			out.WriteString(i18n.Tf("watten.promptRespond",
				"name", cuiPlayerName(g.GetPlayer(respIdx), respIdx),
				"stake", strconv.Itoa(g.GetPendingStake())) + "\n")
			out.WriteString(i18n.T("watten.promptRespondHelp") + "\n")
		case domain.WattenPhaseRoundEnd:
			out.WriteString(i18n.Tf("watten.promptRoundEnd",
				"team", strconv.Itoa(g.GetDealWinnerTeam())) + "\n")
			out.WriteString(i18n.T("watten.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Watten hint.
func (p *WattenCuiPresenter) HintOutput(g interfaces.WattenGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("watten.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, nil)
	switch hint.Action {
	case "declare":
		rank, suit := 0, 0
		if hint.Rank != nil {
			rank = *hint.Rank
		}
		if hint.Suit != nil {
			suit = *hint.Suit
		}
		return color.Yellow(i18n.Tf("watten.hintDeclare",
			"rank", cuiRankLabel(rank),
			"suit", cuiSuitName(suit),
			"reason", reason)) + "\n"
	case "raise":
		return color.Yellow(i18n.Tf("watten.hintRaise", "reason", reason)) + "\n"
	case "hold":
		return color.Yellow(i18n.Tf("watten.hintHold", "reason", reason)) + "\n"
	case "fold":
		return color.Yellow(i18n.Tf("watten.hintFold", "reason", reason)) + "\n"
	case "play":
		if hint.CardIndex == nil {
			return i18n.T("watten.hintNone") + "\n"
		}
		player := g.GetPlayer(0)
		card := player.GetCard(*hint.CardIndex)
		return color.Yellow(i18n.Tf("watten.hintCard",
			"idx", strconv.Itoa(*hint.CardIndex),
			"card", cuiCardStr(card),
			"reason", reason)) + "\n"
	default:
		return i18n.T("watten.hintNone") + "\n"
	}
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *WattenCuiPresenter) ActionLogOutput(g interfaces.WattenGame) string {
	return actionLogOutputText(g)
}
