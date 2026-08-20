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

// RookCuiPresenter renders the Rook (ルーク) CUI view.
type RookCuiPresenter struct{}

// rookCuiHintReasonKeys maps Rook-specific hint reasons to i18n keys.
var rookCuiHintReasonKeys = map[string]string{
	"pass_recommended": "rook.hintReasonPass",
	"strategic_bid":    "rook.hintReasonBid",
	"discard_weakest":  "rook.hintReasonDiscardWeakest",
	"lead_trump":       "rook.hintReasonLeadTrump",
	"lead_strong":      "rook.hintReasonLeadStrong",
	"follow_suit":      "rook.hintReasonFollowSuit",
	"trump_cut":        "rook.hintReasonTrumpCut",
	"discard_weak":     "rook.hintReasonDiscardWeak",
}

// rookColorNameI18n は色番号をローカライズされた色名にする。
func rookColorNameI18n(color int) string {
	switch color {
	case 1:
		return i18n.T("rook.colorRed")
	case 2:
		return i18n.T("rook.colorYellow")
	case 3:
		return i18n.T("rook.colorGreen")
	case 4:
		return i18n.T("rook.colorBlack")
	}
	return "-"
}

// rookCuiCardStr は Rook の札を "赤 7" / "Rook" のように描画する (フレンチスート不使用)。
func rookCuiCardStr(card *domain.Card) string {
	if card == nil {
		return "??"
	}
	if card.GetDesign() == domain.RookBirdDesign {
		return i18n.T("rook.birdName")
	}
	return rookColorNameI18n(card.GetDesign()) + " " + strconv.Itoa(card.GetValue())
}

// rookCuiHandStr は手札をインデックス付きで描画する。
func rookCuiHandStr(player *domain.RookPlayer) string {
	parts := make([]string, player.GetCardsSize())
	for i := range parts {
		parts[i] = "[" + strconv.Itoa(i) + "]" + rookCuiCardStr(player.GetCard(i))
	}
	return strings.Join(parts, "  ")
}

