//go:build !js || !wasm || extra

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// HachiHachiCuiPresenter renders the Hachi-Hachi (八八) CUI view (花札; no French suits).
type HachiHachiCuiPresenter struct{}

// hachihachiCuiCardStr は花札を "松·光" のように描画する。色は札種に応じて付ける。
func hachihachiCuiCardStr(c *domain.Card) string {
	if c == nil {
		return "??"
	}
	label := domain.HachiHachiCardLabel(c)
	switch domain.HachiHachiCardCategory(c) {
	case domain.HachiHachiBright:
		return color.Yellow(label)
	case domain.HachiHachiRibbon:
		if domain.HachiHachiCardRibbonColor(c) == domain.HachiHachiRibbonBlue {
			return label
		}
		return color.Red(label)
	default:
		return label
	}
}

// hachihachiCuiFieldStr は場札を "[0]松·光 [1]萩·タネ" 形式で返す。
func hachihachiCuiFieldStr(cards []*domain.Card) string {
	if len(cards) == 0 {
		return "-"
	}
	parts := make([]string, len(cards))
	for i, c := range cards {
		parts[i] = "[" + strconv.Itoa(i) + "]" + hachihachiCuiCardStr(c)
	}
	return strings.Join(parts, " ")
}

// hachihachiCuiHandStr は人間の手札をインデックス付きで描画する。
func hachihachiCuiHandStr(p *domain.HachiHachiPlayer) string {
	parts := make([]string, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		parts[i] = "[" + strconv.Itoa(i) + "]" + hachihachiCuiCardStr(p.GetCard(i))
	}
	return strings.Join(parts, " ")
}

// hachihachiYakuStr は成立出来役を "五光(100)" 風に描画する。
func hachihachiYakuStr(yakus []domain.HachiHachiYaku) string {
	if len(yakus) == 0 {
		return "-"
	}
	parts := make([]string, len(yakus))
	for i, y := range yakus {
		parts[i] = i18n.T("hachihachi.yaku."+y.Key) + "(" + strconv.Itoa(y.Points) + ")"
	}
	return strings.Join(parts, ", ")
}

func hachihachiPlayerStr(g interfaces.HachiHachiGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	yakus, raw := g.GetYaku(idx)
	var b strings.Builder
	b.WriteString(i18n.Tf("hachihachi.playerLine",
		"name", cuiPlayerName(player, idx),
		"hand", strconv.Itoa(player.GetCardsSize()),
		"captured", strconv.Itoa(player.CapturedCount()),
		"raw", strconv.Itoa(raw),
		"score", strconv.Itoa(player.GetScore()),
		"yaku", hachihachiYakuStr(yakus)) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(hachihachiCuiHandStr(player) + "\n")
	}
	return b.String()
}

// Output renders the current game state for the active locale.
func (p *HachiHachiCuiPresenter) Output(g interfaces.HachiHachiGame, lastErr error) string {
	return buildCuiOutput(i18n.T("hachihachi.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("hachihachi.roundLine",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"total", strconv.Itoa(g.GetConfig().TargetRounds),
			"deck", strconv.Itoa(g.GetRemainingDeck())) + "\n")
		b.WriteString(i18n.Tf("hachihachi.fieldLine", "field", hachihachiCuiFieldStr(g.GetFieldCards())) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(hachihachiPlayerStr(g, i))
		}
		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch g.GetPhase() {
		case domain.HachiHachiPhasePlay:
			currentIdx := g.GetCurrentTurn()
			b.WriteString(i18n.Tf("hachihachi.promptPlay",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
		case domain.HachiHachiPhaseRoundEnd:
			b.WriteString(hachihachiRoundResultStr(g) + "\n")
		case domain.HachiHachiPhaseGameEnd:
			b.WriteString(i18n.T("hachihachi.promptGameEnd") + "\n")
		}
		b.WriteString(i18n.T("hachihachi.promptHelp") + "\n")
	})
}

// hachihachiRoundResultStr はラウンド精算結果の説明文を返す。
func hachihachiRoundResultStr(g interfaces.HachiHachiGame) string {
	res := g.GetLastRoundResult()
	if res == nil {
		return i18n.T("hachihachi.roundDraw")
	}
	parts := make([]string, 0, len(res.Scores))
	for _, s := range res.Scores {
		name := cuiPlayerName(g.GetPlayer(s.PlayerIdx), s.PlayerIdx)
		line := i18n.Tf("hachihachi.scoreLine",
			"name", name,
			"raw", strconv.Itoa(s.RawScore),
			"bonus", strconv.Itoa(s.Bonus),
			"delta", hachihachiSignedStr(s.Delta))
		// 誰がラウンドを制したかを数字の見比べに頼らせない (Web は 👑 を付ける)。
		// Best は総取りが決まらなければ -1 なので、そのときは誰にも付けない。
		if s.PlayerIdx == res.Best {
			line += "  " + color.Bold(i18n.T("hachihachi.roundBestMark"))
		}
		parts = append(parts, line)
	}
	return i18n.T("hachihachi.roundEnd") + "\n" + strings.Join(parts, "\n")
}

// hachihachiSignedStr は符号付きの整数文字列 ("+12"/"-3") を返す。
func hachihachiSignedStr(n int) string {
	if n > 0 {
		return "+" + strconv.Itoa(n)
	}
	return strconv.Itoa(n)
}

// HintOutput emits the current Hachi-Hachi hint.
func (p *HachiHachiCuiPresenter) HintOutput(g interfaces.HachiHachiGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("hachihachi.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, hachihachiHintReasonKeys)
	card := "-"
	if player := g.GetPlayer(g.GetCurrentTurn()); player != nil &&
		hint.CardIndex >= 0 && hint.CardIndex < player.GetCardsSize() {
		card = "[" + strconv.Itoa(hint.CardIndex) + "]" + hachihachiCuiCardStr(player.GetCard(hint.CardIndex))
	}
	return color.Yellow(i18n.Tf("hachihachi.hintCard", "card", card, "reason", reason)) + "\n"
}

// hachihachiHintReasonKeys maps Hachi-Hachi-specific hint reasons to i18n keys.
var hachihachiHintReasonKeys = map[string]string{
	"capture":     "hachihachi.hintReasonCapture",
	"discard_low": "hachihachi.hintReasonDiscard",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *HachiHachiCuiPresenter) ActionLogOutput(g interfaces.HachiHachiGame) string {
	return actionLogOutputText(g)
}
