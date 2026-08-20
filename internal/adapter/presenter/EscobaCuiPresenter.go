//go:build !js || !wasm || classic

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// EscobaCuiPresenter renders the Escoba CUI view.
type EscobaCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *EscobaCuiPresenter) Output(eg interfaces.EscobaGame, lastErr error) string {
	return buildCuiOutput(i18n.T("escoba.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("escoba.roundLine",
			"round", strconv.Itoa(eg.GetRoundNumber()),
			"stock", strconv.Itoa(eg.GetStockRemaining())) + "\n")

		for i := 0; i < eg.GetPlayerCnt(); i++ {
			b.WriteString(escobaPlayerStr(eg.GetPlayer(i), i))
		}
		b.WriteString("----------\n")

		if tableCards := eg.GetTableCards(); len(tableCards) > 0 {
			b.WriteString(i18n.Tf("escoba.tableLine", "cards", cuiCardSliceStr(tableCards)) + "\n")
		} else {
			b.WriteString(i18n.T("escoba.tableEmpty") + "\n")
		}

		cuiErrorBlock(b, lastErr)

		if eg.GetGameEndFlag() {
			b.WriteString(i18n.T("escoba.gameEnd") + "\n")
			b.WriteString(i18n.Tf("escoba.winnerLine",
				"player", strconv.Itoa(eg.GetWinnerIdx())) + "\n")
			escobaScoreDetailStr(b, eg.GetLastRoundDetail())
			return
		}

		if eg.GetPhase() == domain.EscobaPhaseRoundEnd {
			b.WriteString(i18n.T("escoba.roundEnd") + "\n")
			escobaScoreDetailStr(b, eg.GetLastRoundDetail())
			b.WriteString(i18n.T("escoba.promptNext") + "\n")
			return
		}

		currentTurn := eg.GetCurrentTurn()
		b.WriteString(i18n.Tf("escoba.promptCurrentTurn",
			"name", cuiPlayerName(eg.GetPlayer(currentTurn), currentTurn)) + "\n")
		b.WriteString(i18n.T("escoba.promptHelp") + "\n")
	})
}

// escobaPlayerStr returns the display string for a single Escoba player.
func escobaPlayerStr(player *domain.ScopaPlayer, i int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("escoba.playerLine",
		"name", cuiPlayerName(player, i),
		"hand", strconv.Itoa(player.GetCardsSize()),
		"captured", strconv.Itoa(player.CapturedCount()),
		"escoba", strconv.Itoa(player.GetScopaCount()),
		"score", strconv.Itoa(player.GetTotalScore())) + "\n")
	if player.GetIsHuman() {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
		// **「7 は取れているか」「espadas は何枚か」は得点計算に直結する** (#5662)。
		// Web は captured-viewer で実カードをいつでも開けるのに、CUI は枚数の
		// 数字しか出しておらず、そこから読み取る術が無かった。CPU の分は
		// 手札と同じく伏せたままにする。
		if captured := player.GetCapturedCards(); len(captured) > 0 {
			b.WriteString(i18n.Tf("escoba.capturedLine", "cards", cuiCardSliceStr(captured)) + "\n")
		}
	}
	return b.String()
}

// escobaScoreDetailStr renders the per-player score breakdown plus the
// Ace/Seven of swords holders.
func escobaScoreDetailStr(b *strings.Builder, det *domain.EscobaScoreDetail) {
	if det == nil {
		return
	}
	for i := 0; i < domain.EscobaPlayerCnt; i++ {
		b.WriteString(i18n.Tf("escoba.scoreDetailLine",
			"player", strconv.Itoa(i),
			"cards", strconv.Itoa(det.Cards[i]),
			"espadas", strconv.Itoa(det.Espadas[i]),
			"sevens", strconv.Itoa(det.Sevens[i]),
			"oros", strconv.Itoa(det.Oros[i]),
			"escobas", strconv.Itoa(det.Escobas[i]),
			"gained", strconv.Itoa(det.Gained[i])) + "\n")
	}
	b.WriteString(i18n.Tf("escoba.specialLine",
		"ace", strconv.Itoa(det.AceEsp),
		"sete", strconv.Itoa(det.SeteEsp)) + "\n")
}

// HintOutput emits a capture recommendation for the human's turn: the hand
// card and the table cards summing to 15 it captures (flagging an escoba when
// it clears the table), reusing the domain's GetValidCaptures.
func (p *EscobaCuiPresenter) HintOutput(eg interfaces.EscobaGame) string {
	if eg.GetPhase() != domain.EscobaPhasePlayerTurn {
		return i18n.T("escoba.hintNone") + "\n"
	}
	turn := eg.GetCurrentTurn()
	player := eg.GetPlayer(turn)
	if player == nil || !player.GetIsHuman() {
		return i18n.T("escoba.hintNone") + "\n"
	}
	table := eg.GetTableCards()
	bestHand := -1
	var bestCap []int
	bestEscoba := false
	for i := 0; i < player.GetCardsSize(); i++ {
		for _, cap := range eg.GetValidCaptures(i) {
			isEscoba := len(table) > 0 && len(cap) == len(table)
			switch {
			case bestHand == -1:
			case isEscoba && !bestEscoba:
			case isEscoba == bestEscoba && len(cap) > len(bestCap):
			default:
				continue
			}
			bestHand, bestCap, bestEscoba = i, cap, isEscoba
		}
	}
	if bestHand == -1 {
		return color.Yellow(i18n.T("escoba.hintNoCapture")) + "\n"
	}
	capCards := make([]*domain.Card, 0, len(bestCap))
	for _, idx := range bestCap {
		capCards = append(capCards, table[idx])
	}
	key := "escoba.hintCapture"
	if bestEscoba {
		key = "escoba.hintEscoba"
	}
	return color.Yellow(i18n.Tf(key,
		"played", cuiCardSliceStr([]*domain.Card{player.GetCard(bestHand)}),
		"captured", cuiCardSliceStr(capCards))) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *EscobaCuiPresenter) ActionLogOutput(eg interfaces.EscobaGame) string {
	return actionLogOutputText(eg)
}
