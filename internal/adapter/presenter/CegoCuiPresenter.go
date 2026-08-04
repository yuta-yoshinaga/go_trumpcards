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

// cegoCuiCardStr タロック札の CUI 表示文字列 (切り札 "T{n}"、スキュース "Sküs"、スート札 "♠5" 等)。
// 標準の cuiCardStr は design 5/6 を扱えないためローカルに用意する。
func cegoCuiCardStr(c *domain.Card) string {
	if c == nil {
		return "??"
	}
	switch c.GetDesign() {
	case domain.CegoSkusDesign:
		return color.Green("Sküs")
	case domain.CegoTrumpDesign:
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
		s := g + cegoRankLabel(c.GetValue())
		if isRedSuit(c.GetDesign()) {
			return color.Red(s)
		}
		return s
	}
}

// cegoIndexedHand 人間手札をインデックス付きで表示する。
func cegoIndexedHand(p *domain.CegoPlayer) string {
	parts := make([]string, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		parts[i] = fmt.Sprintf("[%d]%s", i, cegoCuiCardStr(p.GetCard(i)))
	}
	return strings.Join(parts, "  ")
}

// cegoBidLabel 入札の i18n ラベルを返す。
func cegoBidLabel(bid domain.CegoBid) string {
	if bid == domain.CegoBidPlay {
		return i18n.T("cego.bidPlay")
	}
	return i18n.T("cego.bidNone")
}

// cegoContractLabel コントラクトの i18n ラベルを返す。
func cegoContractLabel(ct domain.CegoContract) string {
	switch ct {
	case domain.CegoContractCego:
		return i18n.T("cego.contractCego")
	case domain.CegoContractHandspiel:
		return i18n.T("cego.contractHandspiel")
	default:
		return i18n.T("cego.contractNone")
	}
}

// cegoOutcomeLabel ディール結果の i18n ラベルを返す。
func cegoOutcomeLabel(o domain.CegoOutcome) string {
	switch o {
	case domain.CegoOutcomeWin:
		return i18n.T("cego.outcomeWin")
	case domain.CegoOutcomeLoss:
		return i18n.T("cego.outcomeLoss")
	default:
		return i18n.T("cego.outcomeNone")
	}
}

// cegoRoleLabel プレイヤーの役割ラベルを返す。
func cegoRoleLabel(g interfaces.CegoGame, idx int) string {
	if idx == g.GetDeclarerIdx() {
		return i18n.T("cego.roleDeclarer")
	}
	return i18n.T("cego.roleOpponent")
}

// cegoPlayerStr プレイヤー 1 行分の状態文字列を返す。
func cegoPlayerStr(g interfaces.CegoGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	scores := g.GetPlayerScores()
	var b strings.Builder
	b.WriteString(i18n.Tf("cego.playerLine",
		"name", cuiPlayerName(player, idx),
		"role", cegoRoleLabel(g, idx),
		"cards", strconv.Itoa(player.GetCardsSize()),
		"score", strconv.Itoa(scores[idx]),
		"tricks", strconv.Itoa(player.GetTrickCount()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cegoIndexedHand(player) + "\n")
	}
	return b.String()
}

// CegoCuiPresenter renders the Cego CUI view.
type CegoCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *CegoCuiPresenter) Output(g interfaces.CegoGame, lastErr error) string {
	return buildCuiOutput(i18n.T("cego.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("cego.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber()),
			"contract", cegoContractLabel(g.GetContractType())) + "\n")

		// Once the declarer is set, surface who it is and the contract at a glance
		// (the declarer fights the whole field); before that, it isn't shown.
		if declarer := g.GetDeclarerIdx(); declarer >= 0 {
			b.WriteString(i18n.Tf("cego.declarerLine",
				"name", cuiPlayerName(g.GetPlayer(declarer), declarer),
				"contract", cegoContractLabel(g.GetContractType())) + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(cegoPlayerStr(g, i))
		}

		b.WriteString("----------\n")

		cuiTrickBlock(b, g.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cegoCuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winner := g.GetWinnerPlayer()
			var winnerStr string
			if winner >= 0 {
				winnerStr = cuiPlayerName(g.GetPlayer(winner), winner)
			}
			b.WriteString(color.Green(i18n.Tf("cego.gameEnd", "name", winnerStr)) + "\n")
			return
		}
		p.writePrompt(b, g)
	})
}

