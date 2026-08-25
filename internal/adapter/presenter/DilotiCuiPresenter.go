//go:build !js || !wasm || classic

package presenter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// DilotiCuiPresenter renders the Diloti CUI view.
type DilotiCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *DilotiCuiPresenter) Output(g interfaces.DilotiGame, lastErr error) string {
	return buildCuiOutput(i18n.T("diloti.helpTitle"), func(b *strings.Builder) {
		p.writeHeader(b, g)
		p.writeSeats(b, g)
		b.WriteString("----------\n")
		p.writeTable(b, g)
		cuiErrorBlock(b, lastErr)
		if g.GetGameEndFlag() {
			p.writeGameEnd(b, g)
			return
		}
		if g.GetPhase() == domain.DilotiPhaseRoundEnd {
			p.writeRoundResult(b, g)
			return
		}
		p.writeMoves(b, g)
	})
}

// writeHeader は局と山の残りを書く。
func (p *DilotiCuiPresenter) writeHeader(b *strings.Builder, g interfaces.DilotiGame) {
	b.WriteString(i18n.Tf("diloti.round",
		"n", strconv.Itoa(g.GetRoundNumber()),
		"target", strconv.Itoa(g.GetConfig().TargetScore)) + "\n")
	b.WriteString(i18n.Tf("diloti.deck", "n", strconv.Itoa(g.GetDeckRemaining())) + "\n")
}

// writeSeats は席ごとの行と人間の手札を書く。
func (p *DilotiCuiPresenter) writeSeats(b *strings.Builder, g interfaces.DilotiGame) {
	dealer := g.GetDealerIdx()
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		role := i18n.T("diloti.roleNonDealer")
		if i == dealer {
			role = i18n.T("diloti.roleDealer")
		}
		b.WriteString(i18n.Tf("diloti.seat",
			"name", dilotiSeatName(player.GetIsHuman(), i),
			"role", role,
			"captured", strconv.Itoa(len(player.GetCaptured())),
			"xeri", strconv.Itoa(player.GetXeri()),
			"score", strconv.Itoa(player.GetScore())) + "\n")
		if player.GetIsHuman() && player.GetCardsSize() > 0 {
			b.WriteString(dilotiIndexedHand(player) + "\n")
		}
	}
}

// dilotiSeatName は席の呼び名を返す。
func dilotiSeatName(isHuman bool, idx int) string {
	if isHuman {
		return i18n.T("diloti.you")
	}
	return i18n.Tf("diloti.cpu", "n", strconv.Itoa(idx))
}

// dilotiIndexedHand は番号付きの手札を返す。
func dilotiIndexedHand(player *domain.DilotiPlayer) string {
	parts := make([]string, 0, player.GetCardsSize())
	for i := 0; i < player.GetCardsSize(); i++ {
		parts = append(parts, fmt.Sprintf("[%d]%s", i, cuiCardStrEmojiRank(player.GetCard(i))))
	}
	return strings.Join(parts, "  ")
}

// writeTable は場札と宣言を書く。
//
// **場札にも宣言にも番号が要る。** 取る対象はこの番号で指すので、無いと
// 組合せ捕獲も宣言の捕獲も打てない。
func (p *DilotiCuiPresenter) writeTable(b *strings.Builder, g interfaces.DilotiGame) {
	table := g.GetTable()
	parts := make([]string, 0, len(table))
	for i, c := range table {
		parts = append(parts, fmt.Sprintf("[%d]%s", i, cuiCardStrEmojiRank(c)))
	}
	if len(parts) == 0 {
		parts = append(parts, i18n.T("diloti.emptyTable"))
	}
	b.WriteString(i18n.Tf("diloti.table", "cards", strings.Join(parts, "  ")) + "\n")

	for i, d := range g.GetDeclarations() {
		if d == nil {
			continue
		}
		groups := make([]string, 0, len(d.Groups))
		for _, grp := range d.Groups {
			cards := make([]string, 0, len(grp))
			for _, c := range grp {
				cards = append(cards, cuiCardStrEmojiRank(c))
			}
			groups = append(groups, strings.Join(cards, "+"))
		}
		kind := i18n.T("diloti.declPlain")
		if d.IsGroup {
			kind = i18n.T("diloti.declGroup")
		}
		b.WriteString(color.Yellow(i18n.Tf("diloti.declaration",
			"idx", strconv.Itoa(i),
			"value", strconv.Itoa(d.Value),
			"kind", kind,
			"owner", dilotiSeatName(g.GetPlayer(d.OwnerIdx) != nil && g.GetPlayer(d.OwnerIdx).GetIsHuman(), d.OwnerIdx),
			"cards", strings.Join(groups, " | "))) + "\n")
	}
}

