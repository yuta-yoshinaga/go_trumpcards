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

// bauernschnapsenPlayerStr returns the display string for a single Bauernschnapsen player.
func bauernschnapsenPlayerStr(player *domain.BauernschnapsenPlayer, i int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("bauernschnapsen.playerLine",
		"name", cuiPlayerName(player, i),
		"team", strconv.Itoa(player.GetTeam()),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// BauernschnapsenCuiPresenter renders the Bauernschnapsen CUI view.
type BauernschnapsenCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *BauernschnapsenCuiPresenter) Output(g interfaces.BauernschnapsenGame, lastErr error) string {
	return buildCuiOutput(i18n.T("bauernschnapsen.helpTitle"), func(out *strings.Builder) {
		out.WriteString(i18n.Tf("bauernschnapsen.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber())) + "\n")
		dealerIdx := g.GetDealerIdx()
		out.WriteString(i18n.Tf("bauernschnapsen.dealer",
			"name", cuiPlayerName(g.GetPlayer(dealerIdx), dealerIdx)) + "\n")

		// **切り札は宣言で決まる。** クローン元のガイゲルは表向きの 1 枚と
		// 山札の残りを併記していたが、このゲームは 20 枚を配り切るので
		// どちらも存在しない。契約と宣言者を代わりに出す。
		out.WriteString(bauernschnapsenContractStr(g))

		out.WriteString(i18n.Tf("bauernschnapsen.teamScoreLine",
			"t0", strconv.Itoa(g.GetTeamScore(0)),
			"t1", strconv.Itoa(g.GetTeamScore(1))) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			out.WriteString(bauernschnapsenPlayerStr(g.GetPlayer(i), i))
		}

		out.WriteString("----------\n")

		trick := g.GetCurrentTrick()
		cuiTrickBlock(out, trick,
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(out, lastErr)

		if g.GetGameEndFlag() {
			banner := i18n.Tf("bauernschnapsen.gameEnd", "team", strconv.Itoa(g.GetWinnerTeam()))
			out.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.BauernschnapsenPhaseContract:
			out.WriteString(i18n.Tf("bauernschnapsen.promptCurrentPlayer",
				"name", cuiPlayerName(g.GetPlayer(g.GetCurrentPlayerIdx()), g.GetCurrentPlayerIdx())) + "\n")
			out.WriteString(i18n.T("bauernschnapsen.promptContract") + "\n")
			out.WriteString(i18n.T("bauernschnapsen.promptContractHelp") + "\n")
		case domain.BauernschnapsenPhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			out.WriteString(i18n.Tf("bauernschnapsen.promptCurrentPlayer",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			if idxs := g.GetMarriageIndices(currentIdx); len(idxs) > 0 {
				out.WriteString(i18n.T("bauernschnapsen.promptMarriageHint") + "\n")
				// On the human's turn, name the K/Q cards (and their hand indices)
				// that can be declared; never for a CPU, to avoid leaking its hand.
				if human := g.GetPlayer(currentIdx); human != nil && human.GetIsHuman() {
					cards := make([]string, len(idxs))
					for i, idx := range idxs {
						cards[i] = "[" + strconv.Itoa(idx) + "]" + cuiCardStr(human.GetCard(idx))
					}
					out.WriteString(i18n.Tf("bauernschnapsen.promptMarriageCards",
						"cards", strings.Join(cards, ", ")) + "\n")
				}
			}
			out.WriteString(i18n.T("bauernschnapsen.promptPlayHelp") + "\n")
		case domain.BauernschnapsenPhaseTrickEnd:
			out.WriteString(i18n.T("bauernschnapsen.promptTrickEnd") + "\n")
			out.WriteString(i18n.T("bauernschnapsen.promptTrickEndHelp") + "\n")
		case domain.BauernschnapsenPhaseRoundEnd:
			out.WriteString(i18n.T("bauernschnapsen.promptRoundEnd") + "\n")
			out.WriteString(i18n.T("bauernschnapsen.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Bauernschnapsen hint.
func (p *BauernschnapsenCuiPresenter) HintOutput(g interfaces.BauernschnapsenGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("bauernschnapsen.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, bauernschnapsenHintReasonKeys)
	if hint.CardIndex == nil {
		return i18n.T("bauernschnapsen.hintNone") + "\n"
	}
	player := g.GetPlayer(0)
	card := player.GetCard(*hint.CardIndex)
	if hint.IsMarriage {
		return color.Yellow(i18n.Tf("bauernschnapsen.hintMarriage",
			"idx", strconv.Itoa(*hint.CardIndex),
			"card", cuiCardStr(card),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("bauernschnapsen.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", reason)) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *BauernschnapsenCuiPresenter) ActionLogOutput(g interfaces.BauernschnapsenGame) string {
	return actionLogOutputTextForSeats[*domain.BauernschnapsenPlayer](g)
}

// bauernschnapsenHintReasonKeys maps Bauernschnapsen-specific hint-reason identifiers to their
// i18n keys, consumed by hintReasonStr.
var bauernschnapsenHintReasonKeys = map[string]string{
	"lead_trump":  "bauernschnapsen.hintReasonLeadTrump",
	"lead_low":    "bauernschnapsen.hintReasonLeadLow",
	"lead_value":  "bauernschnapsen.hintReasonLeadValue",
	"follow_cut":  "bauernschnapsen.hintReasonFollowCut",
	"follow_win":  "bauernschnapsen.hintReasonFollowWin",
	"follow_dump": "bauernschnapsen.hintReasonFollowDump",
	"marriage":    "bauernschnapsen.hintReasonMarriage",
}

// bauernschnapsenContractStr は契約行を組み立てる。
//
// 契約が決まるまでは切り札も決まらないので、宣言待ちの間は「契約中」とだけ出す。
func bauernschnapsenContractStr(g interfaces.BauernschnapsenGame) string {
	if g.GetPhase() == domain.BauernschnapsenPhaseContract {
		return i18n.T("bauernschnapsen.contractPending") + "\n"
	}
	name := i18n.T(bauernschnapsenContractKey(g.GetContract()))
	declarer := g.GetDeclarerIdx()
	if declarer < 0 || declarer >= g.GetPlayerCnt() {
		return i18n.Tf("bauernschnapsen.contractLineNoDeclarer", "contract", name) + "\n"
	}
	// 切り札の無い契約 (ベテル) と、まだ切り札が決まっていない盤面では
	// スート名を出さない。出すと BauernschnapsenNoTrump が "UNKNOWN" として
	// 画面に漏れる。
	if g.GetContract() == domain.BauernschnapsenContractBettel ||
		g.GetTrumpSuit() < domain.CardDesignSpade || g.GetTrumpSuit() > domain.CardDesignMax {
		return i18n.Tf("bauernschnapsen.contractLineNoTrump",
			"contract", name,
			"name", cuiPlayerName(g.GetPlayer(declarer), declarer)) + "\n"
	}
	return i18n.Tf("bauernschnapsen.contractLine",
		"contract", name,
		"name", cuiPlayerName(g.GetPlayer(declarer), declarer),
		"suit", cuiSuitName(g.GetTrumpSuit())) + "\n"
}

// bauernschnapsenContractKey は契約に対応する i18n キーを返す。
func bauernschnapsenContractKey(c domain.BauernschnapsenContract) string {
	switch c {
	case domain.BauernschnapsenContractFarbenzwang:
		return "bauernschnapsen.contractFarbenzwang"
	case domain.BauernschnapsenContractBettel:
		return "bauernschnapsen.contractBettel"
	default:
		return "bauernschnapsen.contractRufer"
	}
}
