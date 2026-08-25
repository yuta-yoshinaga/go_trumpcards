//go:build !js || !wasm || extra3

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

// CirullaCuiPresenter renders the Cirulla CUI view.
type CirullaCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *CirullaCuiPresenter) Output(g interfaces.CirullaGame, lastErr error) string {
	return buildCuiOutput(i18n.T("cirulla.helpTitle"), func(b *strings.Builder) {
		p.writeHeader(b, g)
		p.writeSeats(b, g)
		b.WriteString("----------\n")
		p.writeTable(b, g)
		cuiErrorBlock(b, lastErr)
		if g.GetGameEndFlag() {
			p.writeGameEnd(b, g)
			return
		}
		p.writePrompt(b, g)
	})
}

// writeHeader はラウンドと山の残りを書く。
func (p *CirullaCuiPresenter) writeHeader(b *strings.Builder, g interfaces.CirullaGame) {
	b.WriteString(i18n.Tf("cirulla.round",
		"n", strconv.Itoa(g.GetRoundNumber()),
		"target", strconv.Itoa(g.GetConfig().TargetScore)) + "\n")
	b.WriteString(i18n.Tf("cirulla.deck", "n", strconv.Itoa(g.GetDeckRemaining())) + "\n")
}

// writeSeats は席ごとの行と人間の手札を書く。
func (p *CirullaCuiPresenter) writeSeats(b *strings.Builder, g interfaces.CirullaGame) {
	bonuses := g.GetLastBonus()
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		role := i18n.T("cirulla.rolePlayer")
		if i == g.GetDealerIdx() {
			role = i18n.T("cirulla.roleDealer")
		}
		denari := 0
		for _, card := range player.GetCaptured() {
			if domain.CirullaIsDenari(card) {
				denari++
			}
		}
		b.WriteString(i18n.Tf("cirulla.playerLine",
			"name", cuiPlayerName(player, i),
			"role", role,
			"captured", strconv.Itoa(len(player.GetCaptured())),
			"denari", strconv.Itoa(denari),
			"scope", strconv.Itoa(player.GetScope()),
			"score", strconv.Itoa(player.GetScore())) + "\n")
		// **配札ボーナスは出た瞬間に見せる。** 集計まで伏せると、なぜ点が
		// 動いたのか分からない。
		if i < len(bonuses) && bonuses[i] != "" {
			b.WriteString(color.Yellow(i18n.Tf("cirulla.bonusLine",
				"name", i18n.T("cirulla.bonus."+bonuses[i]))) + "\n")
		}
		if player.GetIsHuman() && player.GetCardsSize() > 0 {
			b.WriteString(cirullaIndexedHand(player) + "\n")
		}
	}
}

// cirullaIndexedHand は人間の手札を番号付きで返す。
func cirullaIndexedHand(p *domain.CirullaPlayer) string {
	parts := make([]string, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		parts[i] = fmt.Sprintf("[%d]%s", i, cuiCardStrEmojiRank(p.GetCard(i)))
	}
	return strings.Join(parts, "  ")
}

// writeTable は場札を番号付きで書く。
//
// **場札には番号が要る。** 取る札はこの番号で指すので、出さないと組合せ捕獲が
// 打てない。
func (p *CirullaCuiPresenter) writeTable(b *strings.Builder, g interfaces.CirullaGame) {
	table := g.GetTable()
	if len(table) == 0 {
		b.WriteString(i18n.T("cirulla.tableEmpty") + "\n")
		return
	}
	parts := make([]string, 0, len(table))
	for i, card := range table {
		parts = append(parts, fmt.Sprintf("[%d]%s", i, cuiCardStrEmojiRank(card)))
	}
	b.WriteString(i18n.Tf("cirulla.table", "cards", strings.Join(parts, "  ")) + "\n")
}

// writeGameEnd は終局の行を書く。
func (p *CirullaCuiPresenter) writeGameEnd(b *strings.Builder, g interfaces.CirullaGame) {
	idx := g.GetWinnerIdx()
	b.WriteString(color.Green(i18n.Tf("cirulla.gameEnd",
		"name", cuiPlayerName(g.GetPlayer(idx), idx))) + "\n")
}

