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

// guandanComboLabel は役の表示名を返す。
func guandanComboLabel(kind domain.GuandanComboKind) string {
	switch kind {
	case domain.GuandanComboSingle:
		return i18n.T("guandan.comboSingle")
	case domain.GuandanComboPair:
		return i18n.T("guandan.comboPair")
	case domain.GuandanComboTriple:
		return i18n.T("guandan.comboTriple")
	case domain.GuandanComboFullHouse:
		return i18n.T("guandan.comboFullHouse")
	case domain.GuandanComboStraight:
		return i18n.T("guandan.comboStraight")
	case domain.GuandanComboPlate:
		return i18n.T("guandan.comboPlate")
	case domain.GuandanComboTube:
		return i18n.T("guandan.comboTube")
	case domain.GuandanComboBomb:
		return i18n.T("guandan.comboBomb")
	case domain.GuandanComboStraightFlush:
		return i18n.T("guandan.comboStraightFlush")
	case domain.GuandanComboJokerBomb:
		return i18n.T("guandan.comboJokerBomb")
	}
	return "-"
}

// guandanLevelLabel はレベル札の表示名を返す。**2〜A なので数字のままでは読めない。**
func guandanLevelLabel(level int) string {
	switch level {
	case 11:
		return "J"
	case 12:
		return "Q"
	case 13:
		return "K"
	case domain.GuandanMaxLevel:
		return "A"
	}
	return strconv.Itoa(level)
}

// guandanPlayerStr は 1 席分の表示文字列を返す。
func guandanPlayerStr(g interfaces.GuandanGame, i int) string {
	player := g.GetPlayer(i)
	if player == nil {
		return ""
	}
	// **味方の手札も見えない。**掼蛋は手札を隠したまま進む。
	hand := i18n.Tf("guandan.hiddenHand", "count", strconv.Itoa(player.GetCardsSize()))
	if player.GetIsHuman() || g.GetGameEndFlag() {
		var b strings.Builder
		for j := range player.GetCardsSize() {
			b.WriteString(strconv.Itoa(j) + ":" + cuiCardStr(player.GetCard(j)) + " ")
		}
		hand = strings.TrimSpace(b.String())
	}
	turn := ""
	if !g.GetGameEndFlag() && i == g.GetCurrentPlayerIdx() {
		turn = " " + i18n.T("guandan.turnTag")
	}
	// **上がり順は次局の貢と昇級量を決める。**着順が見えないと何も判断できない。
	rank := "-"
	for pos, seat := range g.GetFinished() {
		if seat == i {
			rank = strconv.Itoa(pos + 1)
		}
	}
	return i18n.Tf("guandan.playerLine",
		"seat", strconv.Itoa(i),
		"name", cuiPlayerName(player, i),
		"team", strconv.Itoa(domain.GuandanTeamOf(i)),
		"rank", rank,
		"turn", turn,
		"count", strconv.Itoa(player.GetCardsSize()),
		"hand", hand) + "\n"
}

// GuandanCuiPresenter 掼蛋 (Guandan) CUIプレゼンタークラス
type GuandanCuiPresenter struct{}

// Output はゲーム状態を現在のロケールで描画する。
func (p *GuandanCuiPresenter) Output(g interfaces.GuandanGame, lastErr error) string {
	return buildCuiOutput(i18n.T("guandan.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("guandan.header",
			"hand", strconv.Itoa(g.GetHandNumber()),
			"level", guandanLevelLabel(g.GetLevel()),
			"declarer", strconv.Itoa(g.GetDeclarerTeam()),
			"t0", guandanLevelLabel(g.GetTeamLevel(0)),
			"t1", guandanLevelLabel(g.GetTeamLevel(1))) + "\n")
		// **レベル札は A の上、黒ジョーカーの下。**このゲームの肝なので毎回出す。
		b.WriteString(i18n.Tf("guandan.levelNote", "level", guandanLevelLabel(g.GetLevel())) + "\n")

		for i := range g.GetPlayers() {
			b.WriteString(guandanPlayerStr(g, i))
		}

		if c := g.GetLastCombo(); c != nil {
			b.WriteString(i18n.Tf("guandan.tableCombo",
				"combo", guandanComboLabel(c.Kind),
				"size", strconv.Itoa(c.Size),
				"seat", strconv.Itoa(g.GetLastPlayerIdx())) + "\n")
		} else {
			b.WriteString(i18n.T("guandan.tableEmpty") + "\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			b.WriteString(color.Green(i18n.Tf("guandan.gameEnd",
				"team", strconv.Itoa(g.GetWinnerTeam()))) + "\n")
			return
		}

		switch g.GetPhase() {
		case domain.GuandanPhaseTribute:
			p.writeTribute(b, g)
		case domain.GuandanPhaseHandEnd:
			p.writeHandEnd(b, g)
		default:
			b.WriteString(i18n.Tf("guandan.promptTurn",
				"name", cuiPlayerName(g.GetPlayer(g.GetCurrentPlayerIdx()), g.GetCurrentPlayerIdx())) + "\n")
			b.WriteString(i18n.T("guandan.promptHelp") + "\n")
		}
	})
}

// writeTribute は進貢・還貢の状況を書き出す。
func (p *GuandanCuiPresenter) writeTribute(b *strings.Builder, g interfaces.GuandanGame) {
	// **赤ジョーカー 2 枚で貢は流れる。**流れた理由が出ないと不可解に見える。
	if g.IsTributeCancelled() {
		b.WriteString(i18n.T("guandan.tributeCancelled") + "\n")
		b.WriteString(i18n.T("guandan.promptNext") + "\n")
		return
	}
	for _, t := range g.GetTributes() {
		if t == nil {
			continue
		}
		key := "guandan.tributePending"
		if t.Returned != nil {
			key = "guandan.tributeDone"
		}
		b.WriteString(i18n.Tf(key,
			"from", strconv.Itoa(t.From),
			"to", strconv.Itoa(t.To),
			"card", cuiCardStr(t.Card),
			"back", cuiCardStr(t.Returned)) + "\n")
	}
	b.WriteString(i18n.T("guandan.promptTribute") + "\n")
}

// writeHandEnd は局の精算を書き出す。
func (p *GuandanCuiPresenter) writeHandEnd(b *strings.Builder, g interfaces.GuandanGame) {
	if r := g.GetLastResult(); r != nil {
		key := "guandan.handResult"
		if r.FirstSecond {
			// **1 着 2 着の独占は +4。**通常の +1 と混ざると意味が消える。
			key = "guandan.handResultFirstSecond"
		}
		b.WriteString(i18n.Tf(key,
			"team", strconv.Itoa(r.WinnerTeam),
			"advance", strconv.Itoa(r.Advance),
			"level", guandanLevelLabel(g.GetTeamLevel(r.WinnerTeam))) + "\n")
	}
	b.WriteString(i18n.T("guandan.promptNext") + "\n")
}

// ActionLogOutput は棋譜をテキストで出力する。
func (p *GuandanCuiPresenter) ActionLogOutput(g interfaces.GuandanGame) string {
	return actionLogOutputText(g)
}
