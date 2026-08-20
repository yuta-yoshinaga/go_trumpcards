//go:build !js || !wasm || extra3

package presenter

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// scartoCuiCardStr タロット札の CUI 表示文字列 (切り札 "T{n}"、エクスキューズ "EXC"、
// スート札 "♠5" 等)。標準の cuiCardStr は design 5/6 を扱えないためローカルに用意する。
func scartoCuiCardStr(c *domain.Card) string {
	if c == nil {
		return "??"
	}
	switch c.GetDesign() {
	case domain.ScartoExcuseDesign:
		return color.Green("EXC")
	case domain.ScartoTrumpDesign:
		return color.Yellow(fmt.Sprintf("T%d", c.GetValue()))
	default:
		glyphs := map[int]string{
			domain.CardDesignSpade:   "♠",
			domain.CardDesignClover:  "♣",
			domain.CardDesignHeart:   "♥",
			domain.CardDesignDiamond: "♦",
		}
		g, ok := glyphs[c.GetDesign()]
		if !ok {
			g = "?"
		}
		s := g + scartoRankLabel(c.GetValue())
		if isRedSuit(c.GetDesign()) {
			return color.Red(s)
		}
		return s
	}
}

// scartoCuiDiscardable は通常スカルトに出せる札か (非切り札・非エクスキューズ・非コートの
// ピップ) を返す。ドメインの scartoDiscardable と同じ規則を提示側で再現する。
func scartoCuiDiscardable(c *domain.Card) bool {
	if c == nil || c.GetDesign() == domain.ScartoTrumpDesign || c.GetDesign() == domain.ScartoExcuseDesign {
		return false
	}
	return c.GetValue() < domain.ScartoCourtMin
}

// scartoIndexedHand 人間手札をインデックス付きで表示する。
func scartoIndexedHand(p *domain.ScartoPlayer) string {
	parts := make([]string, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		parts[i] = fmt.Sprintf("[%d]%s", i, scartoCuiCardStr(p.GetCard(i)))
	}
	return strings.Join(parts, "  ")
}

// scartoOutcomeLabel ディール結果の i18n ラベルを返す。
func scartoOutcomeLabel(o domain.ScartoOutcome) string {
	switch o {
	case domain.ScartoOutcomeWin:
		return i18n.T("scarto.outcomeWin")
	case domain.ScartoOutcomeLoss:
		return i18n.T("scarto.outcomeLoss")
	default:
		return i18n.T("scarto.outcomeNone")
	}
}

// scartoPlayerStr プレイヤー 1 行分の状態文字列を返す。
func scartoPlayerStr(g interfaces.ScartoGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	scores := g.GetPlayerScores()
	role := i18n.T("scarto.rolePlayer")
	if idx == g.GetDealerIdx() {
		role = i18n.T("scarto.roleDealer")
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("scarto.playerLine",
		"name", cuiPlayerName(player, idx),
		"role", role,
		"cards", strconv.Itoa(player.GetCardsSize()),
		"score", strconv.Itoa(scores[idx]),
		"tricks", strconv.Itoa(player.GetTrickCount()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(scartoIndexedHand(player) + "\n")
	}
	return b.String()
}

// ScartoCuiPresenter renders the Scarto CUI view.
type ScartoCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *ScartoCuiPresenter) Output(g interfaces.ScartoGame, lastErr error) string {
	return buildCuiOutput(i18n.T("scarto.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("scarto.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber())) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(scartoPlayerStr(g, i))
		}

		b.WriteString("----------\n")

		cuiTrickBlock(b, g.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return scartoCuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winner := g.GetWinnerPlayer()
			var winnerStr string
			if winner >= 0 {
				winnerStr = cuiPlayerName(g.GetPlayer(winner), winner)
			}
			b.WriteString(color.Green(i18n.Tf("scarto.gameEnd", "name", winnerStr)) + "\n")
			return
		}
		p.writePrompt(b, g)
	})
}

