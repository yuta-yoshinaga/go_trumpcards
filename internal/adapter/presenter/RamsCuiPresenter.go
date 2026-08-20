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

// ramsPlayerStr returns the display string for a single player.
func ramsPlayerStr(player *domain.RamsPlayer, idx int, isDealer bool) string {
	var b strings.Builder
	// **参加判断もリードも親の左隣から始まる** (#5748)。誰が親かが出ていないと、
	// 自分が何番目に決断するのかが読めない。3〜5 人卓で毎ラウンド 1 つ回る。
	name := cuiPlayerName(player, idx)
	if isDealer {
		name += i18n.T("rams.dealerMark")
	}
	b.WriteString(i18n.Tf("rams.playerLine",
		"name", name,
		"chips", strconv.Itoa(player.GetChips()),
		"status", ramsStatusStr(player),
		"tricks", strconv.Itoa(player.GetRoundTricks()),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// ramsStatusStr 参加しているか降りたか、まだ選んでいないか
func ramsStatusStr(player *domain.RamsPlayer) string {
	if !player.GetDecided() {
		return i18n.T("rams.statusUndecided")
	}
	if player.GetInRound() {
		return i18n.T("rams.statusIn")
	}
	return i18n.T("rams.statusOut")
}

// RamsCuiPresenter renders the Rams CUI view.
type RamsCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *RamsCuiPresenter) Output(r interfaces.RamsGame, lastErr error) string {
	return buildCuiOutput(i18n.T("rams.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("rams.header",
			"round", strconv.Itoa(r.GetRoundNumber()),
			"rounds", strconv.Itoa(r.GetConfig().Rounds),
			"trick", strconv.Itoa(r.GetTrickNumber()+1),
			"tricks", strconv.Itoa(domain.RamsTricksPerRound),
			"players", strconv.Itoa(r.GetPlayerCnt())) + "\n")
		// **ポットと切り札は常に見せる。** 参加判断の材料そのもの。
		sb.WriteString(color.Yellow(i18n.Tf("rams.potLine", "pot", strconv.Itoa(r.GetPot()))) + "\n")
		if up := r.GetUpCard(); up != nil {
			sb.WriteString(i18n.Tf("rams.trumpLine", "card", cuiCardStr(up)) + "\n")
		}
		// 参加して 0 トリックだと余分に払う、というリスクの明示。
		sb.WriteString(i18n.Tf("rams.riskLine", "penalty", strconv.Itoa(domain.RamsMissPenalty)) + "\n")

		for i := 0; i < r.GetPlayerCnt(); i++ {
			sb.WriteString(ramsPlayerStr(r.GetPlayer(i), i, i == r.GetDealerIdx()))
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
				banner = i18n.T("rams.gameEndTie")
			} else {
				banner = i18n.Tf("rams.gameEndWinner",
					"name", cuiPlayerName(r.GetPlayer(r.GetWinnerIdx()), r.GetWinnerIdx()),
					"chips", strconv.Itoa(r.GetPlayer(r.GetWinnerIdx()).GetChips()))
			}
			sb.WriteString(color.Green(banner) + "\n")
			return
		}

		switch r.GetPhase() {
		case domain.RamsPhaseDecide:
			sb.WriteString(i18n.T("rams.promptDecide") + "\n")
			sb.WriteString(i18n.T("rams.promptDecideHelp") + "\n")
			return
		case domain.RamsPhaseRoundEnd:
			sb.WriteString(i18n.T("rams.promptRoundEnd") + "\n")
			sb.WriteString(i18n.T("rams.promptNext") + "\n")
			return
		}

		// 降りたラウンドは操作しない。待ちに見えないよう明示する。
		if !r.GetPlayer(0).GetInRound() {
			sb.WriteString(i18n.T("rams.promptWatching") + "\n")
			return
		}

		currentIdx := r.GetCurrentPlayerIdx()
		sb.WriteString(i18n.Tf("rams.promptCurrentPlayer",
			"name", cuiPlayerName(r.GetPlayer(currentIdx), currentIdx)) + "\n")
		sb.WriteString(i18n.T("rams.promptPlay") + "\n")
	})
}

// HintOutput emits the current hint.
func (p *RamsCuiPresenter) HintOutput(r interfaces.RamsGame) string {
	hint := r.GetHint()
	if hint == nil {
		return i18n.T("rams.hintNone") + "\n"
	}
	// **選択フェーズでは札ではなく参加可否を助言する。**
	if hint.CardIndex == nil {
		return color.Yellow(i18n.Tf("rams.hintDecide",
			"reason", hintReasonStr(hint.Reason, ramsHintReasonKeys))) + "\n"
	}
	card := r.GetPlayer(0).GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("rams.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", hintReasonStr(hint.Reason, ramsHintReasonKeys))) + "\n"
}

// ramsHintReasonKeys maps hint-reason identifiers to their i18n keys.
var ramsHintReasonKeys = map[string]string{
	"ramsPlayIn":      "rams.hintReasonPlayIn",
	"ramsPassOut":     "rams.hintReasonPassOut",
	"ramsTakeTrick":   "rams.hintReasonTakeTrick",
	"ramsAlreadySafe": "rams.hintReasonAlreadySafe",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *RamsCuiPresenter) ActionLogOutput(r interfaces.RamsGame) string {
	return actionLogOutputTextForSeats[*domain.RamsPlayer](r)
}
