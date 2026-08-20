//go:build !js || !wasm || extra3

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// SakuraCuiPresenter renders the Sakura (さくら/肥後花) CUI view (花札; no French suits).
type SakuraCuiPresenter struct{}

// sakuraCuiCardStr は花札を "松·光(20)" のように点数つきで描画する。
//
// **点数を札そのものに書く。** さくらは役ではなく点数の合計で競うので、どの札が
// 何点かが読めないと打ち手を決められない。
func sakuraCuiCardStr(c *domain.Card) string {
	if c == nil {
		return "??"
	}
	label := domain.KoiKoiCardLabel(c) + "(" + strconv.Itoa(domain.SakuraCardPoints(c)) + ")"
	switch domain.KoiKoiCardCategory(c) {
	case domain.KoiKoiBright:
		return color.Yellow(label)
	case domain.KoiKoiRibbon:
		if domain.KoiKoiCardRibbonColor(c) == domain.KoiKoiRibbonBlue {
			return label
		}
		return color.Red(label)
	default:
		return label
	}
}

// sakuraCuiCardsStr は札の並びをインデックス付きで描画する。
func sakuraCuiCardsStr(cards []*domain.Card) string {
	if len(cards) == 0 {
		return "-"
	}
	parts := make([]string, len(cards))
	for i, c := range cards {
		parts[i] = "[" + strconv.Itoa(i) + "]" + sakuraCuiCardStr(c)
	}
	return strings.Join(parts, " ")
}

// sakuraBonusStr は成立した追加役を "桜酒(30)" 風に描画する。
func sakuraBonusStr(bonuses []domain.SakuraBonus) string {
	if len(bonuses) == 0 {
		return "-"
	}
	parts := make([]string, len(bonuses))
	for i, b := range bonuses {
		parts[i] = i18n.T("sakura.bonus."+domain.SakuraBonusName(b)) +
			"(" + strconv.Itoa(domain.SakuraBonusPoints(b)) + ")"
	}
	return strings.Join(parts, ", ")
}

// sakuraPlayerStr は 1 席ぶんの表示を返す。
func sakuraPlayerStr(g interfaces.SakuraGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("sakura.playerLine",
		"name", cuiPlayerName(player, idx),
		"hand", strconv.Itoa(player.GetCardsSize()),
		"taken", strconv.Itoa(len(player.GetTaken())),
		"points", strconv.Itoa(player.TotalPoints()),
		"bonus", sakuraBonusStr(player.Bonuses()),
		"score", strconv.Itoa(player.GetScore()),
		"wins", strconv.Itoa(player.GetRoundWins())) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(i18n.Tf("sakura.handLine",
			"hand", sakuraCuiCardsStr(player.GetCards())) + "\n")
	}
	return b.String()
}

// Output renders the current game state for the active locale.
func (p *SakuraCuiPresenter) Output(g interfaces.SakuraGame, lastErr error) string {
	return buildCuiOutput(i18n.T("sakura.helpTitle"), func(b *strings.Builder) {
		cfg := g.GetConfig()
		b.WriteString(i18n.Tf("sakura.roundLine",
			"round", strconv.Itoa(g.GetRound()),
			"rounds", strconv.Itoa(cfg.Rounds),
			"stock", strconv.Itoa(g.GetStockCount()),
			"dealer", cuiPlayerName(g.GetPlayer(g.GetDealer()), g.GetDealer())) + "\n")
		b.WriteString(i18n.Tf("sakura.fieldLine", "field", sakuraCuiCardsStr(g.GetField())) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(sakuraPlayerStr(g, i))
		}
		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch g.GetPhase() {
		case domain.SakuraPhasePlay:
			cur := g.GetTurn()
			b.WriteString(i18n.Tf("sakura.promptPlay",
				"name", cuiPlayerName(g.GetPlayer(cur), cur)) + "\n")
		case domain.SakuraPhaseRoundEnd:
			b.WriteString(sakuraRoundResultStr(g) + "\n")
		case domain.SakuraPhaseGameEnd:
			b.WriteString(sakuraGameResultStr(g) + "\n")
		}
		b.WriteString(i18n.T("sakura.promptHelp") + "\n")
	})
}

// sakuraRoundResultStr はラウンド結果の説明文を返す。
func sakuraRoundResultStr(g interfaces.SakuraGame) string {
	res := g.GetLastResult()
	if res == nil || res.Winner < 0 {
		return i18n.T("sakura.roundDraw")
	}
	seat := res.Seats[res.Winner]
	return i18n.Tf("sakura.roundWin",
		"name", cuiPlayerName(g.GetPlayer(res.Winner), res.Winner),
		"total", strconv.Itoa(seat.Total),
		"card", strconv.Itoa(seat.CardPoints),
		"bonus", sakuraBonusStr(seat.Bonuses))
}

// sakuraGameResultStr は終局の説明文を返す。
func sakuraGameResultStr(g interfaces.SakuraGame) string {
	if g.GetWinner() < 0 {
		return i18n.T("sakura.gameDraw")
	}
	winner := g.GetWinner()
	return i18n.Tf("sakura.gameWin",
		"name", cuiPlayerName(g.GetPlayer(winner), winner),
		"score", strconv.Itoa(g.GetPlayer(winner).GetScore()))
}

// HintOutput emits the current Sakura hint.
func (p *SakuraCuiPresenter) HintOutput(g interfaces.SakuraGame) string {
	hint := g.GetHint()
	if hint.CardIndex < 0 {
		return i18n.T("sakura.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, sakuraHintReasonKeys)
	card := "-"
	if player := g.GetPlayer(g.GetTurn()); player != nil &&
		hint.CardIndex < player.GetCardsSize() {
		card = "[" + strconv.Itoa(hint.CardIndex) + "]" + sakuraCuiCardStr(player.GetCard(hint.CardIndex))
	}
	return color.Yellow(i18n.Tf("sakura.hintCard", "card", card, "reason", reason)) + "\n"
}

// sakuraHintReasonKeys maps Sakura-specific hint reasons to i18n keys.
var sakuraHintReasonKeys = map[string]string{
	"capture": "sakura.hintReasonCapture",
	"discard": "sakura.hintReasonDiscard",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *SakuraCuiPresenter) ActionLogOutput(g interfaces.SakuraGame) string {
	return actionLogOutputTextForSeats[*domain.SakuraPlayer](g)
}
