//go:build !js || !wasm || extra

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// ZwanzigerrufenCuiPresenter renders the Zwanzigerrufen CUI view.
type ZwanzigerrufenCuiPresenter struct{}

// zwanzigerrufenIndexedHand 人間の手札をインデックス付きで表示する。
func zwanzigerrufenIndexedHand(p *domain.ZwanzigerrufenPlayer) string {
	parts := make([]string, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		parts[i] = "[" + strconv.Itoa(i) + "]" + koenigrufenCuiCardStr(p.GetCard(i))
	}
	return strings.Join(parts, "  ")
}

// zwanzigerrufenContractLabel 契約の i18n ラベルを返す。
func zwanzigerrufenContractLabel(bid domain.ZwanzigerrufenBid) string {
	return i18n.T("zwanzigerrufen.contract." + domain.ZwanzigerrufenBidName(bid))
}

// zwanzigerrufenRoleLabel 席の役割ラベルを返す。
//
// **パートナーは公開済みのときだけ明示する。** Trischaken には役割そのものが無い。
func zwanzigerrufenRoleLabel(g interfaces.ZwanzigerrufenGame, idx int) string {
	if g.GetContract() == domain.ZwanzigerrufenBidTrischaken {
		return i18n.T("zwanzigerrufen.roleSolo")
	}
	if idx == g.GetDeclarerIdx() {
		return i18n.T("zwanzigerrufen.roleDeclarer")
	}
	if g.GetPartnerRevealed() && g.GetPartnerIdx() == idx {
		return i18n.T("zwanzigerrufen.rolePartner")
	}
	return i18n.T("zwanzigerrufen.roleOpponent")
}

// zwanzigerrufenPlayerStr 席 1 行ぶんの状態文字列を返す。
func zwanzigerrufenPlayerStr(g interfaces.ZwanzigerrufenGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("zwanzigerrufen.playerLine",
		"name", cuiPlayerName(player, idx),
		"role", zwanzigerrufenRoleLabel(g, idx),
		"cards", strconv.Itoa(player.GetCardsSize()),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"points", strconv.Itoa(g.GetCardPoints(idx)),
		"score", strconv.Itoa(g.GetPlayerScore(idx))) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(zwanzigerrufenIndexedHand(player) + "\n")
	}
	return b.String()
}