// rookPlayerStr returns the display string for a single player.
func rookPlayerStr(player *domain.RookPlayer, i int) string {
	var b strings.Builder
	status := ""
	if player.GetIsDeclarer() {
		status = " " + i18n.T("rook.declarerMark")
	} else if player.GetPassed() {
		status = " " + i18n.T("rook.passMark")
	} else if bid := player.GetBid(); bid > 0 {
		status = " [" + strconv.Itoa(bid) + "]"
	}
	b.WriteString(i18n.Tf("rook.playerLine",
		"name", cuiPlayerName(player, i),
		"team", strconv.Itoa(player.GetTeam()),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"points", strconv.Itoa(player.GetPoints()),
		"cards", strconv.Itoa(player.GetCardsSize()),
		"status", status,
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(rookCuiHandStr(player) + "\n")
	}
	return b.String()
}

// Output renders the current game state for the active locale.
func (p *RookCuiPresenter) Output(g interfaces.RookGame, lastErr error) string {
	return buildCuiOutput(i18n.T("rook.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("rook.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber())) + "\n")
		dealerIdx := g.GetDealerIdx()
		b.WriteString(i18n.Tf("rook.dealer",
			"name", cuiPlayerName(g.GetPlayer(dealerIdx), dealerIdx)) + "\n")

		if g.GetContractBid() > 0 {
			declIdx := g.GetDeclarerIdx()
			b.WriteString(i18n.Tf("rook.contractLine",
				"name", cuiPlayerName(g.GetPlayer(declIdx), declIdx),
				"bid", strconv.Itoa(g.GetContractBid()),
				"trump", rookColorNameI18n(g.GetTrumpColor())) + "\n")
		} else if hb := g.GetHighestBid(); hb > 0 {
			b.WriteString(i18n.Tf("rook.highestBid", "bid", strconv.Itoa(hb)) + "\n")
		} else {
			b.WriteString(i18n.T("rook.contractUndecided") + "\n")
		}

		b.WriteString(i18n.Tf("rook.teamScoreLine",
			"t0", strconv.Itoa(g.GetTeamScore(0)),
			"t1", strconv.Itoa(g.GetTeamScore(1))) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(rookPlayerStr(g.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		cuiTrickBlock(b, g.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return rookCuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			banner := i18n.Tf("rook.gameEnd", "team", strconv.Itoa(g.GetWinnerTeam()))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.RookPhaseBid:
			bidIdx := g.GetBidPlayerIdx()
			b.WriteString(i18n.Tf("rook.promptBid",
				"name", cuiPlayerName(g.GetPlayer(bidIdx), bidIdx)) + "\n")
			b.WriteString(i18n.T("rook.promptBidHelp") + "\n")
		case domain.RookPhaseNestExchange:
			b.WriteString(rookCuiNestStr(g) + "\n")
			b.WriteString(i18n.T("rook.promptNestExchange") + "\n")
			b.WriteString(i18n.T("rook.promptNestExchangeHelp") + "\n")
		case domain.RookPhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("rook.promptCurrentPlayer",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			// **追随は強制。**リードスートも出せる札も示さないと、出して拒否
			// されるまで違反に気づけない (#4928)。
			if g.IsHumanTurn() {
				if idx := g.GetPlayableIndices(currentIdx); len(idx) > 0 {
					parts := make([]string, len(idx))
					for i, v := range idx {
						parts[i] = strconv.Itoa(v)
					}
					// **義務の断りは実際に縛られているときだけ。**リード時や
					// ボイド時は全札出せるので、そこで「従う義務」と言うと嘘になる。
					key := "rook.playable"
					if p := g.GetPlayer(currentIdx); p != nil && len(idx) < p.GetCardsSize() {
						key = "rook.playableRestricted"
					}
					b.WriteString(i18n.Tf(key, "indexes", strings.Join(parts, " ")) + "\n")
				}
			}
			b.WriteString(i18n.T("rook.promptPlayHelp") + "\n")
		case domain.RookPhaseTrickEnd:
			b.WriteString(i18n.T("rook.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("rook.promptTrickEndHelp") + "\n")
		case domain.RookPhaseRoundEnd:
			b.WriteString(i18n.T("rook.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("rook.promptRoundEndHelp") + "\n")
		}
	})
}

// rookCuiNestStr は落札者(人間)へネストの内容を表示する。
func rookCuiNestStr(g interfaces.RookGame) string {
	declIdx := g.GetDeclarerIdx()
	if declIdx < 0 || g.GetPlayer(declIdx) == nil || !g.GetPlayer(declIdx).GetIsHuman() {
		return ""
	}
	nest := g.GetNest()
	parts := make([]string, len(nest))
	for i, c := range nest {
		parts[i] = rookCuiCardStr(c)
	}
	return i18n.Tf("rook.nestLine", "cards", strings.Join(parts, "  "))
}

// HintOutput emits the current hint.
func (p *RookCuiPresenter) HintOutput(g interfaces.RookGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("rook.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, rookCuiHintReasonKeys)
	switch {
	case hint.Pass != nil && *hint.Pass:
		return color.Yellow(i18n.Tf("rook.hintPass", "reason", reason)) + "\n"
	case hint.Bid != nil:
		return color.Yellow(i18n.Tf("rook.hintBid",
			"bid", strconv.Itoa(*hint.Bid), "reason", reason)) + "\n"
	case len(hint.DiscardIndices) > 0:
		trump := ""
		if hint.TrumpColor != nil {
			trump = rookColorNameI18n(*hint.TrumpColor)
		}
		return color.Yellow(i18n.Tf("rook.hintDiscard",
			"indices", rookJoinInts(hint.DiscardIndices), "trump", trump, "reason", reason)) + "\n"
	case hint.CardIndex != nil:
		card := g.GetPlayer(0).GetCard(*hint.CardIndex)
		return color.Yellow(i18n.Tf("rook.hintCard",
			"idx", strconv.Itoa(*hint.CardIndex),
			"card", rookCuiCardStr(card),
			"reason", reason)) + "\n"
	}
	return i18n.T("rook.hintNone") + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *RookCuiPresenter) ActionLogOutput(g interfaces.RookGame) string {
	return actionLogOutputTextWithNames(g, func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) })
}

// rookJoinInts formats an int slice as a space-separated string.
func rookJoinInts(xs []int) string {
	parts := make([]string, len(xs))
	for i, x := range xs {
		parts[i] = strconv.Itoa(x)
	}
	return strings.Join(parts, " ")
}
