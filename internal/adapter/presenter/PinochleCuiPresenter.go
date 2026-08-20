package presenter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// pinochleMeldTypeKey maps a meld type constant to the i18n key holding
// its localized name. The lookup indirection lets the displayed string
// follow the active locale.
var pinochleMeldTypeKey = map[domain.PinochleMeldType]string{
	domain.PinochleMeldDix:                "pinochle.meldDix",
	domain.PinochleMeldCommonMarriage:     "pinochle.meldCommonMarriage",
	domain.PinochleMeldRoyalMarriage:      "pinochle.meldRoyalMarriage",
	domain.PinochleMeldPinochle:           "pinochle.meldPinochle",
	domain.PinochleMeldJacksAround:        "pinochle.meldJacksAround",
	domain.PinochleMeldQueensAround:       "pinochle.meldQueensAround",
	domain.PinochleMeldKingsAround:        "pinochle.meldKingsAround",
	domain.PinochleMeldAcesAround:         "pinochle.meldAcesAround",
	domain.PinochleMeldRun:                "pinochle.meldRun",
	domain.PinochleMeldDoublePinochle:     "pinochle.meldDoublePinochle",
	domain.PinochleMeldDoubleJacksAround:  "pinochle.meldDoubleJacksAround",
	domain.PinochleMeldDoubleQueensAround: "pinochle.meldDoubleQueensAround",
	domain.PinochleMeldDoubleKingsAround:  "pinochle.meldDoubleKingsAround",
	domain.PinochleMeldDoubleAcesAround:   "pinochle.meldDoubleAcesAround",
	domain.PinochleMeldDoubleRun:          "pinochle.meldDoubleRun",
}

// pinochleMeldName returns the localized meld name for the given type, or
// the type's numeric form when the key is missing (defensive fallback for
// future meld types added to the domain layer before the locale catches up).
func pinochleMeldName(t domain.PinochleMeldType) string {
	if key, ok := pinochleMeldTypeKey[t]; ok {
		return i18n.T(key)
	}
	return fmt.Sprintf("meld#%d", int(t))
}

// pinochleMeldTablePerLine は早見表の1行に並べるメルドの数。
const pinochleMeldTablePerLine = 5

// pinochleMeldTableStr renders the 15 meld types and their points, cheapest
// first, wrapped a few per line.
//
// **点数は domain.PinochleMeldTable() から引く。**ここに書き写すと、加点の値を
// 直したときに表だけが古いまま残る (#5519)。
func pinochleMeldTableStr() string {
	var b strings.Builder
	b.WriteString(i18n.T("pinochle.meldTableHeader"))
	for i, e := range domain.PinochleMeldTable() {
		switch {
		case i == 0:
		case i%pinochleMeldTablePerLine == 0:
			b.WriteString("\n  ")
		default:
			b.WriteString(" / ")
		}
		b.WriteString(i18n.Tf("pinochle.meldTableEntry",
			"type", pinochleMeldName(e.Type),
			"points", strconv.Itoa(e.Points)))
	}
	b.WriteString("\n")
	return b.String()
}

// pinochlePlayerStr returns the display string for a single Pinochle player.
// legalIndices, when non-nil, lists the hand positions the human may legally
// play this turn and is rendered as a follow-rule legend below their hand.
func pinochlePlayerStr(player *domain.PinochlePlayer, i int, legalIndices []int) string {
	var b strings.Builder
	name := cuiPlayerName(player, i)
	b.WriteString(i18n.Tf("pinochle.playerLine",
		"name", name,
		"team", strconv.Itoa(player.GetTeam()),
		"bid", strconv.Itoa(player.GetBid()),
		"meldPts", strconv.Itoa(player.GetMeldScore()),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"trickPts", strconv.Itoa(player.GetTrickPoints()),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player))
		b.WriteString("\n")
		if len(legalIndices) > 0 {
			parts := make([]string, len(legalIndices))
			for k, idx := range legalIndices {
				parts[k] = "[" + strconv.Itoa(idx) + "]"
			}
			b.WriteString(i18n.Tf("pinochle.legalPlayLegend",
				"indices", strings.Join(parts, " ")) + "\n")
		}
	}
	return b.String()
}

