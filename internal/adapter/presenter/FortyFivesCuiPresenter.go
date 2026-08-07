//go:build !js || !wasm || casino

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// fortyFivesTeamLabels maps a team index (0/1) to its display name.
var fortyFivesTeamLabels = [domain.FortyFivesTeamCnt]string{"A", "B"}

// fortyFivesTeamLabel returns the localized team label (A/B) for a team index.
func fortyFivesTeamLabel(team int) string {
	if team < 0 || team >= len(fortyFivesTeamLabels) {
		return "?"
	}
	return fortyFivesTeamLabels[team]
}

// fortyFivesBidName maps a bid constant (0/15/20/25) to its localized contract name.
func fortyFivesBidName(bid int) string {
	switch domain.FortyFivesBid(bid) {
	case domain.FortyFivesBidFifteen:
		return i18n.T("fortyfives.bid.fifteen")
	case domain.FortyFivesBidTwenty:
		return i18n.T("fortyfives.bid.twenty")
	case domain.FortyFivesBidTwentyFive:
		return i18n.T("fortyfives.bid.twentyfive")
	default:
		return i18n.T("fortyfives.bid.pass")
	}
}

// fortyFivesTrumpStr renders the trump glyph, or a "no trump" label during bidding.
func fortyFivesTrumpStr(suit int) string {
	if suit < domain.CardDesignSpade {
		return i18n.T("fortyfives.noTrump")
	}
	return cuiSuitName(suit)
}

