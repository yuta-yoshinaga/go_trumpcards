//go:build !js || !wasm || extra2

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// julepePlayerStr returns the display string for a single player.
func julepePlayerStr(player *domain.JulepePlayer, idx int, isDealer bool) string {
	var b strings.Builder
	// **参加判断もリードも親の左隣から始まる** (#5748)。誰が親かが出ていないと、
	// 自分が何番目に決断するのかが読めない。3〜5 人卓で毎ラウンド 1 つ回る。
	name := cuiPlayerName(player, idx)
	if isDealer {
		name += i18n.T("julepe.dealerMark")
	}
	b.WriteString(i18n.Tf("julepe.playerLine",
		"name", name,
		"chips", strconv.Itoa(player.GetChips()),
		"status", julepeStatusStr(player),
		"tricks", strconv.Itoa(player.GetRoundTricks()),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// julepeStatusStr 参加しているか降りたか、まだ選んでいないか
func julepeStatusStr(player *domain.JulepePlayer) string {
	if !player.GetDecided() {
		return i18n.T("julepe.statusUndecided")
	}
	if player.GetInRound() {
		return i18n.T("julepe.statusIn")
	}
	return i18n.T("julepe.statusOut")
}

// JulepeCuiPresenter renders the Julepe CUI view.
type JulepeCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *JulepeCuiPresenter) Output(r interfaces.JulepeGame, lastErr error) string {
	return buildCuiOutput(i18n.T("julepe.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("julepe.header",
			"round", strconv.Itoa(r.GetRoundNumber()),
			"rounds", strconv.Itoa(r.GetConfig().Rounds),
			"trick", strconv.Itoa(r.GetTrickNumber()+1),
			"tricks", strconv.Itoa(domain.JulepeTricksPerRound),
			"players", strconv.Itoa(r.GetPlayerCnt())) + "\n")
		// **ポットと切り札は常に見せる。** 参加判断の材料そのもの。
		sb.WriteString(color.Yellow(i18n.Tf("julepe.potLine", "pot", strconv.Itoa(r.GetPot()))) + "\n")
		if up := r.GetUpCard(); up != nil {
			sb.WriteString(i18n.Tf("julepe.trumpLine", "card", cuiCardStr(up)) + "\n")
		}
		// 参加して 0 トリックだと余分に払う、というリスクの明示。
		// **規定トリック数は参加人数で変わる。** 固定値で説明すると
		// ルールを取り違える。
		sb.WriteString(i18n.Tf("julepe.riskLine",
			"penalty", strconv.Itoa(domain.JulepeMissPenalty),
			"required", strconv.Itoa(r.GetRequiredTricks())) + "\n")

		for i := 0; i < r.GetPlayerCnt(); i++ {
			sb.WriteString(julepePlayerStr(r.GetPlayer(i), i, i == r.GetDealerIdx()))
		}

		sb.WriteString("----------\n")

		cuiTrickBlock(sb, r.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(r.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(sb, lastErr)

		if r.GetGameEndFlag() {
			var banner string
			if r.GetWinnerIdx() < 0 {
				banner = i18n.T("julepe.gameEndTie")
			} else {
				banner = i18n.Tf("julepe.gameEndWinner",
					"name", cuiPlayerName(r.GetPlayer(r.GetWinnerIdx()), r.GetWinnerIdx()),
					"chips", strconv.Itoa(r.GetPlayer(r.GetWinnerIdx()).GetChips()))
			}
			sb.WriteString(color.Green(banner) + "\n")
			return
		}

		switch r.GetPhase() {
		case domain.JulepePhaseDecide:
			sb.WriteString(i18n.T("julepe.promptDecide") + "\n")
			sb.WriteString(i18n.T("julepe.promptDecideHelp") + "\n")
			return
		case domain.JulepePhaseRoundEnd:
			sb.WriteString(i18n.T("julepe.promptRoundEnd") + "\n")
			sb.WriteString(i18n.T("julepe.promptNext") + "\n")
			return
		}

		// 降りたラウンドは操作しない。待ちに見えないよう明示する。
		if !r.GetPlayer(0).GetInRound() {
			sb.WriteString(i18n.T("julepe.promptWatching") + "\n")
			return
		}

		currentIdx := r.GetCurrentPlayerIdx()
		sb.WriteString(i18n.Tf("julepe.promptCurrentPlayer",
			"name", cuiPlayerName(r.GetPlayer(currentIdx), currentIdx)) + "\n")
		sb.WriteString(i18n.T("julepe.promptPlay") + "\n")
	})
}

// HintOutput emits the current hint.
func (p *JulepeCuiPresenter) HintOutput(r interfaces.JulepeGame) string {
	hint := r.GetHint()
	if hint == nil {
		return i18n.T("julepe.hintNone") + "\n"
	}
	// **選択フェーズでは札ではなく参加可否を助言する。**
	if hint.CardIndex == nil {
		return color.Yellow(i18n.Tf("julepe.hintDecide",
			"reason", hintReasonStr(hint.Reason, julepeHintReasonKeys))) + "\n"
	}
	card := r.GetPlayer(0).GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("julepe.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", hintReasonStr(hint.Reason, julepeHintReasonKeys))) + "\n"
}

// julepeHintReasonKeys maps hint-reason identifiers to their i18n keys.
var julepeHintReasonKeys = map[string]string{
	"julepePlayIn":      "julepe.hintReasonPlayIn",
	"julepePassOut":     "julepe.hintReasonPassOut",
	"julepeTakeTrick":   "julepe.hintReasonTakeTrick",
	"julepeAlreadySafe": "julepe.hintReasonAlreadySafe",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *JulepeCuiPresenter) ActionLogOutput(r interfaces.JulepeGame) string {
	return actionLogOutputTextForSeats[*domain.JulepePlayer](r)
}
