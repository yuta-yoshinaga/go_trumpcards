//go:build !js || !wasm || casino

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// ofcRowCards renders a slice of cards as a comma-joined text string.
func ofcRowCards(cards []*domain.Card) string {
	if len(cards) == 0 {
		return "-"
	}
	parts := make([]string, len(cards))
	for i, c := range cards {
		parts[i] = cuiCardStr(c)
	}
	return strings.Join(parts, ", ")
}

// ofcPlayerStr returns the display string for a single player's three rows.
func ofcPlayerStr(g interfaces.OpenFaceChineseGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	var b strings.Builder
	status := ""
	if player.GetFouled() {
		status = " " + i18n.T("openfacechinese.fouled")
	}
	if player.GetFantasyland() {
		status += " " + i18n.T("openfacechinese.fantasyland")
	}
	b.WriteString(i18n.Tf("openfacechinese.playerLine",
		"name", cuiPlayerName(player, idx),
		"total", strconv.Itoa(player.GetTotalScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"status", status) + "\n")
	b.WriteString(i18n.Tf("openfacechinese.rowFront", "cards", ofcRowCards(player.GetFront())) + "\n")
	b.WriteString(i18n.Tf("openfacechinese.rowMiddle", "cards", ofcRowCards(player.GetMiddle())) + "\n")
	b.WriteString(i18n.Tf("openfacechinese.rowBack", "cards", ofcRowCards(player.GetBack())) + "\n")
	return b.String()
}

// OpenFaceChineseCuiPresenter renders the Open Face Chinese (OFC) CUI view.
type OpenFaceChineseCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *OpenFaceChineseCuiPresenter) Output(g interfaces.OpenFaceChineseGame, lastErr error) string {
	return buildCuiOutput(i18n.T("openfacechinese.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("openfacechinese.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"target", strconv.Itoa(g.GetConfig().TargetRounds)) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(ofcPlayerStr(g, i))
			b.WriteString("----------\n")
		}

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			b.WriteString(color.Green(p.gameEndBanner(g)) + "\n")
			return
		}
		p.writePrompt(b, g)
	})
}

// gameEndBanner renders the final match-result banner.
func (p *OpenFaceChineseCuiPresenter) gameEndBanner(g interfaces.OpenFaceChineseGame) string {
	winner := g.GetWinnerIdx()
	if winner < 0 {
		return i18n.T("openfacechinese.gameEndDraw")
	}
	if player := g.GetPlayer(winner); player != nil {
		return i18n.Tf("openfacechinese.gameEnd", "name", cuiPlayerName(player, winner))
	}
	return i18n.T("openfacechinese.gameEndDraw")
}

// writePrompt renders the phase-specific prompt block.
// openFaceChineseFoulWarning は、保留カードを置くと確定で反則になる段を挙げた
// 警告行を返す。該当する段が無ければ空文字。
func openFaceChineseFoulWarning(g interfaces.OpenFaceChineseGame, idx int) string {
	player := g.GetPlayer(idx)
	card := g.GetCurrentCard()
	if player == nil || card == nil || !player.GetIsHuman() {
		return ""
	}
	rows := []struct {
		row int
		key string
	}{
		{domain.OpenFaceChineseRowFront, "openfacechinese.rowFront"},
		{domain.OpenFaceChineseRowMiddle, "openfacechinese.rowMiddle"},
		{domain.OpenFaceChineseRowBack, "openfacechinese.rowBack"},
	}
	var fouling []string
	for _, r := range rows {
		if domain.OpenFaceChinesePlacementFouls(
			player.GetFront(), player.GetMiddle(), player.GetBack(), card, r.row) {
			fouling = append(fouling, i18n.T(r.key))
		}
	}
	if len(fouling) == 0 {
		return ""
	}
	return i18n.Tf("openfacechinese.foulRiskWarning", "rows", strings.Join(fouling, ", "))
}

func (p *OpenFaceChineseCuiPresenter) writePrompt(b *strings.Builder, g interfaces.OpenFaceChineseGame) {
	switch g.GetPhase() {
	case domain.OpenFaceChinesePhasePlacing:
		idx := g.GetCurrentPlayerIdx()
		b.WriteString(i18n.Tf("openfacechinese.promptPlace",
			"name", cuiPlayerName(g.GetPlayer(idx), idx),
			"card", cuiCardStr(g.GetCurrentCard())) + "\n")
		b.WriteString(i18n.T("openfacechinese.promptPlaceHelp") + "\n")
		// **反則は全段負け扱い**という重い結果なのに、CUI は置いてラウンドが
		// 終わるまで気づけなかった (#5676)。Web は各段のボタンにその場で警告を
		// 出している。確定するものだけを出す -- 未確定を反則と呼ぶと、まだ挽回
		// できる配置まで避けさせる。
		if warn := openFaceChineseFoulWarning(g, idx); warn != "" {
			b.WriteString(color.BoldYellow(warn) + "\n")
		}
	case domain.OpenFaceChinesePhaseRoundEnd:
		b.WriteString(i18n.T("openfacechinese.promptRoundEnd") + "\n")
		b.WriteString(i18n.T("openfacechinese.promptRoundEndHelp") + "\n")
	}
}

// HintOutput emits the current Open Face Chinese hint.
func (p *OpenFaceChineseCuiPresenter) HintOutput(g interfaces.OpenFaceChineseGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("openfacechinese.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, openFaceChineseHintReasonKeys)
	return color.Yellow(i18n.Tf("openfacechinese.hintRow",
		"row", openFaceChineseRowName(hint.Row),
		"reason", reason)) + "\n"
}

// openFaceChineseRowName maps a row index to its localized name.
func openFaceChineseRowName(row int) string {
	switch row {
	case domain.OpenFaceChineseRowFront:
		return i18n.T("openfacechinese.rowNameFront")
	case domain.OpenFaceChineseRowMiddle:
		return i18n.T("openfacechinese.rowNameMiddle")
	default:
		return i18n.T("openfacechinese.rowNameBack")
	}
}

// openFaceChineseHintReasonKeys maps OFC-specific hint-reason identifiers to i18n keys.
var openFaceChineseHintReasonKeys = map[string]string{
	"strong_back": "openfacechinese.hintReasonStrongBack",
	"weak_front":  "openfacechinese.hintReasonWeakFront",
	"balance":     "openfacechinese.hintReasonBalance",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *OpenFaceChineseCuiPresenter) ActionLogOutput(g interfaces.OpenFaceChineseGame) string {
	return actionLogOutputTextForSeats[*domain.OpenFaceChinesePlayer](g)
}
