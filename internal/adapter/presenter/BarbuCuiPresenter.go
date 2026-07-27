//go:build !js || !wasm || solo

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// BarbuCuiPresenter renders the Barbu CUI view.
type BarbuCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *BarbuCuiPresenter) Output(bg interfaces.BarbuGame, lastErr error) string {
	return buildCuiOutput(i18n.T("barbu.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("barbu.dealLine",
			"deal", strconv.Itoa(bg.GetDealNumber()+1),
			"total", strconv.Itoa(domain.BarbuTotalDeals),
			"dealer", cuiPlayerName(bg.GetPlayer(bg.GetDealerIdx()), bg.GetDealerIdx())) + "\n")

		for i := 0; i < bg.GetPlayerCnt(); i++ {
			b.WriteString(barbuPlayerStr(bg.GetPlayer(i), i))
		}
		b.WriteString("----------\n")

		if bg.GetCurrentContract() >= 0 {
			b.WriteString(i18n.Tf("barbu.contractLine",
				"contract", barbuContractLabel(bg.GetCurrentContract()),
				"trump", barbuTrumpLabel(bg.GetTrumpSuit())) + "\n")
		}

		barbuWriteBoard(b, bg)
		cuiErrorBlock(b, lastErr)

		if bg.GetGameEndFlag() {
			b.WriteString(i18n.T("barbu.gameEnd") + "\n")
			for i := 0; i < bg.GetPlayerCnt(); i++ {
				pl := bg.GetPlayer(i)
				if pl == nil {
					continue
				}
				b.WriteString(i18n.Tf("barbu.scoreEntry",
					"name", cuiPlayerName(pl, i),
					"score", strconv.Itoa(pl.GetTotalScore())) + "\n")
			}
			return
		}

		if bg.GetPhase() == domain.BarbuPhaseSelectContract {
			b.WriteString(i18n.Tf("barbu.selectPrompt",
				"name", cuiPlayerName(bg.GetPlayer(bg.GetDealerIdx()), bg.GetDealerIdx())) + "\n")
		} else {
			b.WriteString(i18n.Tf("barbu.promptCurrentTurn",
				"name", cuiPlayerName(bg.GetPlayer(bg.GetCurrentTurn()), bg.GetCurrentTurn())) + "\n")
		}
		b.WriteString(i18n.T("barbu.promptHelp") + "\n")
	})
}

// barbuWriteBoard renders the current trick (trick contracts) or the placed
// dominoes (Dominoes contract).
func barbuWriteBoard(b *strings.Builder, bg interfaces.BarbuGame) {
	if bg.GetCurrentContract() == domain.BarbuContractDominoes {
		table := bg.GetTablePlaced()
		any := false
		for suit := domain.CardDesignSpade; suit <= domain.CardDesignDiamond; suit++ {
			cards := barbuPlacedCards(suit, table[suit])
			if len(cards) == 0 {
				continue
			}
			any = true
			b.WriteString(i18n.Tf("barbu.dominoRow",
				"suit", barbuTrumpLabel(suit),
				"cards", cuiCardSliceStr(cards)) + "\n")
		}
		if !any {
			b.WriteString(i18n.T("barbu.tableEmpty") + "\n")
		}
		return
	}
	if trick := bg.GetCurrentTrick(); len(trick) > 0 {
		cards := make([]*domain.Card, 0, len(trick))
		for _, tc := range trick {
			cards = append(cards, tc.Card)
		}
		b.WriteString(i18n.Tf("barbu.trickLine", "cards", cuiCardSliceStr(cards)) + "\n")
	} else {
		b.WriteString(i18n.T("barbu.tableEmpty") + "\n")
	}
}

// barbuPlacedCards は 1 スートの bitmask から配置済みカードを復元する。
func barbuPlacedCards(suit int, mask uint16) []*domain.Card {
	cards := make([]*domain.Card, 0)
	for v := 1; v <= domain.CardValueMax; v++ {
		if mask&(uint16(1)<<uint(v)) != 0 {
			cards = append(cards, domain.NewCard(suit, v, false))
		}
	}
	return cards
}

