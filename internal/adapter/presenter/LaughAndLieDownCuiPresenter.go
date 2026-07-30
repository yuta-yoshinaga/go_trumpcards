//go:build !js || !wasm || extra2

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func laughAndLieDownCardListStr(cards []*domain.Card, indexed bool) string {
	if len(cards) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(cards))
	for i, c := range cards {
		if indexed {
			parts = append(parts, "["+strconv.Itoa(i)+"]"+cuiCardStr(c))
			continue
		}
		parts = append(parts, cuiCardStr(c))
	}
	return strings.Join(parts, " ")
}

// LaughAndLieDownCuiPresenter renders the Laugh and Lie Down CUI view.
type LaughAndLieDownCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *LaughAndLieDownCuiPresenter) Output(c interfaces.LaughAndLieDownGame, lastErr error) string {
	return buildCuiOutput(i18n.T("laughandliedown.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("laughandliedown.header",
			"pot", strconv.Itoa(domain.LaughAndLieDownPot),
			"dealer", strconv.Itoa(c.GetDealerIdx())) + "\n")
		sb.WriteString(i18n.T("laughandliedown.ruleLine") + "\n")
		// 場は伏せた山ではなく広がった札。どのランクが何枚残っているかが
		// 見えていないと 3 枚取りの判断ができないので、常に全部出す。
		sb.WriteString(i18n.Tf("laughandliedown.layoutLine",
			"cards", laughAndLieDownCardListStr(c.GetLayout(), false)) + "\n")

		for i, player := range c.GetPlayers() {
			line := i18n.Tf("laughandliedown.playerLine",
				"name", cuiPlayerName(player, i),
				"cards", strconv.Itoa(player.GetCardsSize()),
				"won", strconv.Itoa(c.GetWonCount(i)))
			if c.IsLaidDown(i) {
				line += " " + i18n.T("laughandliedown.laidDownMark")
			}
			sb.WriteString(line + "\n")
			if player.GetIsHuman() && player.GetCardsSize() > 0 {
				hand := make([]*domain.Card, 0, player.GetCardsSize())
				for j := range player.GetCardsSize() {
					hand = append(hand, player.GetCard(j))
				}
				sb.WriteString("  " + laughAndLieDownCardListStr(hand, true) + "\n")
			}
		}

		sb.WriteString("----------\n")
		cuiErrorBlock(sb, lastErr)

		if c.GetGameEndFlag() {
			sb.WriteString(p.endBlock(c))
			return
		}
		sb.WriteString(i18n.T("laughandliedown.promptPlay") + "\n")
	})
}

// endBlock は精算結果を描く。
func (p *LaughAndLieDownCuiPresenter) endBlock(c interfaces.LaughAndLieDownGame) string {
	var sb strings.Builder
	if last := c.GetLastInIdx(); last >= 0 {
		sb.WriteString(i18n.Tf("laughandliedown.lastInLine",
			"name", cuiPlayerName(c.GetPlayer(last), last),
			"amount", strconv.Itoa(domain.LaughAndLieDownLastInBonus)) + "\n")
	}
	for i := range c.GetPlayers() {
		sb.WriteString(i18n.Tf("laughandliedown.settleLine",
			"name", cuiPlayerName(c.GetPlayer(i), i),
			"won", strconv.Itoa(c.GetWonCount(i)),
			"score", strconv.Itoa(c.GetScore(i))) + "\n")
	}
	banner := i18n.T("laughandliedown.gameEndEven")
	switch {
	case c.GetScore(0) > 0:
		banner = i18n.T("laughandliedown.gameEndWin")
	case c.GetScore(0) < 0:
		banner = i18n.T("laughandliedown.gameEndLose")
	}
	sb.WriteString(color.Green(banner) + "\n")
	return sb.String()
}

// HintOutput emits the current Laugh and Lie Down hint.
func (p *LaughAndLieDownCuiPresenter) HintOutput(c interfaces.LaughAndLieDownGame) string {
	hint := laughAndLieDownHint(c)
	key := laughAndLieDownHintReasonKeys[hint.Reason]
	if key == "" {
		key = "laughandliedown.hintNone"
	}
	if hint.CardIndex == nil {
		return color.Yellow(i18n.T(key)) + "\n"
	}
	return color.Yellow(i18n.Tf("laughandliedown.hintPlay",
		"idx", strconv.Itoa(*hint.CardIndex),
		"take", strconv.Itoa(hint.TakeCount),
		"reason", i18n.T(key))) + "\n"
}

// laughAndLieDownHintReasonKeys maps the reason identifiers the hint returns to
// i18n keys. The Web presenter ships the identifier and the frontend resolves
// it; the CUI must resolve it here or it prints the raw key.
var laughAndLieDownHintReasonKeys = map[string]string{
	"laughandliedown.hint.game_end":      "laughandliedown.hintReasonGameEnd",
	"laughandliedown.hint.not_your_turn": "laughandliedown.hintReasonNotYourTurn",
	"laughandliedown.hint.must_lie_down": "laughandliedown.hintReasonMustLieDown",
	"laughandliedown.hint.take_one":      "laughandliedown.hintReasonTakeOne",
	"laughandliedown.hint.take_three":    "laughandliedown.hintReasonTakeThree",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *LaughAndLieDownCuiPresenter) ActionLogOutput(c interfaces.LaughAndLieDownGame) string {
	return actionLogOutputText(c)
}