// PinochleCuiPresenter renders the Pinochle CUI view.
type PinochleCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *PinochleCuiPresenter) Output(g interfaces.PinochleGame, lastErr error) string {
	return buildCuiOutput(i18n.T("pinochle.helpTitle"), func(b *strings.Builder) {
		fmt.Fprintln(b, i18n.Tf("pinochle.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber())))
		fmt.Fprintln(b, i18n.Tf("pinochle.dealer",
			"name", cuiPlayerName(g.GetPlayer(g.GetDealerIdx()), g.GetDealerIdx())))

		trumpSuit := g.GetTrumpSuit()
		switch {
		case trumpSuit > 0 && g.GetHighestBidder() >= 0:
			fmt.Fprintln(b, i18n.Tf("pinochle.trumpWithBidder",
				"suit", cuiSuitName(trumpSuit),
				"team", strconv.Itoa(g.GetPlayer(g.GetHighestBidder()).GetTeam())))
		case trumpSuit > 0:
			fmt.Fprintln(b, i18n.Tf("pinochle.trumpOnly", "suit", cuiSuitName(trumpSuit)))
		default:
			fmt.Fprintln(b, i18n.T("pinochle.trumpUndecided"))
		}

		if g.GetHighestBidder() >= 0 {
			fmt.Fprintln(b, i18n.Tf("pinochle.highestBidWithName",
				"bid", strconv.Itoa(g.GetHighestBid()),
				"name", cuiPlayerName(g.GetPlayer(g.GetHighestBidder()), g.GetHighestBidder())))
		} else {
			fmt.Fprintln(b, i18n.Tf("pinochle.highestBid", "bid", strconv.Itoa(g.GetHighestBid())))
		}

		// Team scores
		fmt.Fprintln(b, i18n.Tf("pinochle.teamScores",
			"a", strconv.Itoa(g.GetTeamScore(0)),
			"b", strconv.Itoa(g.GetTeamScore(1))))

		// Player rows. On the human's play turn, surface the legal follow plays.
		for i := 0; i < g.GetPlayerCnt(); i++ {
			var legalIndices []int
			if g.GetPhase() == domain.PinochlePhasePlay &&
				i == g.GetCurrentPlayerIdx() && g.GetPlayer(i).GetIsHuman() {
				legalIndices = g.GetValidPlayIndices(i)
			}
			b.WriteString(pinochlePlayerStr(g.GetPlayer(i), i, legalIndices))
		}

		// Meld details (only during meld + round-end phases)
		phase := g.GetPhase()
		if phase == domain.PinochlePhaseMeld || phase == domain.PinochlePhaseRoundEnd {
			melds := g.GetPlayerMelds()
			for i := range domain.PinochlePlayerCnt {
				if len(melds[i]) > 0 {
					fmt.Fprintln(b, i18n.Tf("pinochle.playerMeldsHeader",
						"name", cuiPlayerName(g.GetPlayer(i), i)))
					for _, m := range melds[i] {
						fmt.Fprintln(b, i18n.Tf("pinochle.playerMeldLine",
							"type", pinochleMeldName(m.Type),
							"points", strconv.Itoa(m.Points)))
					}
				}
			}
		}

		// メルド早見表: ビッド〜メルド確認の間だけ。ビッド額を決めるときに
		// 「自分の手にいくら分の目があるか」を見る先がどこにも無かった (#5519)。
		// プレイ中は毎トリック流れて盤面が読めなくなるので出さない。
		if phase == domain.PinochlePhaseBid || phase == domain.PinochlePhaseTrump || phase == domain.PinochlePhaseMeld {
			b.WriteString(pinochleMeldTableStr())
		}

		// Current trick
		trick := g.GetCurrentTrick()
		if len(trick) > 0 {
			b.WriteString(i18n.T("pinochle.tableLabel"))
			for j, tc := range trick {
				if j > 0 {
					b.WriteString(", ")
				}
				b.WriteString(i18n.Tf("pinochle.tableCard",
					"name", cuiPlayerName(g.GetPlayer(tc.PlayerIdx), tc.PlayerIdx),
					"card", cuiCardStr(tc.Card)))
			}
			b.WriteString("\n")
		}

		p.buildCuiMessage(b, g, lastErr)
	})
}

// HintOutput emits the current Pinochle hint. Multiple hint fields are
// joined with a single space so future hints that combine, e.g.,
// BidAmount + Pass don't render as run-on text. Today only one field is
// populated at a time, but the join is cheap insurance against the next
// PinochleHint variant.
func (p *PinochleCuiPresenter) HintOutput(g interfaces.PinochleGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("pinochle.hintNone")
	}
	var parts []string
	if hint.BidAmount != nil {
		parts = append(parts, i18n.Tf("pinochle.hintBid", "n", strconv.Itoa(*hint.BidAmount)))
	}
	if hint.Pass != nil && *hint.Pass {
		parts = append(parts, i18n.T("pinochle.hintPass"))
	}
	if hint.Suit != nil {
		parts = append(parts, i18n.Tf("pinochle.hintSuit", "suit", cuiSuitName(*hint.Suit)))
	}
	if hint.CardIndex != nil {
		parts = append(parts, i18n.Tf("pinochle.hintCard", "idx", strconv.Itoa(*hint.CardIndex)))
	}
	return i18n.T("pinochle.hintPrefix") + strings.Join(parts, " ")
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *PinochleCuiPresenter) ActionLogOutput(g interfaces.PinochleGame) string {
	return actionLogOutputTextForSeats[*domain.PinochlePlayer](g)
}

// buildCuiMessage writes the per-phase prompt or end-of-game banner.
func (p *PinochleCuiPresenter) buildCuiMessage(b *strings.Builder, g interfaces.PinochleGame, lastErr error) {
	if lastErr != nil {
		fmt.Fprintln(b, i18n.Tf("pinochle.errorPrefix", "err", lastErr.Error()))
		return
	}
	if g.GetGameEndFlag() {
		fmt.Fprintln(b, i18n.Tf("pinochle.gameEndTeamWin", "team", strconv.Itoa(g.GetWinnerTeam())))
		return
	}
	switch g.GetPhase() {
	case domain.PinochlePhaseBid:
		fmt.Fprintln(b, i18n.Tf("pinochle.promptBid",
			"name", cuiPlayerName(g.GetPlayer(g.GetBidPlayerIdx()), g.GetBidPlayerIdx())))
	case domain.PinochlePhaseTrump:
		fmt.Fprintln(b, i18n.Tf("pinochle.promptTrump",
			"name", cuiPlayerName(g.GetPlayer(g.GetCurrentPlayerIdx()), g.GetCurrentPlayerIdx())))
	case domain.PinochlePhaseMeld:
		fmt.Fprintln(b, i18n.T("pinochle.promptMeld"))
	case domain.PinochlePhasePlay:
		fmt.Fprintln(b, i18n.Tf("pinochle.promptPlay",
			"name", cuiPlayerName(g.GetPlayer(g.GetCurrentPlayerIdx()), g.GetCurrentPlayerIdx())))
	case domain.PinochlePhaseTrickEnd:
		fmt.Fprintln(b, i18n.T("pinochle.promptTrickEnd"))
	case domain.PinochlePhaseRoundEnd:
		fmt.Fprintln(b, i18n.T("pinochle.promptRoundEnd"))
	}
}
