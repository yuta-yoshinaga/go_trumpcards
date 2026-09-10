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

// TappTarockCuiPresenter renders the TappTarock CUI view.
type TappTarockCuiPresenter struct{}

// tapptarockIndexedHand 人間の手札をインデックス付きで表示する。
func tapptarockIndexedHand(p *domain.TappTarockPlayer) string {
	parts := make([]string, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		parts[i] = "[" + strconv.Itoa(i) + "]" + koenigrufenCuiCardStr(p.GetCard(i))
	}
	return strings.Join(parts, "  ")
}

// tapptarockContractLabel 契約の i18n ラベルを返す。
func tapptarockContractLabel(bid domain.TappTarockBid) string {
	return i18n.T("tapptarock.contract." + domain.TappTarockBidName(bid))
}

// tapptarockRoleLabel 席の役割ラベルを返す。
//
// Trischaken にはデクレアラー側の役割が無い。
func tapptarockRoleLabel(g interfaces.TappTarockGame, idx int) string {
	if g.GetContract() == domain.TappTarockBidTrischaken {
		return i18n.T("tapptarock.roleSolo")
	}
	if idx == g.GetDeclarerIdx() {
		return i18n.T("tapptarock.roleDeclarer")
	}
	return i18n.T("tapptarock.roleOpponent")
}

// tapptarockPlayerStr 席 1 行ぶんの状態文字列を返す。
func tapptarockPlayerStr(g interfaces.TappTarockGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("tapptarock.playerLine",
		"name", cuiPlayerName(player, idx),
		"role", tapptarockRoleLabel(g, idx),
		"cards", strconv.Itoa(player.GetCardsSize()),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"points", strconv.Itoa(g.GetCardPoints(idx)),
		"score", strconv.Itoa(g.GetPlayerScore(idx))) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(tapptarockIndexedHand(player) + "\n")
	}
	return b.String()
}

// Output renders the current game state for the active locale.
func (p *TappTarockCuiPresenter) Output(g interfaces.TappTarockGame, lastErr error) string {
	return buildCuiOutput(i18n.T("tapptarock.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("tapptarock.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"rounds", strconv.Itoa(g.GetConfig().TargetDeals),
			"trick", strconv.Itoa(g.GetTrickNumber()),
			"contract", tapptarockContractLabel(g.GetContract())) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(tapptarockPlayerStr(g, i))
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
				b.WriteString(color.Green(i18n.T("tapptarock.gameDraw")) + "\n")
				return
			}
			b.WriteString(color.Green(i18n.Tf("tapptarock.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(winner), winner))) + "\n")
			return
		}
		p.writePrompt(b, g)
	})
}

