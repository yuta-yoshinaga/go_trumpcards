//go:build !js || !wasm || extra

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

// koenigrufenCuiCardStr タロック札の CUI 表示文字列 (切り札 "T{n}"、スキュース "Sküs"、
// スート札 "♠5" 等)。標準の cuiCardStr は design 5/6 を扱えないためローカルに用意する。
func koenigrufenCuiCardStr(c *domain.Card) string {
	if c == nil {
		return "??"
	}
	switch c.GetDesign() {
	case domain.KoenigrufenSkusDesign:
		return color.Green("Sküs")
	case domain.KoenigrufenTrumpDesign:
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
		s := g + koenigrufenRankLabel(c.GetValue())
		if isRedSuit(c.GetDesign()) {
			return color.Red(s)
		}
		return s
	}
}

// koenigrufenIndexedHand 人間手札をインデックス付きで表示する。
func koenigrufenIndexedHand(p *domain.KoenigrufenPlayer) string {
	parts := make([]string, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		parts[i] = fmt.Sprintf("[%d]%s", i, koenigrufenCuiCardStr(p.GetCard(i)))
	}
	return strings.Join(parts, "  ")
}

// koenigrufenBidLabel 入札の i18n ラベルを返す。
func koenigrufenBidLabel(bid domain.KoenigrufenBid) string {
	if bid == domain.KoenigrufenBidRufer {
		return i18n.T("koenigrufen.bidRufer")
	}
	return i18n.T("koenigrufen.bidNone")
}

// koenigrufenOutcomeLabel ディール結果の i18n ラベルを返す。
func koenigrufenOutcomeLabel(o domain.KoenigrufenOutcome) string {
	switch o {
	case domain.KoenigrufenOutcomeWin:
		return i18n.T("koenigrufen.outcomeWin")
	case domain.KoenigrufenOutcomeLoss:
		return i18n.T("koenigrufen.outcomeLoss")
	default:
		return i18n.T("koenigrufen.outcomeNone")
	}
}

// koenigrufenRoleLabel プレイヤーの役割ラベルを返す。パートナーは公開済みのときのみ明示する。
func koenigrufenRoleLabel(g interfaces.KoenigrufenGame, idx int) string {
	if idx == g.GetDeclarerIdx() {
		return i18n.T("koenigrufen.roleDeclarer")
	}
	if g.GetPartnerRevealed() && g.GetPartnerIdx() == idx {
		return i18n.T("koenigrufen.rolePartner")
	}
	return i18n.T("koenigrufen.roleOpponent")
}

// koenigrufenPlayerStr プレイヤー 1 行分の状態文字列を返す。
func koenigrufenPlayerStr(g interfaces.KoenigrufenGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	scores := g.GetPlayerScores()
	var b strings.Builder
	b.WriteString(i18n.Tf("koenigrufen.playerLine",
		"name", cuiPlayerName(player, idx),
		"role", koenigrufenRoleLabel(g, idx),
		"cards", strconv.Itoa(player.GetCardsSize()),
		"score", strconv.Itoa(scores[idx]),
		"tricks", strconv.Itoa(player.GetTrickCount()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(koenigrufenIndexedHand(player) + "\n")
	}
	return b.String()
}

// KoenigrufenCuiPresenter renders the Königrufen CUI view.
type KoenigrufenCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *KoenigrufenCuiPresenter) Output(g interfaces.KoenigrufenGame, lastErr error) string {
	return buildCuiOutput(i18n.T("koenigrufen.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("koenigrufen.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber()),
			"contract", koenigrufenBidLabel(g.GetContract())) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(koenigrufenPlayerStr(g, i))
		}

		b.WriteString("----------\n")

		cuiTrickBlock(b, g.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return koenigrufenCuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winner := g.GetWinnerPlayer()
			var winnerStr string
			if winner >= 0 {
				winnerStr = cuiPlayerName(g.GetPlayer(winner), winner)
			}
			b.WriteString(color.Green(i18n.Tf("koenigrufen.gameEnd", "name", winnerStr)) + "\n")
			return
		}
		p.writePrompt(b, g)
	})
}