// Output renders the current game state for the active locale.
func (p *ZwanzigerrufenCuiPresenter) Output(g interfaces.ZwanzigerrufenGame, lastErr error) string {
	return buildCuiOutput(i18n.T("zwanzigerrufen.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("zwanzigerrufen.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"rounds", strconv.Itoa(g.GetConfig().TargetDeals),
			"trick", strconv.Itoa(g.GetTrickNumber()),
			"contract", zwanzigerrufenContractLabel(g.GetContract())) + "\n")
		if g.GetCalledTrump() > 0 {
			b.WriteString(i18n.Tf("zwanzigerrufen.calledLine",
				"trump", strconv.Itoa(g.GetCalledTrump())) + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(zwanzigerrufenPlayerStr(g, i))
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
			if winner < 0 {
				b.WriteString(color.Green(i18n.T("zwanzigerrufen.gameDraw")) + "\n")
				return
			}
			b.WriteString(color.Green(i18n.Tf("zwanzigerrufen.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(winner), winner))) + "\n")
			return
		}
		p.writePrompt(b, g)
	})
}

// writePrompt 現在のフェーズに応じたプロンプトを書き込む。
func (p *ZwanzigerrufenCuiPresenter) writePrompt(b *strings.Builder, g interfaces.ZwanzigerrufenGame) {
	switch g.GetPhase() {
	case domain.ZwanzigerrufenPhaseBid:
		idx := g.GetBidPlayerIdx()
		b.WriteString(i18n.Tf("zwanzigerrufen.promptBid",
			"name", cuiPlayerName(g.GetPlayer(idx), idx),
			"high", zwanzigerrufenContractLabel(g.GetHighestBid())) + "\n")
		b.WriteString(i18n.T("zwanzigerrufen.promptBidHelp") + "\n")
	case domain.ZwanzigerrufenPhaseTalon:
		idx := g.GetDeclarerIdx()
		b.WriteString(i18n.Tf("zwanzigerrufen.promptTalon",
			"name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
		// **伏せられない札を先に言う。** Web は押せない札を灰色にするが、CLI では
		// 制約が見えないので、キングとトゥルルが対象外であることを書く。
		b.WriteString(i18n.T("zwanzigerrufen.promptTalonHelp") + "\n")
	case domain.ZwanzigerrufenPhasePlay:
		idx := g.GetCurrentPlayerIdx()
		b.WriteString(i18n.Tf("zwanzigerrufen.promptPlay",
			"name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
		b.WriteString(i18n.T("zwanzigerrufen.promptPlayHelp") + "\n")
	case domain.ZwanzigerrufenPhaseTrickEnd:
		winner := g.GetLastTrickWinner()
		b.WriteString(i18n.Tf("zwanzigerrufen.promptTrickEnd",
			"name", cuiPlayerName(g.GetPlayer(winner), winner)) + "\n")
		b.WriteString(i18n.T("zwanzigerrufen.promptTrickEndHelp") + "\n")
	case domain.ZwanzigerrufenPhaseRoundEnd:
		b.WriteString(p.roundResultStr(g) + "\n")
		b.WriteString(i18n.T("zwanzigerrufen.promptRoundEndHelp") + "\n")
	}
}

// roundResultStr ディール結果の説明文を返す。
//
// **Trischaken だけ文が違う。** 達成/失敗ではなく「誰がいちばん取ってしまったか」
// を告げる契約なので、同じ文型に押し込むと勝敗の向きが逆に読める。
func (p *ZwanzigerrufenCuiPresenter) roundResultStr(g interfaces.ZwanzigerrufenGame) string {
	bd := g.GetBreakdown()
	if bd == nil {
		return i18n.T("zwanzigerrufen.promptRoundEndNone")
	}
	if bd.Contract == domain.ZwanzigerrufenBidTrischaken {
		return i18n.Tf("zwanzigerrufen.roundTrischaken",
			"name", cuiPlayerName(g.GetPlayer(bd.Loser), bd.Loser),
			"points", strconv.Itoa(bd.TeamPoints))
	}
	key := "zwanzigerrufen.roundLoss"
	if bd.Won {
		key = "zwanzigerrufen.roundWin"
	}
	declarer := g.GetDeclarerIdx()
	return i18n.Tf(key,
		"name", cuiPlayerName(g.GetPlayer(declarer), declarer),
		"points", strconv.Itoa(bd.TeamPoints),
		"threshold", strconv.Itoa(bd.Threshold))
}

// HintOutput emits the current Zwanzigerrufen hint.
func (p *ZwanzigerrufenCuiPresenter) HintOutput(g interfaces.ZwanzigerrufenGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("zwanzigerrufen.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, zwanzigerrufenHintReasonKeys)
	switch {
	case hint.Bid != nil:
		return color.Yellow(i18n.Tf("zwanzigerrufen.hintCard",
			"cards", zwanzigerrufenContractLabel(domain.ZwanzigerrufenBid(*hint.Bid)),
			"reason", reason)) + "\n"
	case len(hint.DiscardIndices) > 0:
		return color.Yellow(i18n.Tf("zwanzigerrufen.hintCard",
			"cards", p.cardList(g, g.GetDeclarerIdx(), hint.DiscardIndices),
			"reason", reason)) + "\n"
	case hint.CardIndex != nil:
		return color.Yellow(i18n.Tf("zwanzigerrufen.hintCard",
			"cards", p.cardList(g, g.GetCurrentPlayerIdx(), []int{*hint.CardIndex}),
			"reason", reason)) + "\n"
	default:
		return color.Yellow(i18n.Tf("zwanzigerrufen.hintCard", "cards", "-", "reason", reason)) + "\n"
	}
}

// cardList 添字の並びを "[0]T20" のような表示に直す。
func (p *ZwanzigerrufenCuiPresenter) cardList(g interfaces.ZwanzigerrufenGame, seat int, indices []int) string {
	player := g.GetPlayer(seat)
	parts := make([]string, len(indices))
	for i, idx := range indices {
		if player != nil && idx >= 0 && idx < player.GetCardsSize() {
			parts[i] = "[" + strconv.Itoa(idx) + "]" + koenigrufenCuiCardStr(player.GetCard(idx))
			continue
		}
		parts[i] = strconv.Itoa(idx)
	}
	return strings.Join(parts, ", ")
}

// zwanzigerrufenHintReasonKeys maps hint-reason identifiers to i18n keys.
var zwanzigerrufenHintReasonKeys = map[string]string{
	"pass_weak_hand":    "zwanzigerrufen.hintReasonPass",
	"bid_strong_trumps": "zwanzigerrufen.hintReasonBid",
	"bury_cheap_cards":  "zwanzigerrufen.hintReasonBury",
	"play_low":          "zwanzigerrufen.hintReasonPlayLow",
	"avoid_points":      "zwanzigerrufen.hintReasonAvoidPoints",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *ZwanzigerrufenCuiPresenter) ActionLogOutput(g interfaces.ZwanzigerrufenGame) string {
	return actionLogOutputTextForSeats[*domain.ZwanzigerrufenPlayer](g)
}
