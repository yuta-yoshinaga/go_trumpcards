package presenter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// binokelMeldTypeKey maps a meld type constant to the i18n key holding
// its localized name. The lookup indirection lets the displayed string
// follow the active locale.
var binokelMeldTypeKey = map[domain.BinokelMeldType]string{
	domain.BinokelMeldDix:                "binokel.meldDix",
	domain.BinokelMeldCommonMarriage:     "binokel.meldCommonMarriage",
	domain.BinokelMeldRoyalMarriage:      "binokel.meldRoyalMarriage",
	domain.BinokelMeldBinokel:            "binokel.meldBinokel",
	domain.BinokelMeldJacksAround:        "binokel.meldJacksAround",
	domain.BinokelMeldQueensAround:       "binokel.meldQueensAround",
	domain.BinokelMeldKingsAround:        "binokel.meldKingsAround",
	domain.BinokelMeldAcesAround:         "binokel.meldAcesAround",
	domain.BinokelMeldNonTrumpRun:        "binokel.meldNonTrumpRun",
	domain.BinokelMeldRun:                "binokel.meldRun",
	domain.BinokelMeldRundgang:           "binokel.meldRundgang",
	domain.BinokelMeldDoubleBinokel:      "binokel.meldDoubleBinokel",
	domain.BinokelMeldDoubleJacksAround:  "binokel.meldDoubleJacksAround",
	domain.BinokelMeldDoubleQueensAround: "binokel.meldDoubleQueensAround",
	domain.BinokelMeldDoubleKingsAround:  "binokel.meldDoubleKingsAround",
	domain.BinokelMeldDoubleAcesAround:   "binokel.meldDoubleAcesAround",
	domain.BinokelMeldDoubleNonTrumpRun:  "binokel.meldDoubleNonTrumpRun",
	domain.BinokelMeldDoubleRun:          "binokel.meldDoubleRun",
}

// binokelMeldName returns the localized meld name for the given type, or
// the type's numeric form when the key is missing (defensive fallback for
// future meld types added to the domain layer before the locale catches up).
func binokelMeldName(t domain.BinokelMeldType) string {
	if key, ok := binokelMeldTypeKey[t]; ok {
		return i18n.T(key)
	}
	return fmt.Sprintf("meld#%d", int(t))
}

// binokelMeldTablePerLine は早見表の1行に並べるメルドの数。
const binokelMeldTablePerLine = 5

// binokelMeldTableStr renders the meld types and their points, cheapest
// first, wrapped a few per line.
//
// **点数は domain.BinokelMeldTable() から引く。**ここに書き写すと、加点の値を
// 直したときに表だけが古いまま残る (#5519)。
func binokelMeldTableStr() string {
	var b strings.Builder
	b.WriteString(i18n.T("binokel.meldTableHeader"))
	for i, e := range domain.BinokelMeldTable() {
		switch {
		case i == 0:
		case i%binokelMeldTablePerLine == 0:
			b.WriteString("\n  ")
		default:
			b.WriteString(" / ")
		}
		b.WriteString(i18n.Tf("binokel.meldTableEntry",
			"type", binokelMeldName(e.Type),
			"points", strconv.Itoa(e.Points)))
	}
	b.WriteString("\n")
	return b.String()
}

// binokelBidStr returns the localized bid string for a player:
// - "pass" / "パス" if the player has passed
// - "no bid" / "未ビッド" if the player has not made a bid yet (bid <= 0)
// - numeric amount if the player declared a bid
func binokelBidStr(player *domain.BinokelPlayer) string {
	if player == nil {
		return ""
	}
	if player.GetHasPassed() {
		return i18n.T("binokel.bidPass")
	}
	if player.GetBid() <= 0 {
		return i18n.T("binokel.bidPending")
	}
	return strconv.Itoa(player.GetBid())
}

// binokelPlayerStr returns the display string for a single Binokel player.
// legalIndices, when non-nil, lists the hand positions the human may legally
// play this turn and is rendered as a follow-rule legend below their hand.
func binokelPlayerStr(player *domain.BinokelPlayer, i int, score int, legalIndices []int) string {
	var b strings.Builder
	name := cuiPlayerName(player, i)
	b.WriteString(i18n.Tf("binokel.playerLine",
		"name", name,
		"score", strconv.Itoa(score),
		"bid", binokelBidStr(player),
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
			b.WriteString(i18n.Tf("binokel.legalPlayLegend",
				"indices", strings.Join(parts, " ")) + "\n")
		}
	}
	return b.String()
}