// fortyFivesPlayerStr returns the display string for a single player.
func fortyFivesPlayerStr(g interfaces.FortyFivesGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	scores := g.GetTeamScores()
	team := domain.FortyFivesTeamOf(idx)
	role := i18n.T("fortyfives.roleDefender")
	if idx == g.GetDeclarerIdx() {
		role = i18n.T("fortyfives.roleDeclarer")
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("fortyfives.playerLine",
		"name", cuiPlayerName(player, idx),
		"team", fortyFivesTeamLabel(team),
		"role", role,
		"cards", strconv.Itoa(player.GetCardsSize()),
		"score", strconv.Itoa(scores[team]),
		"tricks", strconv.Itoa(player.GetTrickCount()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// FortyFivesCuiPresenter renders the Auction Forty-Fives CUI view.
type FortyFivesCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *FortyFivesCuiPresenter) Output(g interfaces.FortyFivesGame, lastErr error) string {
	return buildCuiOutput(i18n.T("fortyfives.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("fortyfives.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber()),
			"trump", fortyFivesTrumpStr(g.GetTrumpSuit())) + "\n")

		scores := g.GetTeamScores()
		b.WriteString(i18n.Tf("fortyfives.teamScores",
			"teamA", strconv.Itoa(scores[0]),
			"teamB", strconv.Itoa(scores[1])) + "\n")

		if g.GetDeclarerIdx() >= 0 {
			declIdx := g.GetDeclarerIdx()
			b.WriteString(i18n.Tf("fortyfives.contractLine",
				"name", cuiPlayerName(g.GetPlayer(declIdx), declIdx),
				"team", fortyFivesTeamLabel(domain.FortyFivesTeamOf(declIdx)),
				"contract", fortyFivesBidName(int(g.GetContract()))) + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(fortyFivesPlayerStr(g, i))
		}

		b.WriteString("----------\n")

		cuiTrickBlock(b, g.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winner := g.GetWinnerTeam()
			banner := i18n.Tf("fortyfives.gameEnd", "team", fortyFivesTeamLabel(winner))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		p.writePrompt(b, g)
	})
}

// writePrompt renders the phase-specific prompt block.
func (p *FortyFivesCuiPresenter) writePrompt(b *strings.Builder, g interfaces.FortyFivesGame) {
	switch g.GetPhase() {
	case domain.FortyFivesPhaseBid:
		currentIdx := g.GetCurrentPlayerIdx()
		bids := g.GetBids()
		highBid := domain.FortyFivesBidPass
		for _, bid := range bids {
			if bid > highBid {
				highBid = bid
			}
		}
		b.WriteString(i18n.Tf("fortyfives.promptBid",
			"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx),
			"high", fortyFivesBidName(int(highBid))) + "\n")
		// List every player's declared bid (pass included) so the opposing team's
		// bidding is visible; players who have not yet bid show "-".
		bidDone := g.GetBidDone()
		entries := make([]string, g.GetPlayerCnt())
		for i := 0; i < g.GetPlayerCnt(); i++ {
			state := "-"
			if bidDone[i] {
				state = fortyFivesBidName(int(bids[i]))
			}
			entries[i] = cuiPlayerName(g.GetPlayer(i), i) + "=" + state
		}
		b.WriteString(i18n.Tf("fortyfives.bidHistory",
			"bids", strings.Join(entries, ", ")) + "\n")
		b.WriteString(i18n.T("fortyfives.promptBidHelp") + "\n")
	case domain.FortyFivesPhasePlay:
		// **契約の進捗はラウンドが終わるまで一切出ていなかった (#4724)。**
		// Web は ff-contract-progress に「あと何点必要か」を常時出している。
		// 落札チームが何点必要かは、降りるか押すかの判断そのもの。
		writeFortyFivesContractProgress(b, g)
		currentIdx := g.GetCurrentPlayerIdx()
		b.WriteString(i18n.Tf("fortyfives.promptPlay",
			"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
		b.WriteString(i18n.T("fortyfives.promptPlayHelp") + "\n")
	case domain.FortyFivesPhaseTrickEnd:
		b.WriteString(i18n.T("fortyfives.promptTrickEnd") + "\n")
		b.WriteString(i18n.T("fortyfives.promptTrickEndHelp") + "\n")
	case domain.FortyFivesPhaseRoundEnd:
		b.WriteString(i18n.T("fortyfives.promptRoundEnd") + "\n")
		b.WriteString(i18n.T("fortyfives.promptRoundEndHelp") + "\n")
	}
}

// fortyFivesContractStatusKeys は進捗ステータスから i18n キーへの対応。
var fortyFivesContractStatusKeys = map[string]string{
	domain.FortyFivesContractMade:     "fortyfives.contractMade",
	domain.FortyFivesContractFailed:   "fortyfives.contractFailed",
	domain.FortyFivesContractNeedMore: "fortyfives.contractNeedMore",
}

// writeFortyFivesContractProgress は落札チームの契約進捗を1行で書く。
// 落札が決まっていなければ何も書かない。
func writeFortyFivesContractProgress(b *strings.Builder, g interfaces.FortyFivesGame) {
	pr := g.GetContractProgress()
	if pr == nil {
		return
	}
	key, ok := fortyFivesContractStatusKeys[pr.Status]
	if !ok {
		return
	}
	team := i18n.T("fortyfives.teamA")
	if pr.DeclarerTeam == 1 {
		team = i18n.T("fortyfives.teamB")
	}
	b.WriteString(i18n.Tf("fortyfives.contractProgress",
		"team", team,
		"got", strconv.Itoa(pr.Points),
		"contract", strconv.Itoa(pr.Contract),
		"status", i18n.Tf(key, "remaining", strconv.Itoa(pr.Remaining))) + "\n")
}

// HintOutput emits the current Forty-Fives hint.
func (p *FortyFivesCuiPresenter) HintOutput(g interfaces.FortyFivesGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("fortyfives.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, fortyFivesHintReasonKeys)
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
		return color.Yellow(i18n.Tf("fortyfives.hintCard",
			"cards", strings.Join(cards, ", "),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("fortyfives.hintCard",
		"cards", "-",
		"reason", reason)) + "\n"
}

// fortyFivesHintReasonKeys maps Forty-Fives-specific hint-reason identifiers to i18n keys.
var fortyFivesHintReasonKeys = map[string]string{
	"lead_high":   "fortyfives.hintReasonLeadHigh",
	"take_trick":  "fortyfives.hintReasonTakeTrick",
	"discard_low": "fortyfives.hintReasonDiscardLow",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *FortyFivesCuiPresenter) ActionLogOutput(g interfaces.FortyFivesGame) string {
	return actionLogOutputText(g)
}
