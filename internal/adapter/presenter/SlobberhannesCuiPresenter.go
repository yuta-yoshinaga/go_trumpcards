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

// slobberhannesPlayerStr returns the display string for a single player.
func slobberhannesPlayerStr(player *domain.SlobberhannesPlayer, idx int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("slobberhannes.playerLine",
		"name", cuiPlayerName(player, idx),
		"score", strconv.Itoa(player.GetScore()),
		"penalties", slobberhannesPenaltyStr(player),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// slobberhannesPenaltyStr 受けている罰の内訳を短い記号で表す
func slobberhannesPenaltyStr(player *domain.SlobberhannesPlayer) string {
	marks := make([]string, 0, 3)
	if player.GetTookFirstTrick() {
		marks = append(marks, i18n.T("slobberhannes.markFirst"))
	}
	if player.GetTookLastTrick() {
		marks = append(marks, i18n.T("slobberhannes.markLast"))
	}
	if player.GetTookQueen() {
		marks = append(marks, i18n.T("slobberhannes.markQueen"))
	}
	if len(marks) == 0 {
		return i18n.T("slobberhannes.markClean")
	}
	return strings.Join(marks, ",")
}

// SlobberhannesCuiPresenter renders the Slobberhannes CUI view.
type SlobberhannesCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *SlobberhannesCuiPresenter) Output(s interfaces.SlobberhannesGame, lastErr error) string {
	return buildCuiOutput(i18n.T("slobberhannes.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("slobberhannes.header",
			"round", strconv.Itoa(s.GetRoundNumber()),
			"rounds", strconv.Itoa(s.GetConfig().Rounds),
			"trick", strconv.Itoa(s.GetTrickNumber()+1),
			"tricks", strconv.Itoa(domain.SlobberhannesTricksPerRound)) + "\n")
		sb.WriteString(i18n.T("slobberhannes.noTrumpLine") + "\n")

		// 最初と最後のトリックは位置そのものが罰点対象。警告を出す。
		switch s.GetTrickNumber() {
		case 0:
			sb.WriteString(color.Yellow(i18n.T("slobberhannes.warnFirst")) + "\n")
		case domain.SlobberhannesTricksPerRound - 1:
			sb.WriteString(color.Yellow(i18n.T("slobberhannes.warnLast")) + "\n")
		}

		for i := 0; i < s.GetPlayerCnt(); i++ {
			sb.WriteString(slobberhannesPlayerStr(s.GetPlayer(i), i))
		}

		sb.WriteString("----------\n")

		cuiTrickBlock(sb, s.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(s.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(sb, lastErr)

		if s.GetGameEndFlag() {
			var banner string
			if s.GetWinnerIdx() < 0 {
				banner = i18n.T("slobberhannes.gameEndTie")
			} else {
				banner = i18n.Tf("slobberhannes.gameEndWinner",
					"name", cuiPlayerName(s.GetPlayer(s.GetWinnerIdx()), s.GetWinnerIdx()),
					"score", strconv.Itoa(s.GetPlayer(s.GetWinnerIdx()).GetScore()))
			}
			sb.WriteString(color.Green(banner) + "\n")
			return
		}

		if s.GetPhase() == domain.SlobberhannesPhaseRoundEnd {
			sb.WriteString(i18n.T("slobberhannes.promptRoundEnd") + "\n")
			sb.WriteString(i18n.T("slobberhannes.promptNext") + "\n")
			return
		}

		currentIdx := s.GetCurrentPlayerIdx()
		sb.WriteString(i18n.Tf("slobberhannes.promptCurrentPlayer",
			"name", cuiPlayerName(s.GetPlayer(currentIdx), currentIdx)) + "\n")
		sb.WriteString(i18n.T("slobberhannes.promptPlay") + "\n")
	})
}

// HintOutput emits the current hint.
func (p *SlobberhannesCuiPresenter) HintOutput(s interfaces.SlobberhannesGame) string {
	hint := s.GetHint()
	if hint == nil || hint.CardIndex == nil {
		return i18n.T("slobberhannes.hintNone") + "\n"
	}
	card := s.GetPlayer(0).GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("slobberhannes.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", hintReasonStr(hint.Reason, slobberhannesHintReasonKeys))) + "\n"
}

// slobberhannesHintReasonKeys maps hint-reason identifiers to their i18n keys.
var slobberhannesHintReasonKeys = map[string]string{
	"slobberhannesAvoid":   "slobberhannes.hintReasonAvoid",
	"slobberhannesDump":    "slobberhannes.hintReasonDump",
	"slobberhannesLeadLow": "slobberhannes.hintReasonLeadLow",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *SlobberhannesCuiPresenter) ActionLogOutput(s interfaces.SlobberhannesGame) string {
	return actionLogOutputText(s)
}