// BinokelCuiPresenter renders the Binokel CUI view.
type BinokelCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *BinokelCuiPresenter) Output(g interfaces.BinokelGame, lastErr error) string {
	return buildCuiOutput(i18n.T("binokel.helpTitle"), func(b *strings.Builder) {
		fmt.Fprintln(b, i18n.Tf("binokel.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber())))
		fmt.Fprintln(b, i18n.Tf("binokel.dealer",
			"name", cuiPlayerName(g.GetPlayer(g.GetDealerIdx()), g.GetDealerIdx())))

		trumpSuit := g.GetTrumpSuit()
		switch {
		case trumpSuit > 0 && g.GetHighestBidder() >= 0:
			fmt.Fprintln(b, i18n.Tf("binokel.trumpWithBidder",
				"suit", cuiSuitName(trumpSuit),
				"bidder", cuiPlayerName(g.GetPlayer(g.GetHighestBidder()), g.GetHighestBidder())))
		case trumpSuit > 0:
			fmt.Fprintln(b, i18n.Tf("binokel.trumpOnly", "suit", cuiSuitName(trumpSuit)))
		default:
			fmt.Fprintln(b, i18n.T("binokel.trumpUndecided"))
		}

		if g.GetHighestBidder() >= 0 && g.GetHighestBid() > 0 {
			fmt.Fprintln(b, i18n.Tf("binokel.highestBidWithName",
				"bid", strconv.Itoa(g.GetHighestBid()),
				"name", cuiPlayerName(g.GetPlayer(g.GetHighestBidder()), g.GetHighestBidder())))
		} else if g.GetHighestBid() > 0 {
			fmt.Fprintln(b, i18n.Tf("binokel.highestBid", "bid", strconv.Itoa(g.GetHighestBid())))
		} else {
			fmt.Fprintln(b, i18n.Tf("binokel.highestBid", "bid", i18n.T("binokel.highestBidNone")))
		}

		// Player scores
		var scoreParts []string
		for i := 0; i < g.GetPlayerCnt(); i++ {
			scoreParts = append(scoreParts, fmt.Sprintf("%s=%d", cuiPlayerName(g.GetPlayer(i), i), g.GetScore(i)))
		}
		fmt.Fprintln(b, i18n.Tf("binokel.scores", "scores", strings.Join(scoreParts, "  ")))

		// Dabb cards display in Dabb phase
		if g.GetPhase() == domain.BinokelPhaseDabb {
			dabbCards := g.GetDabb()
			var cardStrs []string
			for _, c := range dabbCards {
				cardStrs = append(cardStrs, cuiCardStr(c))
			}
			fmt.Fprintln(b, i18n.Tf("binokel.dabbCards", "cards", strings.Join(cardStrs, " ")))
		}

		// Player rows. On the human's play turn, surface the legal follow plays.
		for i := 0; i < g.GetPlayerCnt(); i++ {
			var legalIndices []int
			if g.GetPhase() == domain.BinokelPhasePlay &&
				i == g.GetCurrentPlayerIdx() && g.GetPlayer(i).GetIsHuman() {
				legalIndices = g.GetValidPlayIndices(i)
			}
			b.WriteString(binokelPlayerStr(g.GetPlayer(i), i, g.GetScore(i), legalIndices))
		}

		// Meld details (only during meld + round-end phases)
		phase := g.GetPhase()
		if phase == domain.BinokelPhaseMeld || phase == domain.BinokelPhaseRoundEnd {
			melds := g.GetPlayerMelds()
			for i := range domain.BinokelPlayerCnt {
				if len(melds[i]) > 0 {
					fmt.Fprintln(b, i18n.Tf("binokel.playerMeldsHeader",
						"name", cuiPlayerName(g.GetPlayer(i), i)))
					for _, m := range melds[i] {
						fmt.Fprintln(b, i18n.Tf("binokel.playerMeldLine",
							"type", binokelMeldName(m.Type),
							"points", strconv.Itoa(m.Points)))
					}
				}
			}
		}

		// メルド早見表: ビッド〜メルド確認の間だけ。
		if phase == domain.BinokelPhaseBid || phase == domain.BinokelPhaseDabb || phase == domain.BinokelPhaseTrump || phase == domain.BinokelPhaseMeld {
			b.WriteString(binokelMeldTableStr())
		}

		// Current trick
		trick := g.GetCurrentTrick()
		if len(trick) > 0 {
			b.WriteString(i18n.T("binokel.tableLabel"))
			for j, tc := range trick {
				if j > 0 {
					b.WriteString(", ")
				}
				b.WriteString(i18n.Tf("binokel.tableCard",
					"name", cuiPlayerName(g.GetPlayer(tc.PlayerIdx), tc.PlayerIdx),
					"card", cuiCardStr(tc.Card)))
			}
			b.WriteString("\n")
		}

		p.buildCuiMessage(b, g, lastErr)
	})
}

// HintOutput emits the current Binokel hint.
func (p *BinokelCuiPresenter) HintOutput(g interfaces.BinokelGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("binokel.hintNone")
	}
	var parts []string
	if hint.BidAmount != nil {
		parts = append(parts, i18n.Tf("binokel.hintBid", "n", strconv.Itoa(*hint.BidAmount)))
	}
	if hint.Pass != nil && *hint.Pass {
		parts = append(parts, i18n.T("binokel.hintPass"))
	}
	if hint.Suit != nil {
		parts = append(parts, i18n.Tf("binokel.hintSuit", "suit", cuiSuitName(*hint.Suit)))
	}
	if hint.CardIndex != nil {
		parts = append(parts, i18n.Tf("binokel.hintCard", "idx", strconv.Itoa(*hint.CardIndex)))
	}
	return i18n.T("binokel.hintPrefix") + strings.Join(parts, " ")
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *BinokelCuiPresenter) ActionLogOutput(g interfaces.BinokelGame) string {
	return actionLogOutputTextForSeats[*domain.BinokelPlayer](g)
}

// buildCuiMessage writes the per-phase prompt or end-of-game banner.
func (p *BinokelCuiPresenter) buildCuiMessage(b *strings.Builder, g interfaces.BinokelGame, lastErr error) {
	if lastErr != nil {
		fmt.Fprintln(b, i18n.Tf("binokel.errorPrefix", "err", lastErr.Error()))
		return
	}
	if g.GetGameEndFlag() {
		winner := g.GetWinnerPlayer()
		fmt.Fprintln(b, i18n.Tf("binokel.gameEndPlayerWin", "name", cuiPlayerName(g.GetPlayer(winner), winner)))
		return
	}
	switch g.GetPhase() {
	case domain.BinokelPhaseBid:
		fmt.Fprintln(b, i18n.Tf("binokel.promptBid",
			"name", cuiPlayerName(g.GetPlayer(g.GetBidPlayerIdx()), g.GetBidPlayerIdx())))
	case domain.BinokelPhaseDabb:
		fmt.Fprintln(b, i18n.Tf("binokel.promptDabb",
			"name", cuiPlayerName(g.GetPlayer(g.GetCurrentPlayerIdx()), g.GetCurrentPlayerIdx())))
	case domain.BinokelPhaseTrump:
		fmt.Fprintln(b, i18n.Tf("binokel.promptTrump",
			"name", cuiPlayerName(g.GetPlayer(g.GetCurrentPlayerIdx()), g.GetCurrentPlayerIdx())))
	case domain.BinokelPhaseMeld:
		fmt.Fprintln(b, i18n.T("binokel.promptMeld"))
	case domain.BinokelPhasePlay:
		fmt.Fprintln(b, i18n.Tf("binokel.promptPlay",
			"name", cuiPlayerName(g.GetPlayer(g.GetCurrentPlayerIdx()), g.GetCurrentPlayerIdx())))
	case domain.BinokelPhaseTrickEnd:
		fmt.Fprintln(b, i18n.T("binokel.promptTrickEnd"))
	case domain.BinokelPhaseRoundEnd:
		fmt.Fprintln(b, i18n.T("binokel.promptRoundEnd"))
	}
}
