//go:build !js || !wasm || solo

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// aluetteCardStr renders one card as text.
//
// **リュエットは必ず呼び名を添える。**「♦3」とだけ出すと、それがデッキ最強の
// Monsieur なのか、ただの 3 なのかが手札から読み取れない。スートは強さに
// 関係しないので、名前が出ていない札は序列表どおりの数字比較で済む。
func aluetteCardStr(card *domain.Card) string {
	if card == nil {
		return "??"
	}
	base := cuiSuitName(card.GetDesign()) + " " + strconv.Itoa(card.GetValue())
	if name := domain.AluetteLuetteName(card); name != "" {
		return i18n.Tf("aluette.cardLuette", "card", base, "name", name)
	}
	return base
}

// aluetteIndexedHandStr renders the human hand as "[0]<card>  [1]<card>  ...".
func aluetteIndexedHandStr(player *domain.AluettePlayer) string {
	parts := make([]string, 0, player.GetCardsSize())
	for i := 0; i < player.GetCardsSize(); i++ {
		parts = append(parts, "["+strconv.Itoa(i)+"]"+aluetteCardStr(player.GetCard(i)))
	}
	return strings.Join(parts, "  ")
}

// aluettePlayerStr returns the display string for a single player.
func aluettePlayerStr(g interfaces.AluetteGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	role := ""
	if idx == g.GetDealerIdx() {
		role = " " + i18n.T("aluette.dealerBadge")
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("aluette.playerLine",
		"name", cuiPlayerName(player, idx),
		"team", strconv.Itoa(domain.AluetteTeamOf(idx)),
		"cards", strconv.Itoa(player.GetCardsSize()),
		"tricks", strconv.Itoa(player.GetTrickCount()),
	))
	b.WriteString(role + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(aluetteIndexedHandStr(player) + "\n")
	}
	return b.String()
}

// AluetteCuiPresenter renders the Aluette CUI view.
type AluetteCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *AluetteCuiPresenter) Output(g interfaces.AluetteGame, lastErr error) string {
	return buildCuiOutput(i18n.T("aluette.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("aluette.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber())) + "\n")

		scores := g.GetTeamScores()
		b.WriteString(i18n.Tf("aluette.teamLine",
			"team0", strconv.Itoa(scores[0]),
			"team1", strconv.Itoa(scores[1]),
			"target", strconv.Itoa(g.GetConfig().TargetPoints)) + "\n")

		// **序列表を毎回出す。**リュエット 6 枚を覚えていないと着手が選べない
		// のに、盤面のどこにも書いていなければ手札の意味が読めない。
		b.WriteString(p.luetteLegend() + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(aluettePlayerStr(g, i))
		}

		b.WriteString("----------\n")

		cuiTrickBlock(b, g.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return aluetteCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			p.writeGameEnd(b, g)
			return
		}
		p.writePrompt(b, g)
	})
}

// luetteLegend はリュエット 6 枚を強い順に 1 行で書く。
func (p *AluetteCuiPresenter) luetteLegend() string {
	table := domain.AluetteLuetteTable()
	parts := make([]string, 0, len(table))
	for _, l := range table {
		parts = append(parts, l.Name+"("+cuiSuitName(l.Design)+strconv.Itoa(l.Value)+")")
	}
	return i18n.Tf("aluette.luetteLegend", "list", strings.Join(parts, " > "))
}

// writeGameEnd マッチ終了のバナーを書く。
func (p *AluetteCuiPresenter) writeGameEnd(b *strings.Builder, g interfaces.AluetteGame) {
	winner := g.GetWinnerTeam()
	if winner < 0 {
		b.WriteString(color.Green(i18n.T("aluette.gameEndDraw")) + "\n")
		return
	}
	b.WriteString(color.Green(i18n.Tf("aluette.gameEnd", "team", strconv.Itoa(winner))) + "\n")
}

// writePrompt renders the phase-specific prompt block.
func (p *AluetteCuiPresenter) writePrompt(b *strings.Builder, g interfaces.AluetteGame) {
	switch g.GetPhase() {
	case domain.AluettePhasePlay:
		currentIdx := g.GetCurrentPlayerIdx()
		b.WriteString(i18n.Tf("aluette.promptPlay",
			"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
		b.WriteString(i18n.T("aluette.promptPlayHelp") + "\n")
	case domain.AluettePhaseTrickEnd:
		b.WriteString(i18n.T("aluette.promptTrickEnd") + "\n")
		b.WriteString(i18n.T("aluette.promptTrickEndHelp") + "\n")
	case domain.AluettePhaseRoundEnd:
		b.WriteString(i18n.T("aluette.promptRoundEnd") + "\n")
		p.writeRoundEndResult(b, g)
		b.WriteString(i18n.T("aluette.promptRoundEndHelp") + "\n")
	}
}

// writeRoundEndResult appends a per-seat trick tally.
func (p *AluetteCuiPresenter) writeRoundEndResult(b *strings.Builder, g interfaces.AluetteGame) {
	entries := make([]string, 0, g.GetPlayerCnt())
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		entries = append(entries, i18n.Tf("aluette.roundEndTrickEntry",
			"name", cuiPlayerName(player, i),
			"tricks", strconv.Itoa(player.GetTrickCount())))
	}
	b.WriteString(i18n.Tf("aluette.roundEndTricks", "list", strings.Join(entries, ", ")) + "\n")
}

// HintOutput emits the current Aluette hint.
func (p *AluetteCuiPresenter) HintOutput(g interfaces.AluetteGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("aluette.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, aluetteHintReasonKeys)
	if len(hint.CardIndices) > 0 {
		player := g.GetPlayer(g.GetCurrentPlayerIdx())
		cards := make([]string, len(hint.CardIndices))
		for i, idx := range hint.CardIndices {
			if player != nil {
				cards[i] = "[" + strconv.Itoa(idx) + "]" + aluetteCardStr(player.GetCard(idx))
			} else {
				cards[i] = strconv.Itoa(idx)
			}
		}
		return color.Yellow(i18n.Tf("aluette.hintCard",
			"cards", strings.Join(cards, ", "), "reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("aluette.hintCard", "cards", "-", "reason", reason)) + "\n"
}

// aluetteHintReasonKeys maps Aluette-specific hint-reason identifiers to i18n keys.
//
// **Aluette.playHintReason が返す 4 種と 1 対 1 で対応させる。**外れた理由は
// hintReasonStr の既定でキー文字列そのものが画面に出る (#4660 で実際に起きた)。
var aluetteHintReasonKeys = map[string]string{
	"lead_low":        "aluette.hintReasonLeadLow",
	"play_luette":     "aluette.hintReasonPlayLuette",
	"partner_winning": "aluette.hintReasonPartnerWinning",
	"follow_low":      "aluette.hintReasonFollowLow",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *AluetteCuiPresenter) ActionLogOutput(g interfaces.AluetteGame) string {
	return actionLogOutputText(g)
}
