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

// viraBidName maps a bid constant (0-4) to its localized contract name.
func viraBidName(bid int) string {
	switch domain.ViraBid(bid) {
	case domain.ViraBidGask:
		return i18n.T("vira.bid.gask")
	case domain.ViraBidSolo:
		return i18n.T("vira.bid.solo")
	case domain.ViraBidMisere:
		return i18n.T("vira.bid.misere")
	case domain.ViraBidVira:
		return i18n.T("vira.bid.vira")
	default:
		return i18n.T("vira.bid.pass")
	}
}

// viraTrumpStr renders the trump glyph, or a "no trump" label when none.
func viraTrumpStr(suit int) string {
	if suit < domain.CardDesignSpade {
		return i18n.T("vira.noTrump")
	}
	return cuiSuitName(suit)
}

// viraPlayerStr returns the display string for a single player.
func viraPlayerStr(g interfaces.ViraGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	scores := g.GetPlayerScores()
	role := i18n.T("vira.roleDefender")
	if idx == g.GetDeclarerIdx() {
		role = i18n.T("vira.roleDeclarer")
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("vira.playerLine",
		"name", cuiPlayerName(player, idx),
		"role", role,
		"cards", strconv.Itoa(player.GetCardsSize()),
		"score", strconv.Itoa(scores[idx]),
		"tricks", strconv.Itoa(player.GetTrickCount()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// ViraCuiPresenter renders the Vira CUI view.
type ViraCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *ViraCuiPresenter) Output(g interfaces.ViraGame, lastErr error) string {
	return buildCuiOutput(i18n.T("vira.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("vira.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber()),
			"trump", viraTrumpStr(g.GetTrumpSuit())) + "\n")

		// **ポットは毎フレーム出す。**局をまたいで積み上がる唯一の数字で、
		// 精算はここから配られる。Web 版は出しているので、CUI で落とすと
		// 同じ局面が二つの画面で別物に見える。
		b.WriteString(i18n.Tf("vira.potLine", "pot", strconv.Itoa(g.GetPot())) + "\n")

		if g.GetDeclarerIdx() >= 0 {
			declIdx := g.GetDeclarerIdx()
			b.WriteString(i18n.Tf("vira.contractLine",
				"name", cuiPlayerName(g.GetPlayer(declIdx), declIdx),
				"contract", viraBidName(int(g.GetContract()))) + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(viraPlayerStr(g, i))
		}

		b.WriteString("----------\n")

		cuiTrickBlock(b, g.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winner := g.GetWinnerPlayer()
			var winnerStr string
			if winner >= 0 {
				winnerStr = cuiPlayerName(g.GetPlayer(winner), winner)
			}
			banner := i18n.Tf("vira.gameEnd", "name", winnerStr)
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		p.writePrompt(b, g)
	})
}

// writePrompt renders the phase-specific prompt block.
func (p *ViraCuiPresenter) writePrompt(b *strings.Builder, g interfaces.ViraGame) {
	switch g.GetPhase() {
	case domain.ViraPhaseBid:
		currentIdx := g.GetCurrentPlayerIdx()
		b.WriteString(i18n.Tf("vira.promptBid",
			"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx),
			"contract", viraBidName(int(g.GetContract()))) + "\n")
		b.WriteString(i18n.T("vira.promptBidHelp") + "\n")
	case domain.ViraPhasePlay:
		currentIdx := g.GetCurrentPlayerIdx()
		b.WriteString(i18n.Tf("vira.promptPlay",
			"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
		b.WriteString(i18n.T("vira.promptPlayHelp") + "\n")
	case domain.ViraPhaseTrickEnd:
		b.WriteString(i18n.T("vira.promptTrickEnd") + "\n")
		b.WriteString(i18n.T("vira.promptTrickEndHelp") + "\n")
	case domain.ViraPhaseRoundEnd:
		b.WriteString(i18n.T("vira.promptRoundEnd") + "\n")
		p.writeRoundEndResult(b, g)
		b.WriteString(i18n.T("vira.promptRoundEndHelp") + "\n")
	}
}

// writeRoundEndResult appends the declarer's contract outcome and a one-line
// trick tally for every player, matching the information the Web view already
// shows in its round-result block.
func (p *ViraCuiPresenter) writeRoundEndResult(b *strings.Builder, g interfaces.ViraGame) {
	declIdx := g.GetDeclarerIdx()
	if declIdx < 0 {
		return
	}
	decl := g.GetPlayer(declIdx)
	if decl == nil {
		return
	}
	contract := g.GetContract()
	declTricks := decl.GetTrickCount()
	// Six/Seven/Eight need at least the target tricks; Misère needs exactly zero.
	achieved := declTricks >= viraContractTarget(contract)
	if contract == domain.ViraBidMisere {
		achieved = declTricks == 0
	}
	outcome := i18n.T("vira.contractFailed")
	if achieved {
		outcome = i18n.T("vira.contractAchieved")
	}
	b.WriteString(i18n.Tf("vira.promptRoundEndResult",
		"name", cuiPlayerName(decl, declIdx),
		"contract", viraBidName(int(contract)),
		"tricks", strconv.Itoa(declTricks),
		"outcome", outcome) + "\n")

	entries := make([]string, 0, g.GetPlayerCnt())
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		entries = append(entries, i18n.Tf("vira.roundEndTrickEntry",
			"name", cuiPlayerName(player, i),
			"tricks", strconv.Itoa(player.GetTrickCount())))
	}
	b.WriteString(i18n.Tf("vira.roundEndTricks", "list", strings.Join(entries, ", ")) + "\n")
}

// viraContractTarget returns the number of tricks the contract requires.
//
// **The table lives in the domain.** Copying it here is how the two drift when
// the ladder changes; `ViraBidTarget` is exported for exactly this reason.
func viraContractTarget(bid domain.ViraBid) int { return domain.ViraBidTarget(bid) }

// HintOutput emits the current Vira hint.
func (p *ViraCuiPresenter) HintOutput(g interfaces.ViraGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("vira.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, viraHintReasonKeys)
	if len(hint.CardIndices) > 0 {
		playerIdx := g.GetCurrentPlayerIdx()
		player := g.GetPlayer(playerIdx)
		cards := make([]string, len(hint.CardIndices))
		for i, idx := range hint.CardIndices {
			if player != nil {
				cards[i] = "[" + strconv.Itoa(idx) + "]" + cuiCardStr(player.GetCard(idx))
			} else {
				cards[i] = strconv.Itoa(idx)
			}
		}
		return color.Yellow(i18n.Tf("vira.hintCard",
			"cards", strings.Join(cards, ", "),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("vira.hintCard",
		"cards", "-",
		"reason", reason)) + "\n"
}

// viraHintReasonKeys maps Vira-specific hint-reason identifiers to i18n keys.
//
// **ここは Vira.playHintReason が返す 6 種と 1 対 1 で対応させる。**外れた
// 理由は hintReasonStr の既定でキー文字列そのものが画面に出る。
var viraHintReasonKeys = map[string]string{
	"lead_high":    "vira.hintReasonLeadHigh",
	"lead_low":     "vira.hintReasonLeadLow",
	"follow_win":   "vira.hintReasonFollowWin",
	"follow_block": "vira.hintReasonFollowBlock",
	"misere_duck":  "vira.hintReasonMisereDuck",
	"misere_force": "vira.hintReasonMisereForce",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *ViraCuiPresenter) ActionLogOutput(g interfaces.ViraGame) string {
	return actionLogOutputTextForSeats[*domain.ViraPlayer](g)
}
