//go:build !js || !wasm || extra3

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func pochCardListStr(cards []*domain.Card, indexed bool) string {
	if len(cards) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(cards))
	for i, c := range cards {
		if indexed {
			parts = append(parts, "["+strconv.Itoa(i)+"]"+cuiCardStr(c))
			continue
		}
		parts = append(parts, cuiCardStr(c))
	}
	return strings.Join(parts, " ")
}

// pochPoolLabel はプールの表示名を返す。
func pochPoolLabel(pool domain.PochPool) string {
	return i18n.T("poch.pool." + pool.String())
}

// PochCuiPresenter renders the Poch CUI view.
type PochCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *PochCuiPresenter) Output(c interfaces.PochGame, lastErr error) string {
	return buildCuiOutput(i18n.T("poch.helpTitle"), func(sb *strings.Builder) {
		turnUp := "-"
		if t := c.GetTurnUp(); t != nil {
			turnUp = cuiCardStr(t)
		}
		sb.WriteString(i18n.Tf("poch.header",
			"deal", strconv.Itoa(c.GetDealNumber()+1),
			"target", strconv.Itoa(c.GetConfig().TargetDeals),
			"turnup", turnUp) + "\n")
		// 「pay suit の札でなければプールは取れない」と「Zwick ではなく組の
		// 比べ合い」がこの game で最も誤解されやすいので毎回出す。
		sb.WriteString(i18n.T("poch.ruleLine") + "\n")

		// **9 区画は全部出す。**持ち越しが貯まっている区画がどれかは、
		// このディールで何を狙うかそのものになる。
		board := c.GetBoard()
		pools := make([]string, 0, domain.PochPoolCount)
		for i := range domain.PochPoolCount {
			pool := domain.PochPool(i)
			pools = append(pools, pochPoolLabel(pool)+":"+strconv.Itoa(board.Get(pool)))
		}
		sb.WriteString(i18n.Tf("poch.boardLine", "pools", strings.Join(pools, " ")) + "\n")

		// 第 1 段階は自動で解決するので、結果を出さないと何が起きたのか判らない。
		for _, a := range c.GetStakingAwards() {
			if a == nil {
				continue
			}
			sb.WriteString(i18n.Tf("poch.awardLine",
				"name", cuiPlayerName(c.GetPlayer(a.Player), a.Player),
				"pool", pochPoolLabel(a.Pool),
				"chips", strconv.Itoa(a.Chips)) + "\n")
		}

		if pile := c.GetPlayedPile(); len(pile) > 0 {
			sb.WriteString(i18n.Tf("poch.pileLine", "cards", pochCardListStr(pile, false)) + "\n")
		}

		for i, player := range c.GetPlayers() {
			if player == nil {
				continue
			}
			line := i18n.Tf("poch.playerLine",
				"name", cuiPlayerName(player, i),
				"chips", strconv.Itoa(player.GetChips()),
				"cards", strconv.Itoa(player.GetCardsSize()),
				"bet", strconv.Itoa(player.GetBet()))
			if player.IsFolded() {
				line += " " + i18n.T("poch.foldedMark")
			}
			sb.WriteString(line + "\n")
			if player.GetIsHuman() && player.GetCardsSize() > 0 {
				hand := make([]*domain.Card, 0, player.GetCardsSize())
				for j := range player.GetCardsSize() {
					hand = append(hand, player.GetCard(j))
				}
				sb.WriteString("  " + pochCardListStr(hand, true) + "\n")
			}
		}

		sb.WriteString("----------\n")
		cuiErrorBlock(sb, lastErr)
		sb.WriteString(p.promptBlock(c))
	})
}

// promptBlock はフェーズごとの案内を返す。
func (p *PochCuiPresenter) promptBlock(c interfaces.PochGame) string {
	if c.GetGameEndFlag() {
		key := "poch.gameEndLose"
		if c.GetWinnerIdx() == 0 {
			key = "poch.gameEndWin"
		}
		return color.Green(i18n.T(key)) + "\n"
	}
	switch c.GetPhase() {
	case domain.PochPhaseStaking:
		return i18n.T("poch.promptStaking") + "\n"
	case domain.PochPhasePochen:
		return i18n.Tf("poch.promptPochen", "target", strconv.Itoa(c.GetBetTarget())) + "\n"
	case domain.PochPhaseStops:
		var sb strings.Builder
		if w := c.GetPochenWinner(); w >= 0 {
			sb.WriteString(i18n.Tf("poch.pochenResult",
				"name", cuiPlayerName(c.GetPlayer(w), w),
				"pot", strconv.Itoa(c.GetPochenPot())) + "\n")
		}
		// 並びが止まっているかどうかで、出せる札がまるで違う。
		if c.GetStopsSuit() < 0 {
			sb.WriteString(i18n.T("poch.promptFreeLead") + "\n")
		} else {
			sb.WriteString(i18n.T("poch.promptFollow") + "\n")
		}
		return sb.String()
	case domain.PochPhaseDealEnd:
		var sb strings.Builder
		if w := c.GetDealWinner(); w >= 0 {
			sb.WriteString(i18n.Tf("poch.dealResult",
				"name", cuiPlayerName(c.GetPlayer(w), w)) + "\n")
		}
		sb.WriteString(i18n.T("poch.promptNext") + "\n")
		return sb.String()
	case domain.PochPhaseGameEnd:
		return ""
	}
	return ""
}

// HintOutput emits the current Poch hint.
func (p *PochCuiPresenter) HintOutput(c interfaces.PochGame) string {
	hint := pochHint(c)
	key := pochHintReasonKeys[hint.Reason]
	if key == "" {
		key = "poch.hintNone"
	}
	switch {
	case hint.Action == "play" && hint.CardIndex != nil:
		return color.Yellow(i18n.Tf("poch.hintPlay",
			"idx", strconv.Itoa(*hint.CardIndex), "reason", i18n.T(key))) + "\n"
	case hint.Action == "bet":
		return color.Yellow(i18n.Tf("poch.hintBet", "reason", i18n.T(key))) + "\n"
	case hint.Action == "fold":
		return color.Yellow(i18n.Tf("poch.hintFold", "reason", i18n.T(key))) + "\n"
	default:
		return color.Yellow(i18n.T(key)) + "\n"
	}
}

// pochHintReasonKeys maps the reason identifiers pochHint returns to i18n keys.
// The Web presenter ships the identifier and the frontend resolves it; the CUI
// must resolve it here or it prints the raw key.
var pochHintReasonKeys = map[string]string{
	"poch.hint.game_end":      "poch.hintReasonGameEnd",
	"poch.hint.deal_end":      "poch.hintReasonDealEnd",
	"poch.hint.not_your_turn": "poch.hintReasonNotYourTurn",
	"poch.hint.bet":           "poch.hintReasonBet",
	"poch.hint.fold":          "poch.hintReasonFold",
	"poch.hint.play":          "poch.hintReasonPlay",
	"poch.hint.none":          "poch.hintNone",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *PochCuiPresenter) ActionLogOutput(c interfaces.PochGame) string {
	return actionLogOutputText(c)
}
