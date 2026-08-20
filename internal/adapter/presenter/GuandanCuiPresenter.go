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
	return actionLogOutputTextForSeats[*domain.GuandanPlayer](g)
}

// guandanHumanIdx は人間の席を返す (いなければ -1)。
func guandanHumanIdx(g interfaces.GuandanGame) int {
	for i := range domain.GuandanPlayerCnt {
		if p := g.GetPlayer(i); p != nil && p.GetIsHuman() {
			return i
		}
	}
	return -1
}

// CheckOutput は手札の組み合わせが何の役になるかを下読みする。
//
// **CUI は打って初めて拒否される。**Web は選択に合わせてカード下に役名と
// 「場のどの役にも勝つ」を出しているのに、CUI には判定手段が無く、無効役だと
// 分かるのが play を弾かれたときだった (#5734)。
func (p *GuandanCuiPresenter) CheckOutput(g interfaces.GuandanGame, idxs []int) string {
	human := guandanHumanIdx(g)
	if human < 0 {
		return color.Yellow(i18n.T("guandan.checkNoHand")) + "\n"
	}
	player := g.GetPlayer(human)
	if len(idxs) == 0 {
		return color.Yellow(i18n.T("guandan.checkNeedsIndexes")) + "\n"
	}

	cards := make([]*domain.Card, 0, len(idxs))
	seen := make(map[int]bool, len(idxs))
	for _, i := range idxs {
		if i < 0 || i >= player.GetCardsSize() || seen[i] {
			// **同じ添字を 2 度書いた場合も無効。**手札は 1 枚しかないので、
			// 通してしまうと存在しない役を「作れる」と答えることになる。
			return color.Yellow(i18n.Tf("guandan.checkOutOfRange",
				"val", strconv.Itoa(i), "max", strconv.Itoa(player.GetCardsSize()-1))) + "\n"
		}
		seen[i] = true
		cards = append(cards, player.GetCard(i))
	}

	combo := domain.GuandanEvaluate(cards, g.GetLevel())
	if combo == nil || combo.Kind == domain.GuandanComboNone {
		return color.Yellow(i18n.T("guandan.checkInvalid")) + "\n"
	}

	msg := i18n.Tf("guandan.checkCombo",
		"combo", guandanComboLabel(combo.Kind), "size", strconv.Itoa(combo.Size))
	if domain.GuandanIsBomb(combo.Kind) {
		msg += " " + i18n.T("guandan.checkBeatsAll")
	}
	switch last := g.GetLastCombo(); {
	case last == nil:
		msg += " " + i18n.T("guandan.checkLead")
	case domain.GuandanBeats(combo, last):
		msg += " " + i18n.T("guandan.checkBeatsTable")
	default:
		msg += " " + i18n.Tf("guandan.checkLosesToTable",
			"combo", guandanComboLabel(last.Kind), "size", strconv.Itoa(last.Size))
	}
	return color.Yellow(msg) + "\n"
}
