package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// pitchBidStr renders the bid column for the player line.
func pitchBidStr(bid int) string {
	switch {
	case bid == 0:
		return i18n.T("pitch.bidPass")
	case bid > 0:
		return strconv.Itoa(bid)
	default:
		return i18n.T("pitch.bidPending")
	}
}

// pitchPlayerStr returns the display string for a single Pitch player.
func pitchPlayerStr(player *domain.PitchPlayer, i int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("pitch.playerLine",
		"name", cuiPlayerName(player, i),
		"bid", pitchBidStr(player.GetBid()),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player))
		b.WriteString("\n")
	}
	return b.String()
}

// pitchSuitName renders the trump suit using a Unicode glyph. Returns the
// localized "(undecided)" placeholder when no trump has been chosen yet.
func pitchSuitName(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return "♠"
	case domain.CardDesignClover:
		return "♣"
	case domain.CardDesignHeart:
		return "♥"
	case domain.CardDesignDiamond:
		return "♦"
	}
	return i18n.T("pitch.trumpUndecided")
}

// PitchCuiPresenter renders the Pitch CUI view.
type PitchCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
// pitchBreakdownBlock は High/Low/Jack/Game をそれぞれ誰が取ったかを 1 行に並べる。
// 誰も取っていないカテゴリは「なし」と出す ── 黙って省くと、そのラウンドで
// そもそも争われなかったのか見落としたのか分からない。
func pitchBreakdownBlock(g interfaces.PitchGame) string {
	bd := g.GetRoundBreakdown()
	name := func(idx int) string {
		if idx == domain.PitchNoScorer {
			return i18n.T("pitch.scoringNobody")
		}
		return cuiPlayerName(g.GetPlayer(idx), idx)
	}
	return i18n.Tf("pitch.scoringLine",
		"high", name(bd.High),
		"low", name(bd.Low),
		"jack", name(bd.Jack),
		"game", name(bd.Game))
}

func (p *PitchCuiPresenter) Output(s interfaces.PitchGame, lastErr error) string {
	return buildCuiOutput(i18n.T("pitch.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("pitch.header",
			"round", strconv.Itoa(s.GetRoundNumber()),
			"trick", strconv.Itoa(s.GetTrickNumber()),
			"dealer", cuiPlayerName(s.GetPlayer(s.GetDealerIdx()), s.GetDealerIdx())) + "\n")
		b.WriteString(i18n.Tf("pitch.bidLine",
			"bid", strconv.Itoa(s.GetCurrentBid()),
			"suit", pitchSuitName(s.GetTrumpSuit())) + "\n")
		if s.GetBidWinnerIdx() >= 0 {
			b.WriteString(i18n.Tf("pitch.bidWinner",
				"name", cuiPlayerName(s.GetPlayer(s.GetBidWinnerIdx()), s.GetBidWinnerIdx())) + "\n")
		}

		// Player rows
		for i := 0; i < s.GetPlayerCnt(); i++ {
			b.WriteString(pitchPlayerStr(s.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		// Current trick
		trick := s.GetCurrentTrick()
		cuiTrickBlock(b, trick,
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(s.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		// Game state
		if s.GetGameEndFlag() {
			winnerIdx := s.GetWinnerIdx()
			banner := i18n.Tf("pitch.gameEnd",
				"name", cuiPlayerName(s.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch s.GetPhase() {
		case domain.PitchPhaseBid:
			bidIdx := s.GetBidPlayerIdx()
			b.WriteString(i18n.Tf("pitch.promptBid",
				"name", cuiPlayerName(s.GetPlayer(bidIdx), bidIdx)) + "\n")
			// **入札前に手札の得点価値を暗算させていた (#4751)。**Web は入札中に
			// ゲーム得点バッジと内訳を出している。強気に入札してよいかの判断材料。
			writePitchHandPips(b, s)
			b.WriteString(i18n.T("pitch.promptBidHelp") + "\n")
		case domain.PitchPhasePlay:
			currentIdx := s.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("pitch.promptPlay",
				"name", cuiPlayerName(s.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("pitch.promptPlayHelp") + "\n")
		case domain.PitchPhaseTrickEnd:
			b.WriteString(i18n.T("pitch.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("pitch.promptTrickEndHelp") + "\n")
		case domain.PitchPhaseRoundEnd:
			// **4 種の得点がこのゲームの骨格。**合計だけでは、1 点差の理由が
			// Jack を取られたからなのか Game で並ばれたからなのか分からない
			// (#5584)。獲得者はドメインが得点と同時に記録している。
			b.WriteString(pitchBreakdownBlock(s) + "\n")
			b.WriteString(i18n.T("pitch.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("pitch.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Pitch hint.
func (p *PitchCuiPresenter) HintOutput(s interfaces.PitchGame) string {
	hint := s.GetHint()
	if hint == nil {
		return i18n.T("pitch.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, pitchHintReasonKeys)
	if hint.Bid != nil {
		return color.Yellow(i18n.Tf("pitch.hintBid",
			"bid", pitchBidStr(*hint.Bid),
			"reason", reason)) + "\n"
	}
	if hint.CardIndex == nil {
		return i18n.T("pitch.hintNone") + "\n"
	}
	humanIdx := -1
	for i := 0; i < s.GetPlayerCnt(); i++ {
		if pl := s.GetPlayer(i); pl != nil && pl.GetIsHuman() {
			humanIdx = i
			break
		}
	}
	if humanIdx < 0 {
		return i18n.T("pitch.hintNone") + "\n"
	}
	card := s.GetPlayer(humanIdx).GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("pitch.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", reason)) + "\n"
}

// pitchHintReasonKeys maps Pitch-specific hint-reason identifiers to their
// i18n keys. Reasons not in this map fall through to hintReasonStr →
// cui_common.
var pitchHintReasonKeys = map[string]string{
	"set_trump_lead": "pitch.hintReasonSetTrumpLead",
	"trump_cut":      "pitch.hintReasonTrumpCut",
	"bid_strong":     "pitch.hintReasonBidStrong",
	"bid_pass":       "pitch.hintReasonBidPass",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *PitchCuiPresenter) ActionLogOutput(s interfaces.PitchGame) string {
	return actionLogOutputText(s)
}

// writePitchHandPips は人間の手札のゲームピップ合計と内訳を書く。
// 人間が見つからない、または手札が空なら何も書かない。
func writePitchHandPips(b *strings.Builder, s interfaces.PitchGame) {
	for i := 0; i < s.GetPlayerCnt(); i++ {
		pl := s.GetPlayer(i)
		if pl == nil || !pl.GetIsHuman() || pl.GetCardsSize() == 0 {
			continue
		}
		cards := make([]*domain.Card, 0, pl.GetCardsSize())
		// **0点の札も内訳に並べる。**並べないと「見落としている札がある」
		// のか「その札が0点」なのか区別が付かない。
		parts := make([]string, 0, pl.GetCardsSize())
		for j := 0; j < pl.GetCardsSize(); j++ {
			c := pl.GetCard(j)
			cards = append(cards, c)
			parts = append(parts, cuiCardStr(c)+"="+strconv.Itoa(domain.PitchHandPips([]*domain.Card{c})))
		}
		b.WriteString(i18n.Tf("pitch.handPips",
			"total", strconv.Itoa(domain.PitchHandPips(cards)),
			"breakdown", strings.Join(parts, " ")) + "\n")
		return
	}
}