// writePrompt 現在のフェーズに応じたプロンプトを書き込む。
func (p *KoenigrufenCuiPresenter) writePrompt(b *strings.Builder, g interfaces.KoenigrufenGame) {
	switch g.GetPhase() {
	case domain.KoenigrufenPhaseBid:
		b.WriteString(i18n.Tf("koenigrufen.promptBid",
			"name", cuiPlayerName(g.GetPlayer(g.GetBidPlayerIdx()), g.GetBidPlayerIdx()),
			"high", koenigrufenBidLabel(g.GetHighestBid())) + "\n")
		b.WriteString(i18n.T("koenigrufen.promptBidHelp") + "\n")
	case domain.KoenigrufenPhaseCall:
		b.WriteString(i18n.Tf("koenigrufen.promptCall",
			"name", cuiPlayerName(g.GetPlayer(g.GetDeclarerIdx()), g.GetDeclarerIdx())) + "\n")
		b.WriteString(i18n.T("koenigrufen.promptCallHelp") + "\n")
	case domain.KoenigrufenPhaseTalon:
		b.WriteString(i18n.Tf("koenigrufen.promptTalon",
			"name", cuiPlayerName(g.GetPlayer(g.GetDeclarerIdx()), g.GetDeclarerIdx())) + "\n")
		b.WriteString(i18n.T("koenigrufen.promptTalonHelp") + "\n")
		// The web UI greys out cards that can't be buried; on the CLI, state the
		// fill constraints (Kings and the trull honours are never buriable, plain
		// trumps only when short of plain-suit cards) so the choice isn't guesswork.
		b.WriteString(i18n.T("koenigrufen.promptTalonLegend") + "\n")
	case domain.KoenigrufenPhasePlay:
		currentIdx := g.GetCurrentPlayerIdx()
		b.WriteString(i18n.Tf("koenigrufen.promptPlay",
			"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
		b.WriteString(i18n.T("koenigrufen.promptPlayHelp") + "\n")
	case domain.KoenigrufenPhaseTrickEnd:
		b.WriteString(i18n.T("koenigrufen.promptTrickEnd") + "\n")
		b.WriteString(i18n.T("koenigrufen.promptTrickEndHelp") + "\n")
	case domain.KoenigrufenPhaseRoundEnd:
		b.WriteString(i18n.Tf("koenigrufen.promptRoundEnd",
			"declarer", cuiPlayerName(g.GetPlayer(g.GetDeclarerIdx()), g.GetDeclarerIdx()),
			"outcome", koenigrufenOutcomeLabel(g.GetOutcome())) + "\n")
		b.WriteString(i18n.T("koenigrufen.promptRoundEndHelp") + "\n")
	}
}

// HintOutput emits the current Königrufen hint.
func (p *KoenigrufenCuiPresenter) HintOutput(g interfaces.KoenigrufenGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("koenigrufen.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, koenigrufenHintReasonKeys)
	if len(hint.CardIndices) > 0 {
		playerIdx := g.GetCurrentPlayerIdx()
		if g.GetPhase() == domain.KoenigrufenPhaseTalon {
			playerIdx = g.GetDeclarerIdx()
		}
		player := g.GetPlayer(playerIdx)
		cards := make([]string, len(hint.CardIndices))
		for i, idx := range hint.CardIndices {
			if player != nil && idx >= 0 && idx < player.GetCardsSize() {
				cards[i] = "[" + strconv.Itoa(idx) + "]" + koenigrufenCuiCardStr(player.GetCard(idx))
			} else {
				cards[i] = strconv.Itoa(idx)
			}
		}
		return color.Yellow(i18n.Tf("koenigrufen.hintCard",
			"cards", strings.Join(cards, ", "),
			"reason", reason)) + "\n"
	}
	if hint.CallSuit != nil {
		return color.Yellow(i18n.Tf("koenigrufen.hintCard",
			// **スート名に直す。**呼び王のスート番号 (1..4) はカードの design と
			// 同じ値なので、他ゲーム (Cinch/King/Watten) と同じく名前で出す。
			// `callking <1-4>` の入力構文は数値のまま (#4858)。
			"cards", i18n.Tf("koenigrufen.hintCallSuit", "suit", cuiSuitName(*hint.CallSuit)),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("koenigrufen.hintCard", "cards", "-", "reason", reason)) + "\n"
}

// koenigrufenHintReasonKeys maps hint-reason identifiers to i18n keys.
var koenigrufenHintReasonKeys = map[string]string{
	"bid_take":     "koenigrufen.hintReasonBidTake",
	"bid_pass":     "koenigrufen.hintReasonBidPass",
	"call_king":    "koenigrufen.hintReasonCallKing",
	"discard_weak": "koenigrufen.hintReasonDiscardWeak",
	"lead_high":    "koenigrufen.hintReasonLeadHigh",
	"lead_low":     "koenigrufen.hintReasonLeadLow",
	"follow_win":   "koenigrufen.hintReasonFollowWin",
	"follow_duck":  "koenigrufen.hintReasonFollowDuck",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *KoenigrufenCuiPresenter) ActionLogOutput(g interfaces.KoenigrufenGame) string {
	return actionLogOutputText(g)
}
