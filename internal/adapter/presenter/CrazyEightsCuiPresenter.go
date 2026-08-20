package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// crazyEightsIsLegal reports whether a card may be played onto the current
// discard top (mirrors the domain's unexported isValidPlay): eights are always
// legal; once a suit is chosen only that suit matches; otherwise suit or rank.
func crazyEightsIsLegal(card, top *domain.Card, chosenSuit int) bool {
	if card.GetValue() == domain.CrazyEightsWildValue {
		return true
	}
	if top == nil {
		return true
	}
	if chosenSuit > 0 {
		return card.GetDesign() == chosenSuit
	}
	return card.GetDesign() == top.GetDesign() || card.GetValue() == top.GetValue()
}

// crazyEightsHandStr renders the human hand as an indexed list, appending "*"
// to each card that is legal to play right now.
func crazyEightsHandStr(player *domain.CrazyEightsPlayer, top *domain.Card, chosenSuit int) string {
	parts := make([]string, player.GetCardsSize())
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		s := "[" + strconv.Itoa(i) + "]" + cuiCardStr(c)
		if crazyEightsIsLegal(c, top, chosenSuit) {
			s += "*"
		}
		parts[i] = s
	}
	return strings.Join(parts, "  ")
}

// crazyEightsPlayerStr returns the display string for a single CrazyEights
// player. When markLegal is set (the human's play turn), legal cards are
// starred to spare the player from matching suit/rank by hand.
func crazyEightsPlayerStr(player *domain.CrazyEightsPlayer, i int, markLegal bool, top *domain.Card, chosenSuit int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("crazyeights.playerLine",
		"name", cuiPlayerName(player, i),
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize())) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		if markLegal {
			b.WriteString(crazyEightsHandStr(player, top, chosenSuit) + "\n")
		} else {
			b.WriteString(cuiIndexedCardListStr(player) + "\n")
		}
	}
	return b.String()
}

// CrazyEightsCuiPresenter renders the Crazy Eights CUI view.
type CrazyEightsCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *CrazyEightsCuiPresenter) Output(g interfaces.CrazyEightsGame, lastErr error) string {
	return buildCuiOutput(i18n.T("crazyeights.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("crazyeights.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"stock", strconv.Itoa(g.GetDrawPileCount())) + "\n")

		// Top of discard pile
		if top := g.GetDiscardTop(); top != nil {
			b.WriteString(i18n.Tf("crazyeights.discardLine", "card", cuiCardStr(top)))
			if g.GetChosenSuit() > 0 {
				b.WriteString(i18n.Tf("crazyeights.chosenSuit",
					"suit", suitDisplayName(g.GetChosenSuit())))
			}
			b.WriteString("\n")
		}

		phase := g.GetPhase()
		currentIdx := g.GetCurrentPlayerIdx()
		top := g.GetDiscardTop()
		chosenSuit := g.GetChosenSuit()
		for i := 0; i < g.GetPlayerCnt(); i++ {
			markLegal := phase == domain.CrazyEightsPhasePlay && i == currentIdx
			b.WriteString(crazyEightsPlayerStr(g.GetPlayer(i), i, markLegal, top, chosenSuit))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerIdx := g.GetWinnerIdx()
			banner := i18n.Tf("crazyeights.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch phase {
		case domain.CrazyEightsPhasePlay:
			b.WriteString(i18n.Tf("crazyeights.promptCurrentPlayer",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("crazyeights.promptPlayHelp") + "\n")
			b.WriteString(i18n.T("crazyeights.promptDrawHelp") + "\n")
		case domain.CrazyEightsPhaseChooseSuit:
			b.WriteString(i18n.T("crazyeights.promptChooseSuit") + "\n")
			b.WriteString(i18n.T("crazyeights.promptChooseSuitHelp") + "\n")
		case domain.CrazyEightsPhaseRoundEnd:
			b.WriteString(i18n.T("crazyeights.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("crazyeights.promptRoundEndHelp") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *CrazyEightsCuiPresenter) ActionLogOutput(g interfaces.CrazyEightsGame) string {
	return actionLogOutputTextForSeats[*domain.CrazyEightsPlayer](g)
}

// suitDisplayName returns the suit display string.
func suitDisplayName(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return "♠"
	case domain.CardDesignClover:
		return "♣"
	case domain.CardDesignHeart:
		return "♥"
	case domain.CardDesignDiamond:
		return "♦"
	default:
		return "?"
	}
}

// HintOutput emits the current Crazy Eights hint.
//
// **Hearts / Spades はサーバー計算の理由付きヒントを返すのに、CrazyEights には
// これが無く、全ゲーム共通の簡易ヒューリスティックしか支援が無かった (#4737)。**
func (p *CrazyEightsCuiPresenter) HintOutput(g interfaces.CrazyEightsGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("crazyeights.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, crazyEightsHintReasonKeys)
	if hint.Suit != nil {
		return color.Yellow(i18n.Tf("crazyeights.hintSuit",
			"suit", cuiSuitName(*hint.Suit),
			"reason", reason)) + "\n"
	}
	if hint.CardIndex == nil {
		return i18n.T("crazyeights.hintNone") + "\n"
	}
	card := g.GetPlayer(0).GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("crazyeights.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", reason)) + "\n"
}

// crazyEightsHintReasonKeys maps hint-reason identifiers to their i18n keys.
var crazyEightsHintReasonKeys = map[string]string{
	"match_suit":          "crazyeights.hintReasonMatchSuit",
	"match_rank":          "crazyeights.hintReasonMatchRank",
	"play_wild":           "crazyeights.hintReasonPlayWild",
	"play_valid":          "crazyeights.hintReasonPlayValid",
	"choose_longest_suit": "crazyeights.hintReasonChooseLongestSuit",
}
