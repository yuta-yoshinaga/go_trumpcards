//go:build !js || !wasm || extra4

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

// piedmonteseTarotCuiCardStr はタロット札の CUI 表示を返す (切り札 "T{n}"、
// Matto "MATTO"、スート札 "♠5" 等)。標準の cuiCardStr は design 5/6 を扱えない。
func piedmonteseTarotCuiCardStr(c *domain.Card) string {
	if c == nil {
		return "??"
	}
	switch c.GetDesign() {
	case domain.Tarot78ExcuseDesign:
		return color.Green("MATTO")
	case domain.Tarot78TrumpDesign:
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
		s := g + piedmonteseTarotRankLabel(c.GetValue())
		if isRedSuit(c.GetDesign()) {
			return color.Red(s)
		}
		return s
	}
}

// piedmonteseTarotRankLabel はスート札の位を返す (11..14 はコート札)。
func piedmonteseTarotRankLabel(v int) string {
	switch v {
	case 11:
		return "V" // Valet
	case 12:
		return "C" // Cavalier
	case 13:
		return "D" // Dame
	case domain.Tarot78KingValue:
		return "R" // Roi
	default:
		return strconv.Itoa(v)
	}
}

// piedmonteseTarotIndexedHand は人間の手札を番号付きで返す。
func piedmonteseTarotIndexedHand(p *domain.PiedmonteseTarotPlayer) string {
	parts := make([]string, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		parts[i] = fmt.Sprintf("[%d]%s", i, piedmonteseTarotCuiCardStr(p.GetCard(i)))
	}
	return strings.Join(parts, "  ")
}

// piedmonteseTarotOutcomeLabel はディール結果のラベルを返す。
func piedmonteseTarotOutcomeLabel(o domain.PiedmonteseTarotOutcome) string {
	switch o {
	case domain.PiedmonteseTarotOutcomeWin:
		return i18n.T("piedmontesetarot.outcomeWin")
	case domain.PiedmonteseTarotOutcomeLoss:
		return i18n.T("piedmontesetarot.outcomeLoss")
	default:
		return i18n.T("piedmontesetarot.outcomeNone")
	}
}

// piedmonteseTarotPlayerStr は 1 席ぶんの行を返す。
func piedmonteseTarotPlayerStr(g interfaces.PiedmonteseTarotGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	scores := g.GetPlayerScores()
	score := 0
	if idx < len(scores) {
		score = scores[idx]
	}
	role := i18n.T("piedmontesetarot.rolePlayer")
	if idx == g.GetDealerIdx() {
		role = i18n.T("piedmontesetarot.roleDealer")
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("piedmontesetarot.playerLine",
		"name", cuiPlayerName(player, idx),
		"role", role,
		"cards", strconv.Itoa(player.GetCardsSize()),
		"score", strconv.Itoa(score),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"points", domain.PiedmonteseTarotFormatThirds(g.GetCardThirds(idx)),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(piedmonteseTarotIndexedHand(player) + "\n")
	}
	return b.String()
}

// PiedmonteseTarotCuiPresenter renders the Tarocco Piemontese CUI view.
type PiedmonteseTarotCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *PiedmonteseTarotCuiPresenter) Output(g interfaces.PiedmonteseTarotGame, lastErr error) string {
	return buildCuiOutput(i18n.T("piedmontesetarot.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("piedmontesetarot.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber()),
			"tricks", strconv.Itoa(g.HandSize())) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(piedmonteseTarotPlayerStr(g, i))
		}

		b.WriteString("----------\n")

		cuiTrickBlock(b, g.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return piedmonteseTarotCuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winner := g.GetWinnerPlayer()
			var winnerStr string
			if winner >= 0 {
				winnerStr = cuiPlayerName(g.GetPlayer(winner), winner)
			}
			b.WriteString(color.Green(i18n.Tf("piedmontesetarot.gameEnd", "name", winnerStr)) + "\n")
			return
		}
		p.writePrompt(b, g)
	})
}

