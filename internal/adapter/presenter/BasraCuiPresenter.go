//go:build !js || !wasm || extra3

package presenter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// BasraCuiPresenter renders the Basra CUI view.
type BasraCuiPresenter struct{}

// basraTableStr は場札を "[0]♠5 [1]♥J" 形式で返す。
func basraTableStr(cards []*domain.Card) string {
	if len(cards) == 0 {
		return "-"
	}
	parts := make([]string, len(cards))
	for i, c := range cards {
		parts[i] = "[" + strconv.Itoa(i) + "]" + cuiCardStr(c)
	}
	return strings.Join(parts, " ")
}

func basraPlayerStr(g interfaces.BasraGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("basra.playerLine",
		"name", cuiPlayerName(player, idx),
		"hand", strconv.Itoa(player.GetCardsSize()),
		"captured", strconv.Itoa(player.CapturedCount()),
		"basra", strconv.Itoa(player.GetBasraCount())) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
		// **どの札で何を取れるかを見せる。**Web は選択中の札が捕獲できる場札を
		// リングとチェックで示すのに、CUI はヒントを叩かない限り分からなかった
		// (#4922)。判定はドメインの GetCaptureOptions をそのまま使う。
		if line := cuiCaptureHintLine(player, g.GetCaptureOptions(idx), "basra.captureHint"); line != "" {
			b.WriteString(line + "\n")
		}
	}
	return b.String()
}

// Output renders the current game state for the active locale.
func (p *BasraCuiPresenter) Output(g interfaces.BasraGame, lastErr error) string {
	return buildCuiOutput(i18n.T("basra.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("basra.deckLine",
			"deal", strconv.Itoa(g.GetRoundNumber()),
			"deck", strconv.Itoa(g.GetRemainingDeck())) + "\n")
		b.WriteString(i18n.Tf("basra.tableLine", "table", basraTableStr(g.GetTableCards())) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(basraPlayerStr(g, i))
		}
		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch g.GetPhase() {
		case domain.BasraPhasePlay:
			currentIdx := g.GetCurrentTurn()
			b.WriteString(i18n.Tf("basra.promptPlay",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
		case domain.BasraPhaseGameEnd:
			// 勝者も最終得点も CUI のどこにも出ていなかった (Web の basra-result と
			// 同じ情報源: GetWinners と各プレイヤーの GetScore)。
			winners := make([]string, 0, len(g.GetWinners()))
			for _, w := range g.GetWinners() {
				winners = append(winners, cuiPlayerName(g.GetPlayer(w), w))
			}
			if len(winners) > 0 {
				b.WriteString(i18n.Tf("basra.resultWinner",
					"names", strings.Join(winners, ", ")) + "\n")
			}
			scores := make([]string, 0, g.GetPlayerCnt())
			for i := 0; i < g.GetPlayerCnt(); i++ {
				scores = append(scores, fmt.Sprintf("%s %d",
					cuiPlayerName(g.GetPlayer(i), i), g.GetPlayer(i).GetScore()))
			}
			b.WriteString(i18n.Tf("basra.resultScores",
				"scores", strings.Join(scores, " / ")) + "\n")
			b.WriteString(i18n.T("basra.promptGameEnd") + "\n")
		}
		b.WriteString(i18n.T("basra.promptHelp") + "\n")
	})
}

// HintOutput emits the current Basra hint.
func (p *BasraCuiPresenter) HintOutput(g interfaces.BasraGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("basra.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, basraHintReasonKeys)
	if len(hint.CardIndices) > 0 {
		playerIdx := g.GetCurrentTurn()
		player := g.GetPlayer(playerIdx)
		cards := make([]string, len(hint.CardIndices))
		for i, idx := range hint.CardIndices {
			if player != nil && idx >= 0 && idx < player.GetCardsSize() {
				cards[i] = "[" + strconv.Itoa(idx) + "]" + cuiCardStr(player.GetCard(idx))
			} else {
				cards[i] = strconv.Itoa(idx)
			}
		}
		return color.Yellow(i18n.Tf("basra.hintCard",
			"cards", strings.Join(cards, ", "),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("basra.hintCard", "cards", "-", "reason", reason)) + "\n"
}

// basraHintReasonKeys maps Basra-specific hint-reason identifiers to i18n keys.
var basraHintReasonKeys = map[string]string{
	"basra_sweep": "basra.hintReasonBasra",
	"jack_sweep":  "basra.hintReasonJack",
	"capture":     "basra.hintReasonCapture",
	"trail_low":   "basra.hintReasonTrail",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *BasraCuiPresenter) ActionLogOutput(g interfaces.BasraGame) string {
	return actionLogOutputTextForSeats[*domain.BasraPlayer](g)
}
