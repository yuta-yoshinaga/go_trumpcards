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

// frenchTarotCuiCardStr タロット札の CUI 表示文字列 (切り札 "T{n}"、エクスキューズ "EXC"、
// スート札 "♠5" 等)。標準の cuiCardStr は design 5/6 を扱えないためローカルに用意する。
func frenchTarotCuiCardStr(c *domain.Card) string {
	if c == nil {
		return "??"
	}
	switch c.GetDesign() {
	case domain.FrenchTarotExcuseDesign:
		return color.Green("EXC")
	case domain.FrenchTarotTrumpDesign:
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
		s := g + frenchTarotRankLabel(c.GetValue())
		if isRedSuit(c.GetDesign()) {
			return color.Red(s)
		}
		return s
	}
}

// frenchTarotIndexedHand 人間手札をインデックス付きで表示する。
func frenchTarotIndexedHand(p *domain.FrenchTarotPlayer) string {
	parts := make([]string, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		parts[i] = fmt.Sprintf("[%d]%s", i, frenchTarotCuiCardStr(p.GetCard(i)))
	}
	return strings.Join(parts, "  ")
}

// frenchTarotBidLabel 入札の i18n ラベルを返す。
func frenchTarotBidLabel(bid domain.FrenchTarotBid) string {
	switch bid {
	case domain.FrenchTarotBidPetite:
		return i18n.T("frenchtarot.bidPetite")
	case domain.FrenchTarotBidGarde:
		return i18n.T("frenchtarot.bidGarde")
	case domain.FrenchTarotBidGardeSans:
		return i18n.T("frenchtarot.bidGardeSans")
	case domain.FrenchTarotBidGardeContre:
		return i18n.T("frenchtarot.bidGardeContre")
	default:
		return i18n.T("frenchtarot.bidNone")
	}
}

// frenchTarotOutcomeLabel ディール結果の i18n ラベルを返す。
func frenchTarotOutcomeLabel(o domain.FrenchTarotOutcome) string {
	switch o {
	case domain.FrenchTarotOutcomeWin:
		return i18n.T("frenchtarot.outcomeWin")
	case domain.FrenchTarotOutcomeLoss:
		return i18n.T("frenchtarot.outcomeLoss")
	default:
		return i18n.T("frenchtarot.outcomeNone")
	}
}

// frenchTarotPlayerStr プレイヤー 1 行分の状態文字列を返す。
func frenchTarotPlayerStr(g interfaces.FrenchTarotGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	scores := g.GetPlayerScores()
	role := i18n.T("frenchtarot.roleDefender")
	if idx == g.GetDeclarerIdx() {
		role = i18n.T("frenchtarot.roleDeclarer")
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("frenchtarot.playerLine",
		"name", cuiPlayerName(player, idx),
		"role", role,
		"cards", strconv.Itoa(player.GetCardsSize()),
		"score", strconv.Itoa(scores[idx]),
		"tricks", strconv.Itoa(player.GetTrickCount()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(frenchTarotIndexedHand(player) + "\n")
	}
	return b.String()
}

// FrenchTarotCuiPresenter renders the French Tarot CUI view.
type FrenchTarotCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *FrenchTarotCuiPresenter) Output(g interfaces.FrenchTarotGame, lastErr error) string {
	return buildCuiOutput(i18n.T("frenchtarot.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("frenchtarot.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber()),
			"contract", frenchTarotBidLabel(g.GetContract())) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(frenchTarotPlayerStr(g, i))
		}

		b.WriteString("----------\n")

		cuiTrickBlock(b, g.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return frenchTarotCuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winner := g.GetWinnerPlayer()
			var winnerStr string
			if winner >= 0 {
				winnerStr = cuiPlayerName(g.GetPlayer(winner), winner)
			}
			b.WriteString(color.Green(i18n.Tf("frenchtarot.gameEnd", "name", winnerStr)) + "\n")
			return
		}
		p.writePrompt(b, g)
	})
}