// writeMoves は手番と打てる手を書く。
func (p *DilotiCuiPresenter) writeMoves(b *strings.Builder, g interfaces.DilotiGame) {
	cur := g.GetCurrentPlayerIdx()
	player := g.GetPlayer(cur)
	if player == nil {
		return
	}
	b.WriteString(i18n.Tf("diloti.turn", "name", dilotiSeatName(player.GetIsHuman(), cur)) + "\n")
	if !g.IsHumanTurn() {
		return
	}
	human := cur
	for i := 0; i < player.GetCardsSize(); i++ {
		card := cuiCardStrEmojiRank(player.GetCard(i))
		if takes := g.GetTakeOptions(human, i); len(takes) > 0 {
			b.WriteString("  " + i18n.Tf("diloti.takeOption",
				"idx", strconv.Itoa(i), "card", card,
				"groups", dilotiFormatTakes(takes)) + "\n")
		}
		if cands := g.GetDeclareOptions(human, i); len(cands) > 0 {
			b.WriteString("  " + i18n.Tf("diloti.declareOption",
				"idx", strconv.Itoa(i), "card", card,
				"groups", dilotiFormatDeclCandidates(cands)) + "\n")
		}
	}
	b.WriteString(i18n.T("diloti.commandHint") + "\n")
}

// dilotiFormatTakes は取り手を "(0 1) (d0)" の形に整える。
func dilotiFormatTakes(takes []domain.DilotiTake) string {
	parts := make([]string, 0, len(takes))
	for _, t := range takes {
		tokens := make([]string, 0, len(t.TableIdxs)+len(t.DeclIdxs))
		for _, i := range t.TableIdxs {
			tokens = append(tokens, strconv.Itoa(i))
		}
		for _, i := range t.DeclIdxs {
			tokens = append(tokens, "d"+strconv.Itoa(i))
		}
		parts = append(parts, "("+strings.Join(tokens, " ")+")")
	}
	return strings.Join(parts, " ")
}

// dilotiFormatDeclCandidates は宣言候補を "8:(0 1)" の形に整える。
func dilotiFormatDeclCandidates(cands []domain.DilotiDeclCandidate) string {
	parts := make([]string, 0, len(cands))
	for _, c := range cands {
		tokens := make([]string, 0, len(c.TableIdxs))
		for _, i := range c.TableIdxs {
			tokens = append(tokens, strconv.Itoa(i))
		}
		parts = append(parts, strconv.Itoa(c.Value)+":("+strings.Join(tokens, " ")+")")
	}
	return strings.Join(parts, " ")
}

// writeRoundResult は局の集計を書く。
func (p *DilotiCuiPresenter) writeRoundResult(b *strings.Builder, g interfaces.DilotiGame) {
	res := g.GetLastResult()
	if res == nil {
		return
	}
	b.WriteString(i18n.T("diloti.resultTitle") + "\n")
	for _, l := range res.Lines {
		points := make([]string, 0, len(l.Points))
		for _, pt := range l.Points {
			points = append(points, strconv.Itoa(pt))
		}
		b.WriteString("  " + i18n.Tf("diloti.resultLine",
			"key", i18n.T("diloti.score."+l.Key),
			"points", strings.Join(points, " - ")) + "\n")
	}
	b.WriteString("  " + i18n.Tf("diloti.resultTotal",
		"a", strconv.Itoa(res.Totals[0]), "b", strconv.Itoa(res.Totals[1])) + "\n")
	b.WriteString(i18n.T("diloti.nextRoundHint") + "\n")
}

// writeGameEnd は終局の行を書く。
func (p *DilotiCuiPresenter) writeGameEnd(b *strings.Builder, g interfaces.DilotiGame) {
	winner := g.GetWinnerIdx()
	if winner < 0 {
		return
	}
	player := g.GetPlayer(winner)
	b.WriteString(color.Green(i18n.Tf("diloti.winner",
		"name", dilotiSeatName(player != nil && player.GetIsHuman(), winner))) + "\n")
}

// HintOutput renders the recommended move.
func (p *DilotiCuiPresenter) HintOutput(g interfaces.DilotiGame) string {
	return buildCuiOutput(i18n.T("diloti.helpTitle"), func(b *strings.Builder) {
		hint := g.GetHint()
		if hint == nil || hint.Move.HandIdx < 0 {
			b.WriteString(i18n.T("diloti.noHint") + "\n")
			return
		}
		human := g.GetCurrentPlayerIdx()
		player := g.GetPlayer(human)
		card := "??"
		if player != nil && hint.Move.HandIdx < player.GetCardsSize() {
			card = cuiCardStrEmojiRank(player.GetCard(hint.Move.HandIdx))
		}
		b.WriteString(i18n.Tf("diloti.hintCard",
			"idx", strconv.Itoa(hint.Move.HandIdx),
			"card", card,
			"action", i18n.T("diloti.action."+hint.Move.Action)) + "\n")
		b.WriteString(i18n.Tf("diloti.hintReason",
			"reason", i18n.T("diloti.reason."+hint.Move.Reason)) + "\n")
	})
}

// ActionLogOutput renders the action log.
func (p *DilotiCuiPresenter) ActionLogOutput(g interfaces.DilotiGame) string {
	return actionLogOutputTextForSeats[*domain.DilotiPlayer](g)
}
