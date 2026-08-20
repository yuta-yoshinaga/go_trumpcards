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

// shengJiLevelLabel はレベル札の表示名を返す。**2〜A なので数字のままでは読めない。**
func shengJiLevelLabel(level int) string {
	switch level {
	case 11:
		return "J"
	case 12:
		return "Q"
	case 13:
		return "K"
	case domain.ShengJiMaxLevel:
		return "A"
	}
	return strconv.Itoa(level)
}

// shengJiSuitLabel はスートの表示名を返す (0 は無主)。
func shengJiSuitLabel(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return "♠"
	case domain.CardDesignClover:
		return "♣"
	case domain.CardDesignHeart:
		return "♥"
	case domain.CardDesignDiamond:
		return "♦"
	}
	return i18n.T("shengji.noTrump")
}

// shengJiComboLabel は手の形の表示名を返す。
func shengJiComboLabel(kind domain.ShengJiComboKind) string {
	switch kind {
	case domain.ShengJiComboSingle:
		return i18n.T("shengji.comboSingle")
	case domain.ShengJiComboPair:
		return i18n.T("shengji.comboPair")
	case domain.ShengJiComboTractor:
		return i18n.T("shengji.comboTractor")
	}
	return "-"
}

// shengJiPlayerStr は 1 席分の表示文字列を返す。
func shengJiPlayerStr(g interfaces.ShengJiGame, i int) string {
	player := g.GetPlayer(i)
	if player == nil {
		return ""
	}
	// **味方の手札も見えない。**
	hand := i18n.Tf("shengji.hiddenHand", "count", strconv.Itoa(player.GetCardsSize()))
	if player.GetIsHuman() || g.GetGameEndFlag() {
		var b strings.Builder
		for j := range player.GetCardsSize() {
			b.WriteString(strconv.Itoa(j) + ":" + cuiCardStr(player.GetCard(j)) + " ")
		}
		hand = strings.TrimSpace(b.String())
	}
	turn := ""
	if !g.GetGameEndFlag() && i == g.GetCurrentPlayerIdx() {
		turn = " " + i18n.T("shengji.turnTag")
	}
	// **宣言側と守備側で目的が逆。**どちら側かが見えないと打ちようがない。
	role := i18n.T("shengji.defender")
	if domain.ShengJiTeamOf(i) == g.GetDeclarerTeam() {
		role = i18n.T("shengji.declarer")
	}
	return i18n.Tf("shengji.playerLine",
		"seat", strconv.Itoa(i),
		"name", cuiPlayerName(player, i),
		"team", strconv.Itoa(domain.ShengJiTeamOf(i)),
		"role", role,
		"turn", turn,
		"count", strconv.Itoa(player.GetCardsSize()),
		"hand", hand) + "\n"
}

// ShengJiCuiPresenter 升级 (Sheng Ji) CUIプレゼンタークラス
type ShengJiCuiPresenter struct{}

// Output はゲーム状態を現在のロケールで描画する。
func (p *ShengJiCuiPresenter) Output(g interfaces.ShengJiGame, lastErr error) string {
	return buildCuiOutput(i18n.T("shengji.helpTitle"), func(b *strings.Builder) {
		defenders := 1 - g.GetDeclarerTeam()
		b.WriteString(i18n.Tf("shengji.header",
			"hand", strconv.Itoa(g.GetHandNumber()),
			"level", shengJiLevelLabel(g.GetLevel()),
			"trump", shengJiSuitLabel(g.GetTrumpSuit()),
			"declarer", strconv.Itoa(g.GetDeclarerTeam()),
			"t0", shengJiLevelLabel(g.GetTeamLevel(0)),
			"t1", shengJiLevelLabel(g.GetTeamLevel(1))) + "\n")
		// **切札は切札スートだけではない。**これが読めないと序列が分からない。
		b.WriteString(i18n.Tf("shengji.trumpNote", "level", shengJiLevelLabel(g.GetLevel())) + "\n")
		// **点を集めるのは守備側。**
		b.WriteString(i18n.Tf("shengji.pointsLine",
			"points", strconv.Itoa(g.GetTeamPoints(defenders)),
			"target", strconv.Itoa(domain.ShengJiDefenderTarget),
			"total", strconv.Itoa(domain.ShengJiTotalPoints),
			"team", strconv.Itoa(defenders)) + "\n")

		for i := range g.GetPlayers() {
			b.WriteString(shengJiPlayerStr(g, i))
		}

		p.writeTrick(b, g)

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			b.WriteString(color.Green(i18n.Tf("shengji.gameEnd",
				"team", strconv.Itoa(g.GetWinnerTeam()))) + "\n")
			return
		}

		switch g.GetPhase() {
		case domain.ShengJiPhaseDeclare:
			p.writeDeclare(b, g)
		case domain.ShengJiPhaseKitty:
			// **なぜ埋めてはいけないのかまで言う。**Web は kittyRules で
			// 倍率リスクを説明しているのに、CUI は禁止だけを伝えていた (#5735)。
			b.WriteString(i18n.Tf("shengji.kittyPrompt",
				"count", strconv.Itoa(domain.ShengJiKittySize),
				"mult", strconv.Itoa(domain.ShengJiKittyMultiplierPerCard)) + "\n")
		case domain.ShengJiPhaseHandEnd:
			p.writeHandEnd(b, g)
		default:
			b.WriteString(i18n.Tf("shengji.promptTurn",
				"name", cuiPlayerName(g.GetPlayer(g.GetCurrentPlayerIdx()), g.GetCurrentPlayerIdx())) + "\n")
			b.WriteString(i18n.T("shengji.promptHelp") + "\n")
		}
	})
}

