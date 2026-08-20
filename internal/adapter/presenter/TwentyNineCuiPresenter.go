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

// twentyNineTeamLabels maps a team index (0/1) to its display name.
var twentyNineTeamLabels = [domain.TwentyNineTeamCnt]string{"A", "B"}

// twentyNineTeamLabel returns the localized team label (A/B) for a team index.
func twentyNineTeamLabel(team int) string {
	if team < 0 || team >= len(twentyNineTeamLabels) {
		return "?"
	}
	return twentyNineTeamLabels[team]
}

// twentyNineBidName maps a bid constant (0/16/20/24/28) to its localized contract name.
func twentyNineBidName(bid int) string {
	switch domain.TwentyNineBid(bid) {
	case domain.TwentyNineBidSixteen:
		return i18n.T("twentynine.bid.sixteen")
	case domain.TwentyNineBidTwenty:
		return i18n.T("twentynine.bid.twenty")
	case domain.TwentyNineBidTwentyFour:
		return i18n.T("twentynine.bid.twentyfour")
	case domain.TwentyNineBidTwentyEight:
		return i18n.T("twentynine.bid.twentyeight")
	default:
		return i18n.T("twentynine.bid.pass")
	}
}

// twentyNineTrumpStr renders the trump glyph only when the hidden trump has been
// revealed; otherwise it shows a "hidden" label (29 uses a hidden trump).
func twentyNineTrumpStr(suit int, revealed bool) string {
	if !revealed || suit < domain.CardDesignSpade {
		return i18n.T("twentynine.hiddenTrump")
	}
	return cuiSuitName(suit)
}

// twentyNinePlayerStr returns the display string for a single player.
func twentyNinePlayerStr(g interfaces.TwentyNineGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	scores := g.GetTeamScores()
	team := domain.TwentyNineTeamOf(idx)
	role := i18n.T("twentynine.roleDefender")
	if idx == g.GetDeclarerIdx() {
		role = i18n.T("twentynine.roleDeclarer")
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("twentynine.playerLine",
		"name", cuiPlayerName(player, idx),
		"team", twentyNineTeamLabel(team),
		"role", role,
		"cards", strconv.Itoa(player.GetCardsSize()),
		"score", strconv.Itoa(scores[team]),
		"tricks", strconv.Itoa(player.GetTrickCount()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		// **Web は playableIndices をリング表示しているのに、CUI は素の一覧だけで、
		// 番号を入力してエラーを踏むまで合法手が分からなかった (#4725)。**
		// 目印を出すのはプレイフェーズでこのプレイヤーの手番のときだけ -- ビッド中や
		// 相手の手番では制限そのものが決まっていない。
		var playable []int
		if g.GetPhase() == domain.TwentyNinePhasePlay && g.GetCurrentPlayerIdx() == idx {
			playable = g.GetPlayableIndices(idx)
		}
		b.WriteString(cuiPlayableMarkedCardListStr(player, playable) + "\n")
	}
	return b.String()
}

// TwentyNineCuiPresenter renders the Twenty-Nine (29) CUI view.
type TwentyNineCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *TwentyNineCuiPresenter) Output(g interfaces.TwentyNineGame, lastErr error) string {
	return buildCuiOutput(i18n.T("twentynine.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("twentynine.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber()),
			"trump", twentyNineTrumpStr(g.GetTrumpSuit(), g.GetTrumpRevealed())) + "\n")

		scores := g.GetTeamScores()
		b.WriteString(i18n.Tf("twentynine.teamScores",
			"teamA", strconv.Itoa(scores[0]),
			"teamB", strconv.Itoa(scores[1])) + "\n")

		points := g.GetRoundTeamPoints()
		b.WriteString(i18n.Tf("twentynine.roundPoints",
			"teamA", strconv.Itoa(points[0]),
			"teamB", strconv.Itoa(points[1])) + "\n")

		if g.GetDeclarerIdx() >= 0 {
			declIdx := g.GetDeclarerIdx()
			b.WriteString(i18n.Tf("twentynine.contractLine",
				"name", cuiPlayerName(g.GetPlayer(declIdx), declIdx),
				"team", twentyNineTeamLabel(domain.TwentyNineTeamOf(declIdx)),
				"contract", twentyNineBidName(int(g.GetContract()))) + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(twentyNinePlayerStr(g, i))
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
			banner := i18n.Tf("twentynine.gameEnd", "team", twentyNineTeamLabel(winner))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		p.writePrompt(b, g)
	})
}

