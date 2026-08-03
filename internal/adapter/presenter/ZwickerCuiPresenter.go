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

// zwickerValuesStr は札の取りうるマッチ値を "1/11" のように並べる。
func zwickerValuesStr(c *domain.Card) string {
	vals := domain.ZwickerCardValues(c)
	parts := make([]string, 0, len(vals))
	for _, v := range vals {
		parts = append(parts, strconv.Itoa(v))
	}
	return strings.Join(parts, "/")
}

// zwickerCardListStr は札の並びを、マッチ値つきで返す。
//
// **値を書かないと遊べない。**A と絵札は 2 つの値を持ち、ジョーカーは 15/20/25
// なので、見た目のランクからは何と取れるか分からない。
func zwickerCardListStr(cards []*domain.Card, indexed bool) string {
	if len(cards) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(cards))
	for i, c := range cards {
		s := cuiCardStr(c) + "(" + zwickerValuesStr(c) + ")"
		if indexed {
			s = "[" + strconv.Itoa(i) + "]" + s
		}
		parts = append(parts, s)
	}
	return strings.Join(parts, " ")
}

// ZwickerCuiPresenter renders the Zwicker CUI view.
type ZwickerCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *ZwickerCuiPresenter) Output(c interfaces.ZwickerGame, lastErr error) string {
	return buildCuiOutput(i18n.T("zwicker.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("zwicker.header",
			"stock", strconv.Itoa(c.GetStockCount()),
			"us", strconv.Itoa(c.GetTeamScore(domain.ZwickerTeamOf(0))),
			"them", strconv.Itoa(c.GetTeamScore(1-domain.ZwickerTeamOf(0))),
			"target", strconv.Itoa(c.GetConfig().TargetScore)) + "\n")
		// 2 つの値を持つ札があること、Zwick が「場を空にする」ことの 2 点が
		// 最も誤解されやすいので毎回出す。
		sb.WriteString(i18n.T("zwicker.ruleLine") + "\n")
		sb.WriteString(i18n.Tf("zwicker.tableLine",
			"cards", zwickerCardListStr(c.GetTableCards(), true)) + "\n")

		for i, b := range c.GetBuilds() {
			if b == nil {
				continue
			}
			sb.WriteString(i18n.Tf("zwicker.buildLine",
				"idx", strconv.Itoa(i),
				"value", strconv.Itoa(b.Value),
				"owner", strconv.Itoa(b.Owner),
				"cards", zwickerCardListStr(b.Cards, false)) + "\n")
		}

		for i, player := range c.GetPlayers() {
			if player == nil {
				continue
			}
			sb.WriteString(i18n.Tf("zwicker.playerLine",
				"name", cuiPlayerName(player, i),
				"team", strconv.Itoa(domain.ZwickerTeamOf(i)),
				"cards", strconv.Itoa(player.GetCardsSize()),
				"taken", strconv.Itoa(len(player.GetCaptured())),
				"zwicks", strconv.Itoa(player.GetZwicks())) + "\n")
			if player.GetIsHuman() && player.GetCardsSize() > 0 {
				hand := make([]*domain.Card, 0, player.GetCardsSize())
				for j := range player.GetCardsSize() {
					hand = append(hand, player.GetCard(j))
				}
				sb.WriteString("  " + zwickerCardListStr(hand, true) + "\n")
			}
		}

		sb.WriteString("----------\n")
		cuiErrorBlock(sb, lastErr)
		sb.WriteString(p.promptBlock(c))
	})
}

// promptBlock はフェーズごとの案内を返す。
func (p *ZwickerCuiPresenter) promptBlock(c interfaces.ZwickerGame) string {
	if c.GetGameEndFlag() {
		key := "zwicker.gameEndLose"
		if c.GetWinnerTeam() == domain.ZwickerTeamOf(0) {
			key = "zwicker.gameEndWin"
		}
		return color.Green(i18n.T(key)) + "\n"
	}
	switch c.GetPhase() {
	case domain.ZwickerPhasePlay:
		return i18n.T("zwicker.promptPlay") + "\n"
	case domain.ZwickerPhaseRoundEnd:
		var sb strings.Builder
		if s := c.GetLastRoundScore(); s != nil {
			us, them := domain.ZwickerTeamOf(0), 1-domain.ZwickerTeamOf(0)
			sb.WriteString(i18n.Tf("zwicker.roundResult",
				"us", strconv.Itoa(s.Total[us]),
				"them", strconv.Itoa(s.Total[them])) + "\n")
			// 枚数最多の 3 点は同数だと誰にも行かない。黙っていると
			// 合計が合わないように見える。
			if s.MajorityTeam < 0 {
				sb.WriteString(i18n.T("zwicker.majorityTied") + "\n")
			}
		}
		sb.WriteString(i18n.T("zwicker.promptNext") + "\n")
		return sb.String()
	case domain.ZwickerPhaseGameEnd:
		return ""
	}
	return ""
}

// HintOutput emits the current Zwicker hint.
func (p *ZwickerCuiPresenter) HintOutput(c interfaces.ZwickerGame) string {
	hint := zwickerHint(c)
	key := zwickerHintReasonKeys[hint.Reason]
	if key == "" {
		key = "zwicker.hintNone"
	}
	switch {
	case hint.Take && hint.CardIndex != nil:
		parts := make([]string, 0, len(hint.TableIdxs))
		for _, i := range hint.TableIdxs {
			parts = append(parts, strconv.Itoa(i))
		}
		return color.Yellow(i18n.Tf("zwicker.hintTake",
			"idx", strconv.Itoa(*hint.CardIndex),
			"value", strconv.Itoa(hint.Value),
			"table", strings.Join(parts, ","),
			"reason", i18n.T(key))) + "\n"
	case hint.CardIndex != nil:
		return color.Yellow(i18n.Tf("zwicker.hintTrail",
			"idx", strconv.Itoa(*hint.CardIndex), "reason", i18n.T(key))) + "\n"
	default:
		return color.Yellow(i18n.T(key)) + "\n"
	}
}

// zwickerHintReasonKeys maps the reason identifiers zwickerHint returns to
// i18n keys. The Web presenter ships the identifier and the frontend resolves
// it; the CUI must resolve it here or it prints the raw key.
var zwickerHintReasonKeys = map[string]string{
	"zwicker.hint.game_end":      "zwicker.hintReasonGameEnd",
	"zwicker.hint.round_end":     "zwicker.hintReasonRoundEnd",
	"zwicker.hint.not_your_turn": "zwicker.hintReasonNotYourTurn",
	"zwicker.hint.take":          "zwicker.hintReasonTake",
	"zwicker.hint.trail":         "zwicker.hintReasonTrail",
	"zwicker.hint.none":          "zwicker.hintNone",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *ZwickerCuiPresenter) ActionLogOutput(c interfaces.ZwickerGame) string {
	return actionLogOutputText(c)
}