// writePrompt 現在のフェーズに応じたプロンプトを書き込む。
func (p *ScartoCuiPresenter) writePrompt(b *strings.Builder, g interfaces.ScartoGame) {
	switch g.GetPhase() {
	case domain.ScartoPhaseScarto:
		dealerIdx := g.GetDealerIdx()
		b.WriteString(i18n.Tf("scarto.promptScarto",
			"name", cuiPlayerName(g.GetPlayer(dealerIdx), dealerIdx)) + "\n")
		b.WriteString(i18n.T("scarto.promptScartoHelp") + "\n")
		// The web UI disables trumps/Excuse/courts; on the CLI, spell out which
		// of the human dealer's cards may actually be discarded, plus the legend
		// of excluded kinds so the choice isn't trial-and-error.
		if dealer := g.GetPlayer(dealerIdx); dealer != nil && dealer.GetIsHuman() {
			var idxs []string
			for i := 0; i < dealer.GetCardsSize(); i++ {
				if scartoCuiDiscardable(dealer.GetCard(i)) {
					idxs = append(idxs, "["+strconv.Itoa(i)+"]"+scartoCuiCardStr(dealer.GetCard(i)))
				}
			}
			list := strings.Join(idxs, "  ")
			if list == "" {
				list = i18n.T("scarto.discardableNone")
			}
			b.WriteString(i18n.Tf("scarto.discardableList", "cards", list) + "\n")
			b.WriteString(i18n.T("scarto.discardableLegend") + "\n")
		}
	case domain.ScartoPhasePlay:
		currentIdx := g.GetCurrentPlayerIdx()
		b.WriteString(i18n.Tf("scarto.promptPlay",
			"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
		b.WriteString(i18n.T("scarto.promptPlayHelp") + "\n")
	case domain.ScartoPhaseTrickEnd:
		b.WriteString(i18n.T("scarto.promptTrickEnd") + "\n")
		b.WriteString(i18n.T("scarto.promptTrickEndHelp") + "\n")
	case domain.ScartoPhaseRoundEnd:
		b.WriteString(i18n.Tf("scarto.promptRoundEnd",
			"outcome", scartoOutcomeLabel(g.GetOutcome())) + "\n")
		// 精算は N × (自分のカード点 − 卓平均) のゼロサム。平均と式を出さないと
		// dealScores の数字がどこから来たのか検算できない (Web は出している)。
		scartoWriteAverageBreakdown(b, g)
		b.WriteString(i18n.T("scarto.promptRoundEndHelp") + "\n")
	}
}

// scartoPointsStr renders a half-point figure, dropping a trailing ".0".
func scartoPointsStr(v float64) string {
	if v == math.Trunc(v) {
		return strconv.Itoa(int(v))
	}
	return strconv.FormatFloat(v, 'f', 1, 64)
}

// scartoSignedPointsStr renders a signed difference ("+3", "-1.5").
func scartoSignedPointsStr(v float64) string {
	if v > 0 {
		return "+" + scartoPointsStr(v)
	}
	return scartoPointsStr(v)
}

// scartoWriteAverageBreakdown appends the table average, the conversion formula
// and each seat's difference from it, mirroring the Web's scarto-breakdown.
func scartoWriteAverageBreakdown(b *strings.Builder, g interfaces.ScartoGame) {
	n := g.GetPlayerCnt()
	if n == 0 {
		return
	}
	total := 0
	for i := 0; i < n; i++ {
		total += g.GetCardPoints(i)
	}
	avg := float64(total) / float64(n)
	b.WriteString(i18n.Tf("scarto.roundEndAverage", "avg", scartoPointsStr(avg)) + "\n")
	b.WriteString(i18n.T("scarto.roundEndFormulaLine") + "\n")
	for i := 0; i < n; i++ {
		points := g.GetCardPoints(i)
		// 変動は整数になる (N×h_i − Σh)。平均差のほうは割り切れないことがある。
		b.WriteString(i18n.Tf("scarto.roundEndEarned",
			"name", cuiPlayerName(g.GetPlayer(i), i),
			"points", strconv.Itoa(points),
			"diff", scartoSignedPointsStr(float64(points)-avg),
			"scaled", scartoSignedPointsStr(float64(n*points-total))) + "\n")
	}
}

// HintOutput emits the current Scarto hint.
func (p *ScartoCuiPresenter) HintOutput(g interfaces.ScartoGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("scarto.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, scartoHintReasonKeys)
	if len(hint.CardIndices) > 0 {
		playerIdx := g.GetCurrentPlayerIdx()
		if g.GetPhase() == domain.ScartoPhaseScarto {
			playerIdx = g.GetDealerIdx()
		}
		player := g.GetPlayer(playerIdx)
		cards := make([]string, len(hint.CardIndices))
		for i, idx := range hint.CardIndices {
			if player != nil && idx >= 0 && idx < player.GetCardsSize() {
				cards[i] = "[" + strconv.Itoa(idx) + "]" + scartoCuiCardStr(player.GetCard(idx))
			} else {
				cards[i] = strconv.Itoa(idx)
			}
		}
		return color.Yellow(i18n.Tf("scarto.hintCard",
			"cards", strings.Join(cards, ", "),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("scarto.hintCard", "cards", "-", "reason", reason)) + "\n"
}

// scartoHintReasonKeys maps hint-reason identifiers to i18n keys.
var scartoHintReasonKeys = map[string]string{
	"scarto_weak": "scarto.hintReasonScartoWeak",
	"lead_low":    "scarto.hintReasonLeadLow",
	"follow_win":  "scarto.hintReasonFollowWin",
	"follow_duck": "scarto.hintReasonFollowDuck",
	"play_excuse": "scarto.hintReasonPlayExcuse",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *ScartoCuiPresenter) ActionLogOutput(g interfaces.ScartoGame) string {
	return actionLogOutputTextForSeats[*domain.ScartoPlayer](g)
}