// writePrompt はフェーズに応じた案内を書く。
func (p *PiedmonteseTarotCuiPresenter) writePrompt(b *strings.Builder, g interfaces.PiedmonteseTarotGame) {
	switch g.GetPhase() {
	case domain.PiedmonteseTarotPhaseScarto:
		dealerIdx := g.GetDealerIdx()
		b.WriteString(i18n.Tf("piedmontesetarot.promptScarto",
			"name", cuiPlayerName(g.GetPlayer(dealerIdx), dealerIdx),
			"n", strconv.Itoa(g.TalonSize())) + "\n")
		b.WriteString(i18n.Tf("piedmontesetarot.promptScartoHelp", "n", strconv.Itoa(g.TalonSize())) + "\n")
		// **捨てられる札を並べて出す。** Web はボタンを無効化して示すが、CUI では
		// 一覧が無いと総当たりになる。
		if dealer := g.GetPlayer(dealerIdx); dealer != nil && dealer.GetIsHuman() {
			// **規則はドメインに訊く。** ここで作り直していたせいで切り札が常に
			// 除外され、ピップが足りない手では捨てられる札が実際より少なく
			// 見えていた (#6236)。
			var idxs []string
			for _, i := range g.GetDiscardableIndices() {
				idxs = append(idxs, "["+strconv.Itoa(i)+"]"+piedmonteseTarotCuiCardStr(dealer.GetCard(i)))
			}
			list := strings.Join(idxs, "  ")
			if list == "" {
				list = i18n.T("piedmontesetarot.discardableNone")
			}
			b.WriteString(i18n.Tf("piedmontesetarot.discardableList", "cards", list) + "\n")
			b.WriteString(i18n.T("piedmontesetarot.discardableLegend") + "\n")
		}
	case domain.PiedmonteseTarotPhasePlay:
		currentIdx := g.GetCurrentPlayerIdx()
		b.WriteString(i18n.Tf("piedmontesetarot.promptPlay",
			"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
		// **出せる札を並べて出す。** Web は `playableIndices` でボタンを絞るのに、
		// CUI にそれが無いと「フォロー義務・切り札義務」を総当たりで探すことになる。
		// 同じアクセサが既にあるので、配線しないほうが不自然。
		if player := g.GetPlayer(currentIdx); player != nil && player.GetIsHuman() {
			var parts []string
			for _, idx := range g.GetPlayableIndices(currentIdx) {
				if idx >= 0 && idx < player.GetCardsSize() {
					parts = append(parts, "["+strconv.Itoa(idx)+"]"+piedmonteseTarotCuiCardStr(player.GetCard(idx)))
				}
			}
			if len(parts) > 0 {
				b.WriteString(i18n.Tf("piedmontesetarot.playableList", "cards", strings.Join(parts, "  ")) + "\n")
			}
		}
		b.WriteString(i18n.T("piedmontesetarot.promptPlayHelp") + "\n")
	case domain.PiedmonteseTarotPhaseTrickEnd:
		b.WriteString(i18n.T("piedmontesetarot.promptTrickEnd") + "\n")
		b.WriteString(i18n.T("piedmontesetarot.promptTrickEndHelp") + "\n")
	case domain.PiedmonteseTarotPhaseRoundEnd:
		b.WriteString(i18n.Tf("piedmontesetarot.promptRoundEnd",
			"outcome", piedmonteseTarotOutcomeLabel(g.GetOutcome())) + "\n")
		piedmonteseTarotWriteBreakdown(b, g)
		b.WriteString(i18n.T("piedmontesetarot.promptRoundEndHelp") + "\n")
	}
}

// piedmonteseTarotWriteBreakdown は精算の内訳を書く。
//
// **式まで出す。** 精算は「席数 × 自分の取り分 − 卓の合計」のゼロサムなので、
// 取り分だけを出すと画面の数字がどこから来たのか検算できない。
func piedmonteseTarotWriteBreakdown(b *strings.Builder, g interfaces.PiedmonteseTarotGame) {
	n := g.GetPlayerCnt()
	if n == 0 {
		return
	}
	total := 0
	for i := 0; i < n; i++ {
		total += g.GetCardThirds(i)
	}
	b.WriteString(i18n.Tf("piedmontesetarot.roundEndTotal",
		"total", domain.PiedmonteseTarotFormatThirds(total)) + "\n")
	b.WriteString(i18n.T("piedmontesetarot.roundEndFormulaLine") + "\n")
	deal := g.GetDealScores()
	for i := 0; i < n; i++ {
		scaled := 0
		if i < len(deal) {
			scaled = deal[i]
		}
		b.WriteString(i18n.Tf("piedmontesetarot.roundEndEarned",
			"name", cuiPlayerName(g.GetPlayer(i), i),
			"points", domain.PiedmonteseTarotFormatThirds(g.GetCardThirds(i)),
			"scaled", piedmonteseTarotSignedStr(scaled)) + "\n")
	}
}

// piedmonteseTarotSignedStr は符号付きの整数表記を返す。
func piedmonteseTarotSignedStr(v int) string {
	if v > 0 {
		return "+" + strconv.Itoa(v)
	}
	return strconv.Itoa(v)
}

// HintOutput emits the current hint.
func (p *PiedmonteseTarotCuiPresenter) HintOutput(g interfaces.PiedmonteseTarotGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("piedmontesetarot.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, piedmonteseTarotHintReasonKeys)
	if len(hint.CardIndices) == 0 {
		return color.Yellow(i18n.Tf("piedmontesetarot.hintCard", "cards", "-", "reason", reason)) + "\n"
	}
	playerIdx := g.GetCurrentPlayerIdx()
	if g.GetPhase() == domain.PiedmonteseTarotPhaseScarto {
		playerIdx = g.GetDealerIdx()
	}
	player := g.GetPlayer(playerIdx)
	cards := make([]string, len(hint.CardIndices))
	for i, idx := range hint.CardIndices {
		if player != nil && idx >= 0 && idx < player.GetCardsSize() {
			cards[i] = "[" + strconv.Itoa(idx) + "]" + piedmonteseTarotCuiCardStr(player.GetCard(idx))
			continue
		}
		cards[i] = strconv.Itoa(idx)
	}
	return color.Yellow(i18n.Tf("piedmontesetarot.hintCard",
		"cards", strings.Join(cards, ", "),
		"reason", reason)) + "\n"
}

// piedmonteseTarotHintReasonKeys はヒント理由と i18n キーの対応。
var piedmonteseTarotHintReasonKeys = map[string]string{
	"scarto_weak": "piedmontesetarot.hintReasonScarto",
	"lead_low":    "piedmontesetarot.hintReasonLead",
	"follow_play": "piedmontesetarot.hintReasonFollow",
	"overtrump":   "piedmontesetarot.hintReasonOvertrump",
	"next_trick":  "piedmontesetarot.hintReasonNextTrick",
	"next_round":  "piedmontesetarot.hintReasonNextRound",
	"none":        "piedmontesetarot.hintNone",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *PiedmonteseTarotCuiPresenter) ActionLogOutput(g interfaces.PiedmonteseTarotGame) string {
	return actionLogOutputTextForSeats[*domain.PiedmonteseTarotPlayer](g)
}