// writePrompt 現在のフェーズに応じたプロンプトを書き込む。
func (p *CegoCuiPresenter) writePrompt(b *strings.Builder, g interfaces.CegoGame) {
	switch g.GetPhase() {
	case domain.CegoPhaseBid:
		b.WriteString(i18n.Tf("cego.promptBid",
			"name", cuiPlayerName(g.GetPlayer(g.GetBidPlayerIdx()), g.GetBidPlayerIdx()),
			"high", cegoBidLabel(g.GetHighestBid())) + "\n")
		b.WriteString(i18n.T("cego.promptBidHelp") + "\n")
	case domain.CegoPhaseContract:
		b.WriteString(i18n.Tf("cego.promptContract",
			"name", cuiPlayerName(g.GetPlayer(g.GetDeclarerIdx()), g.GetDeclarerIdx())) + "\n")
		b.WriteString(i18n.T("cego.promptContractHelp") + "\n")
		// **コマンド名だけでは選べない。**Web には契約ごとのリスク・リターンを
		// 説明する箱があるのに、CUI は構文リマインダーしか出していなかった (#4931)。
		b.WriteString(i18n.T("cego.contractExplainTitle") + "\n")
		b.WriteString(i18n.Tf("cego.contractCegoDesc",
			"count", strconv.Itoa(g.GetBlindCount())) + "\n")
		b.WriteString(i18n.T("cego.contractHandspielDesc") + "\n")

	case domain.CegoPhaseExchange:
		b.WriteString(i18n.Tf("cego.promptExchange",
			"name", cuiPlayerName(g.GetPlayer(g.GetDeclarerIdx()), g.GetDeclarerIdx())) + "\n")
		b.WriteString(i18n.T("cego.promptExchangeHelp") + "\n")
	case domain.CegoPhasePlay:
		currentIdx := g.GetCurrentPlayerIdx()
		b.WriteString(i18n.Tf("cego.promptPlay",
			"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
		b.WriteString(i18n.T("cego.promptPlayHelp") + "\n")
	case domain.CegoPhaseTrickEnd:
		b.WriteString(i18n.T("cego.promptTrickEnd") + "\n")
		b.WriteString(i18n.T("cego.promptTrickEndHelp") + "\n")
	case domain.CegoPhaseRoundEnd:
		b.WriteString(i18n.Tf("cego.promptRoundEnd",
			"declarer", cuiPlayerName(g.GetPlayer(g.GetDeclarerIdx()), g.GetDeclarerIdx()),
			"outcome", cegoOutcomeLabel(g.GetOutcome())) + "\n")
		b.WriteString(i18n.T("cego.promptRoundEndHelp") + "\n")
	}
}

// HintOutput emits the current Cego hint.
func (p *CegoCuiPresenter) HintOutput(g interfaces.CegoGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("cego.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, cegoHintReasonKeys)
	if len(hint.CardIndices) > 0 {
		playerIdx := g.GetCurrentPlayerIdx()
		if g.GetPhase() == domain.CegoPhaseExchange {
			playerIdx = g.GetDeclarerIdx()
		}
		player := g.GetPlayer(playerIdx)
		cards := make([]string, len(hint.CardIndices))
		for i, idx := range hint.CardIndices {
			if player != nil && idx >= 0 && idx < player.GetCardsSize() {
				cards[i] = "[" + strconv.Itoa(idx) + "]" + cegoCuiCardStr(player.GetCard(idx))
			} else {
				cards[i] = strconv.Itoa(idx)
			}
		}
		return color.Yellow(i18n.Tf("cego.hintCard",
			"cards", strings.Join(cards, ", "),
			"reason", reason)) + "\n"
	}
	if hint.Contract != nil {
		return color.Yellow(i18n.Tf("cego.hintCard",
			"cards", cegoContractLabel(domain.CegoContract(*hint.Contract)),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("cego.hintCard", "cards", "-", "reason", reason)) + "\n"
}

// cegoHintReasonKeys maps hint-reason identifiers to i18n keys.
var cegoHintReasonKeys = map[string]string{
	"bid_take":           "cego.hintReasonBidTake",
	"bid_pass":           "cego.hintReasonBidPass",
	"contract_cego":      "cego.hintReasonContractCego",
	"contract_handspiel": "cego.hintReasonContractHandspiel",
	"keep_best":          "cego.hintReasonKeepBest",
	"lead_high":          "cego.hintReasonLeadHigh",
	"lead_low":           "cego.hintReasonLeadLow",
	"follow_win":         "cego.hintReasonFollowWin",
	"follow_duck":        "cego.hintReasonFollowDuck",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *CegoCuiPresenter) ActionLogOutput(g interfaces.CegoGame) string {
	return actionLogOutputText(g)
}