// writeTrick はいまのトリックを書き出す。
func (p *ShengJiCuiPresenter) writeTrick(b *strings.Builder, g interfaces.ShengJiGame) {
	trick := g.GetTrick()
	if len(trick) == 0 {
		b.WriteString(i18n.T("shengji.trickEmpty") + "\n")
		return
	}
	if c := g.GetLeadCombo(); c != nil {
		b.WriteString(i18n.Tf("shengji.trickLead",
			"combo", shengJiComboLabel(c.Kind),
			"size", strconv.Itoa(c.Size)) + "\n")
	}
	for i, play := range trick {
		var cards strings.Builder
		for _, c := range play {
			cards.WriteString(cuiCardStr(c) + " ")
		}
		b.WriteString("  " + i18n.Tf("shengji.trickLine",
			"seat", strconv.Itoa((g.GetTrickLeader()+i)%domain.ShengJiPlayerCnt),
			"cards", strings.TrimSpace(cards.String())) + "\n")
	}
}

// writeDeclare は亮牌の状況を書き出す。
func (p *ShengJiCuiPresenter) writeDeclare(b *strings.Builder, g interfaces.ShengJiGame) {
	if d := g.GetDeclaration(); d != nil {
		b.WriteString(i18n.Tf("shengji.declaredLine",
			"seat", strconv.Itoa(d.Seat),
			"suit", shengJiSuitLabel(d.Suit),
			"strength", strconv.Itoa(d.Strength)) + "\n")
	} else {
		b.WriteString(i18n.T("shengji.notDeclared") + "\n")
	}

	// **持っていないスートは宣言できない。**出せる選択肢を並べる。
	seat := g.GetCurrentPlayerIdx()
	if player := g.GetPlayer(seat); player != nil && player.GetIsHuman() {
		var opts strings.Builder
		for suit := domain.CardDesignSpade; suit <= domain.CardDesignDiamond; suit++ {
			if st := g.ShengJiDeclareStrength(seat, suit); st > 0 {
				opts.WriteString(strconv.Itoa(suit) + "=" + shengJiSuitLabel(suit) +
					"(x" + strconv.Itoa(st) + ") ")
			}
		}
		if opts.Len() > 0 {
			b.WriteString(i18n.Tf("shengji.declarable", "suits", strings.TrimSpace(opts.String())) + "\n")
		} else {
			b.WriteString(i18n.T("shengji.noDeclarable") + "\n")
		}
	}
	b.WriteString(i18n.T("shengji.declarePrompt") + "\n")
}

// writeHandEnd は局の精算を書き出す。
func (p *ShengJiCuiPresenter) writeHandEnd(b *strings.Builder, g interfaces.ShengJiGame) {
	if r := g.GetLastResult(); r != nil {
		key := "shengji.handTakenLine"
		if r.DeclarerHeld {
			key = "shengji.handHeldLine"
		}
		b.WriteString(i18n.Tf(key,
			"points", strconv.Itoa(r.DefenderPoints),
			"target", strconv.Itoa(domain.ShengJiDefenderTarget),
			"team", strconv.Itoa(r.AdvancingTeam),
			"advance", strconv.Itoa(r.Advance)) + "\n")
		// **底牌の倍率は最終トリックを取った側にしか掛からない。**
		if r.KittyMultiplier > 0 {
			b.WriteString(i18n.Tf("shengji.kittyLine",
				"points", strconv.Itoa(r.KittyPoints),
				"mult", strconv.Itoa(r.KittyMultiplier)) + "\n")
		}
	}
	b.WriteString(i18n.T("shengji.promptNext") + "\n")
}

// ActionLogOutput は棋譜をテキストで出力する。
func (p *ShengJiCuiPresenter) ActionLogOutput(g interfaces.ShengJiGame) string {
	return actionLogOutputTextForSeats[*domain.ShengJiPlayer](g)
}