// writePrompt 現在のフェーズに応じたプロンプトを書き込む。
func (p *FrenchTarotCuiPresenter) writePrompt(b *strings.Builder, g interfaces.FrenchTarotGame) {
	switch g.GetPhase() {
	case domain.FrenchTarotPhaseBid:
		b.WriteString(i18n.Tf("frenchtarot.promptBid",
			"name", cuiPlayerName(g.GetPlayer(g.GetBidPlayerIdx()), g.GetBidPlayerIdx()),
			"high", frenchTarotBidLabel(g.GetHighestBid())) + "\n")
		b.WriteString(i18n.T("frenchtarot.promptBidHelp") + "\n")
		// The contract multipliers are fixed rules; spell them out and remind the
		// player a new bid must outrank the current highest, since the CLI has no
		// button greying to signal it.
		b.WriteString(i18n.T("frenchtarot.promptBidLegend") + "\n")
	case domain.FrenchTarotPhaseChien:
		b.WriteString(i18n.Tf("frenchtarot.promptChien",
			"name", cuiPlayerName(g.GetPlayer(g.GetDeclarerIdx()), g.GetDeclarerIdx())) + "\n")
		b.WriteString(i18n.T("frenchtarot.promptChienHelp") + "\n")
		// **どの札が捨てられるか**を先に見せる。Web はカードごとのツールチップで
		// 理由を出しているのに、CUI はサーバに拒否されるまで分からなかった (#5712)。
		// 判定は domain の FrenchTarotBuriableIndices が唯一の出どころ。
		if human := g.GetPlayer(g.GetDeclarerIdx()); human != nil && human.GetIsHuman() {
			idxs := domain.FrenchTarotBuriableIndices(human)
			cards := make([]string, 0, len(idxs))
			for _, i := range idxs {
				cards = append(cards, "["+strconv.Itoa(i)+"]")
			}
			if len(cards) > 0 {
				b.WriteString(i18n.Tf("frenchtarot.promptChienBuriable",
					"cards", strings.Join(cards, " ")) + "\n")
			}
			b.WriteString(i18n.T("frenchtarot.promptChienLegend") + "\n")
		}
	case domain.FrenchTarotPhasePlay:
		currentIdx := g.GetCurrentPlayerIdx()
		b.WriteString(i18n.Tf("frenchtarot.promptPlay",
			"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
		b.WriteString(i18n.T("frenchtarot.promptPlayHelp") + "\n")
	case domain.FrenchTarotPhaseTrickEnd:
		b.WriteString(i18n.T("frenchtarot.promptTrickEnd") + "\n")
		b.WriteString(i18n.T("frenchtarot.promptTrickEndHelp") + "\n")
	case domain.FrenchTarotPhaseRoundEnd:
		b.WriteString(i18n.Tf("frenchtarot.promptRoundEnd",
			"declarer", cuiPlayerName(g.GetPlayer(g.GetDeclarerIdx()), g.GetDeclarerIdx()),
			"outcome", frenchTarotOutcomeLabel(g.GetOutcome())) + "\n")
		b.WriteString(i18n.T("frenchtarot.promptRoundEndHelp") + "\n")
	}
}

// HintOutput emits the current French Tarot hint.
func (p *FrenchTarotCuiPresenter) HintOutput(g interfaces.FrenchTarotGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("frenchtarot.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, frenchTarotHintReasonKeys)
	if len(hint.CardIndices) > 0 {
		playerIdx := g.GetCurrentPlayerIdx()
		if g.GetPhase() == domain.FrenchTarotPhaseChien {
			playerIdx = g.GetDeclarerIdx()
		}
		player := g.GetPlayer(playerIdx)
		cards := make([]string, len(hint.CardIndices))
		for i, idx := range hint.CardIndices {
			if player != nil && idx >= 0 && idx < player.GetCardsSize() {
				cards[i] = "[" + strconv.Itoa(idx) + "]" + frenchTarotCuiCardStr(player.GetCard(idx))
			} else {
				cards[i] = strconv.Itoa(idx)
			}
		}
		return color.Yellow(i18n.Tf("frenchtarot.hintCard",
			"cards", strings.Join(cards, ", "),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("frenchtarot.hintCard", "cards", "-", "reason", reason)) + "\n"
}

// frenchTarotHintReasonKeys maps hint-reason identifiers to i18n keys.
var frenchTarotHintReasonKeys = map[string]string{
	"bid_take":     "frenchtarot.hintReasonBidTake",
	"bid_pass":     "frenchtarot.hintReasonBidPass",
	"discard_weak": "frenchtarot.hintReasonDiscardWeak",
	"lead_high":    "frenchtarot.hintReasonLeadHigh",
	"lead_low":     "frenchtarot.hintReasonLeadLow",
	"follow_win":   "frenchtarot.hintReasonFollowWin",
	"follow_duck":  "frenchtarot.hintReasonFollowDuck",
	"play_excuse":  "frenchtarot.hintReasonPlayExcuse",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *FrenchTarotCuiPresenter) ActionLogOutput(g interfaces.FrenchTarotGame) string {
	return actionLogOutputTextWithNames(g, func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) })
}