// writePrompt renders the phase-specific prompt block.
func (p *TwentyNineCuiPresenter) writePrompt(b *strings.Builder, g interfaces.TwentyNineGame) {
	switch g.GetPhase() {
	case domain.TwentyNinePhaseBid:
		currentIdx := g.GetCurrentPlayerIdx()
		bids := g.GetBids()
		highBid := domain.TwentyNineBidPass
		for _, bid := range bids {
			if bid > highBid {
				highBid = bid
			}
		}
		b.WriteString(i18n.Tf("twentynine.promptBid",
			"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx),
			"high", twentyNineBidName(int(highBid))) + "\n")
		b.WriteString(i18n.T("twentynine.promptBidHelp") + "\n")
	case domain.TwentyNinePhasePlay:
		writeTwentyNineContractProgress(b, g)
		currentIdx := g.GetCurrentPlayerIdx()
		b.WriteString(i18n.Tf("twentynine.promptPlay",
			"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
		b.WriteString(i18n.T("twentynine.promptPlayHelp") + "\n")
	case domain.TwentyNinePhaseTrickEnd:
		// Web は TrickEnd でも進捗パネルを出している。ここで消すと、トリックの
		// 合間だけ同じ局面が別物に見える (PR #6011 のレビュー指摘)。
		writeTwentyNineContractProgress(b, g)
		b.WriteString(i18n.T("twentynine.promptTrickEnd") + "\n")
		b.WriteString(i18n.T("twentynine.promptTrickEndHelp") + "\n")
	case domain.TwentyNinePhaseRoundEnd:
		b.WriteString(i18n.T("twentynine.promptRoundEnd") + "\n")
		b.WriteString(i18n.T("twentynine.promptRoundEndHelp") + "\n")
	}
}

// twentyNineContractStatusKeys は進捗ステータスから i18n キーへの対応。
var twentyNineContractStatusKeys = map[string]string{
	domain.TwentyNineContractMade:     "twentynine.contractMade",
	domain.TwentyNineContractFailed:   "twentynine.contractFailed",
	domain.TwentyNineContractNeedMore: "twentynine.contractNeedMore",
}

// writeTwentyNineContractProgress は落札チームの契約進捗を1行で書く。
// 落札が決まっていなければ何も書かない (#5644)。
func writeTwentyNineContractProgress(b *strings.Builder, g interfaces.TwentyNineGame) {
	pr := g.GetContractProgress()
	if pr == nil {
		return
	}
	key, ok := twentyNineContractStatusKeys[pr.Status]
	if !ok {
		return
	}
	team := i18n.T("twentynine.teamA")
	if pr.DeclarerTeam == 1 {
		team = i18n.T("twentynine.teamB")
	}
	b.WriteString(i18n.Tf("twentynine.contractProgress",
		"team", team,
		"got", strconv.Itoa(pr.Points),
		"contract", strconv.Itoa(pr.Contract),
		"status", i18n.Tf(key, "remaining", strconv.Itoa(pr.Remaining))) + "\n")
}

// HintOutput emits the current Twenty-Nine hint.
func (p *TwentyNineCuiPresenter) HintOutput(g interfaces.TwentyNineGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("twentynine.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, twentyNineHintReasonKeys)
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
		return color.Yellow(i18n.Tf("twentynine.hintCard",
			"cards", strings.Join(cards, ", "),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("twentynine.hintCard",
		"cards", "-",
		"reason", reason)) + "\n"
}

// twentyNineHintReasonKeys maps Twenty-Nine-specific hint-reason identifiers to i18n keys.
var twentyNineHintReasonKeys = map[string]string{
	"lead_low":    "twentynine.hintReasonLeadLow",
	"follow_win":  "twentynine.hintReasonFollowWin",
	"follow_duck": "twentynine.hintReasonFollowDuck",
	"discard_low": "twentynine.hintReasonDiscardLow",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *TwentyNineCuiPresenter) ActionLogOutput(g interfaces.TwentyNineGame) string {
	return actionLogOutputTextWithNames(g, func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) })
}