// barbuPlayerStr returns the display string for a single Barbu player.
func barbuPlayerStr(player *domain.BarbuPlayer, i int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("barbu.playerLine",
		"name", cuiPlayerName(player, i),
		"hand", strconv.Itoa(player.GetCardsSize()),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"total", strconv.Itoa(player.GetTotalScore())) + "\n")
	if player.GetIsHuman() {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// barbuContractLabel はコントラクトの表示名を i18n で返す。
func barbuContractLabel(contract int) string {
	keys := []string{
		"barbu.cNoTricks", "barbu.cNoHearts", "barbu.cNoQueens", "barbu.cBarbu",
		"barbu.cNoLastTrick", "barbu.cTrumps", "barbu.cDominoes",
	}
	if contract < 0 || contract >= len(keys) {
		return "-"
	}
	return i18n.T(keys[contract])
}

// barbuTrumpLabel は切り札スートの表示名を返す (-1 = なし)。
func barbuTrumpLabel(suit int) string {
	if suit < domain.CardDesignSpade || suit > domain.CardDesignDiamond {
		return "-"
	}
	return suitNames[suit]
}

// barbuLegalTrickIndices returns the hand positions the player may legally play
// in a trick contract: cards of the lead suit, or the whole hand when leading
// or void in the lead suit.
func barbuLegalTrickIndices(player *domain.BarbuPlayer, trick []*domain.TrickCard) []int {
	cardsSize := player.GetCardsSize()
	makeAll := func() []int {
		all := make([]int, cardsSize)
		for i := range all {
			all[i] = i
		}
		return all
	}
	// Find the lead card defensively: a malformed/empty trick falls back to
	// "all cards legal" rather than dereferencing a nil entry.
	var leadCard *domain.Card
	for _, tc := range trick {
		if tc != nil && tc.Card != nil {
			leadCard = tc.Card
			break
		}
	}
	if leadCard == nil {
		return makeAll()
	}
	leadSuit := leadCard.GetDesign()
	follow := make([]int, 0, cardsSize)
	for i := 0; i < cardsSize; i++ {
		if c := player.GetCard(i); c != nil && c.GetDesign() == leadSuit {
			follow = append(follow, i)
		}
	}
	if len(follow) == 0 {
		return makeAll()
	}
	return follow
}

// HintOutput emits a play recommendation for the human's turn: the playable
// hand positions for the Dominoes contract (via GetDominoPlayableIndices) or
// the legal follow-suit plays for a trick contract.
func (p *BarbuCuiPresenter) HintOutput(bg interfaces.BarbuGame) string {
	if bg.GetPhase() != domain.BarbuPhasePlay {
		return i18n.T("barbu.hintNone") + "\n"
	}
	turn := bg.GetCurrentTurn()
	player := bg.GetPlayer(turn)
	if player == nil || !player.GetIsHuman() {
		return i18n.T("barbu.hintNone") + "\n"
	}
	idxStr := func(idxs []int) string {
		parts := make([]string, len(idxs))
		for i, idx := range idxs {
			parts[i] = "[" + strconv.Itoa(idx) + "]"
		}
		return strings.Join(parts, " ")
	}
	if bg.GetCurrentContract() == domain.BarbuContractDominoes {
		playable := bg.GetDominoPlayableIndices(turn)
		if len(playable) == 0 {
			return color.Yellow(i18n.T("barbu.hintDominoPass")) + "\n"
		}
		return color.Yellow(i18n.Tf("barbu.hintDomino", "cards", idxStr(playable))) + "\n"
	}
	legal := barbuLegalTrickIndices(player, bg.GetCurrentTrick())
	return color.Yellow(i18n.Tf("barbu.hintLegal", "cards", idxStr(legal))) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *BarbuCuiPresenter) ActionLogOutput(bg interfaces.BarbuGame) string {
	return actionLogOutputText(bg)
}