// writePrompt はフェーズに応じた案内を書く。
func (p *CirullaCuiPresenter) writePrompt(b *strings.Builder, g interfaces.CirullaGame) {
	switch g.GetPhase() {
	case domain.CirullaPhasePlay:
		idx := g.GetCurrentPlayerIdx()
		b.WriteString(i18n.Tf("cirulla.promptPlay",
			"name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
		player := g.GetPlayer(idx)
		if player == nil || !player.GetIsHuman() {
			return
		}
		p.writeCaptureOptions(b, g, idx, player)
		b.WriteString(i18n.T("cirulla.promptPlayHelp") + "\n")
	case domain.CirullaPhaseRoundEnd:
		p.writeRoundEnd(b, g)
	}
}

// writeCaptureOptions は手札ごとに取れる場札の組を並べる。
//
// **合計 15 と合計一致とアッソの総取りが混ざる。** 何が取れるのかを出さないと、
// 端末からは総当たりで探すことになる。
func (p *CirullaCuiPresenter) writeCaptureOptions(b *strings.Builder, g interfaces.CirullaGame, idx int, player *domain.CirullaPlayer) {
	any := false
	for i := 0; i < player.GetCardsSize(); i++ {
		groups := g.GetCaptureOptions(idx, i)
		if len(groups) == 0 {
			continue
		}
		any = true
		parts := make([]string, 0, len(groups))
		for _, group := range groups {
			nums := make([]string, 0, len(group))
			for _, t := range group {
				nums = append(nums, strconv.Itoa(t))
			}
			parts = append(parts, "("+strings.Join(nums, " ")+")")
		}
		b.WriteString(i18n.Tf("cirulla.captureOption",
			"idx", strconv.Itoa(i),
			"card", cuiCardStrEmojiRank(player.GetCard(i)),
			"groups", strings.Join(parts, " ")) + "\n")
	}
	if !any {
		b.WriteString(i18n.T("cirulla.noCapture") + "\n")
	}
}

// writeRoundEnd はラウンドの集計を項目別に書く。
func (p *CirullaCuiPresenter) writeRoundEnd(b *strings.Builder, g interfaces.CirullaGame) {
	if res := g.GetLastResult(); res != nil {
		for _, line := range res.Lines {
			if line.Points[0] == 0 && line.Points[1] == 0 {
				continue
			}
			b.WriteString(i18n.Tf("cirulla.scoreLine",
				"name", i18n.T("cirulla.score."+line.Key),
				"a", strconv.Itoa(line.Points[0]),
				"b", strconv.Itoa(line.Points[1])) + "\n")
		}
		b.WriteString(i18n.Tf("cirulla.roundTotal",
			"a", strconv.Itoa(res.Totals[0]), "b", strconv.Itoa(res.Totals[1])) + "\n")
	}
	b.WriteString(i18n.T("cirulla.promptRoundEndHelp") + "\n")
}

// HintOutput emits the current hint.
func (p *CirullaCuiPresenter) HintOutput(g interfaces.CirullaGame) string {
	hint := g.GetHint()
	if hint == nil || hint.HandIdx < 0 {
		return i18n.T("cirulla.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, cirullaHintReasonKeys)
	player := g.GetPlayer(g.GetCurrentPlayerIdx())
	card := "-"
	if player != nil && hint.HandIdx < player.GetCardsSize() {
		card = "[" + strconv.Itoa(hint.HandIdx) + "]" + cuiCardStrEmojiRank(player.GetCard(hint.HandIdx))
	}
	target := i18n.T("cirulla.hintLayOff")
	if len(hint.CaptureIdxs) > 0 {
		nums := make([]string, 0, len(hint.CaptureIdxs))
		for _, t := range hint.CaptureIdxs {
			nums = append(nums, strconv.Itoa(t))
		}
		target = i18n.Tf("cirulla.hintTake", "cards", strings.Join(nums, " "))
	}
	return color.Yellow(i18n.Tf("cirulla.hintCard",
		"card", card, "target", target, "reason", reason)) + "\n"
}

// cirullaHintReasonKeys はヒント理由と i18n キーの対応。
var cirullaHintReasonKeys = map[string]string{
	"capture":    "cirulla.hintReasonCapture",
	"sweep":      "cirulla.hintReasonSweep",
	"lay_off":    "cirulla.hintReasonLayOff",
	"next_round": "cirulla.hintReasonNextRound",
	"none":       "cirulla.hintNone",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *CirullaCuiPresenter) ActionLogOutput(g interfaces.CirullaGame) string {
	return actionLogOutputTextForSeats[*domain.CirullaPlayer](g)
}