// writePrompt 現在のフェーズに応じたプロンプトを書き込む。
func (p *TappTarockCuiPresenter) writePrompt(b *strings.Builder, g interfaces.TappTarockGame) {
	switch g.GetPhase() {
	case domain.TappTarockPhaseBid:
		idx := g.GetBidPlayerIdx()
		b.WriteString(i18n.Tf("tapptarock.promptBid",
			"name", cuiPlayerName(g.GetPlayer(idx), idx),
			"high", tapptarockContractLabel(g.GetHighestBid())) + "\n")
		b.WriteString(i18n.T("tapptarock.promptBidHelp") + "\n")
	case domain.TappTarockPhaseTalon:
		idx := g.GetDeclarerIdx()
		b.WriteString(i18n.Tf("tapptarock.promptTalon",
			"name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
		// **伏せられない札を先に言う。** Web は押せない札を灰色にするが、CLI では
		// 制約が見えないので、キングとトゥルルが対象外であることを書く。
		b.WriteString(i18n.T("tapptarock.promptTalonHelp") + "\n")
	case domain.TappTarockPhasePlay:
		idx := g.GetCurrentPlayerIdx()
		b.WriteString(i18n.Tf("tapptarock.promptPlay",
			"name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
		b.WriteString(i18n.T("tapptarock.promptPlayHelp") + "\n")
	case domain.TappTarockPhaseTrickEnd:
		winner := g.GetLastTrickWinner()
		b.WriteString(i18n.Tf("tapptarock.promptTrickEnd",
			"name", cuiPlayerName(g.GetPlayer(winner), winner)) + "\n")
		b.WriteString(i18n.T("tapptarock.promptTrickEndHelp") + "\n")
	case domain.TappTarockPhaseRoundEnd:
		b.WriteString(p.roundResultStr(g) + "\n")
		b.WriteString(i18n.T("tapptarock.promptRoundEndHelp") + "\n")
	}
}

// roundResultStr ディール結果の説明文を返す。
//
// **Trischaken だけ文が違う。** 達成/失敗ではなく「誰がいちばん取ってしまったか」
// を告げる契約なので、同じ文型に押し込むと勝敗の向きが逆に読める。
func (p *TappTarockCuiPresenter) roundResultStr(g interfaces.TappTarockGame) string {
	bd := g.GetBreakdown()
	if bd == nil {
		return i18n.T("tapptarock.promptRoundEndNone")
	}
	if bd.Contract == domain.TappTarockBidTrischaken {
		return i18n.Tf("tapptarock.roundTrischaken",
			"name", cuiPlayerName(g.GetPlayer(bd.Loser), bd.Loser),
			"points", strconv.Itoa(bd.TeamPoints))
	}
	key := "tapptarock.roundLoss"
	if bd.Won {
		key = "tapptarock.roundWin"
	}
	declarer := g.GetDeclarerIdx()
	return i18n.Tf(key,
		"name", cuiPlayerName(g.GetPlayer(declarer), declarer),
		"points", strconv.Itoa(bd.TeamPoints),
		"threshold", strconv.Itoa(bd.Threshold))
}

// HintOutput emits the current TappTarock hint.
func (p *TappTarockCuiPresenter) HintOutput(g interfaces.TappTarockGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("tapptarock.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, tapptarockHintReasonKeys)
	switch {
	case hint.Bid != nil:
		return color.Yellow(i18n.Tf("tapptarock.hintCard",
			"cards", tapptarockContractLabel(domain.TappTarockBid(*hint.Bid)),
			"reason", reason)) + "\n"
	case len(hint.DiscardIndices) > 0:
		return color.Yellow(i18n.Tf("tapptarock.hintCard",
			"cards", p.cardList(g, g.GetDeclarerIdx(), hint.DiscardIndices),
			"reason", reason)) + "\n"
	case hint.CardIndex != nil:
		return color.Yellow(i18n.Tf("tapptarock.hintCard",
			"cards", p.cardList(g, g.GetCurrentPlayerIdx(), []int{*hint.CardIndex}),
			"reason", reason)) + "\n"
	default:
		return color.Yellow(i18n.Tf("tapptarock.hintCard", "cards", "-", "reason", reason)) + "\n"
	}
}

// cardList 添字の並びを "[0]T20" のような表示に直す。
func (p *TappTarockCuiPresenter) cardList(g interfaces.TappTarockGame, seat int, indices []int) string {
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

// tapptarockHintReasonKeys maps hint-reason identifiers to i18n keys.
var tapptarockHintReasonKeys = map[string]string{
	"pass_weak_hand":    "tapptarock.hintReasonPass",
	"bid_strong_trumps": "tapptarock.hintReasonBid",
	"bury_cheap_cards":  "tapptarock.hintReasonBury",
	"play_low":          "tapptarock.hintReasonPlayLow",
	"avoid_points":      "tapptarock.hintReasonAvoidPoints",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *TappTarockCuiPresenter) ActionLogOutput(g interfaces.TappTarockGame) string {
	return actionLogOutputTextForSeats[*domain.TappTarockPlayer](g)
}
