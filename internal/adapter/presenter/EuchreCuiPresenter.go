//go:build !js || !wasm || solo

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// euchrePlayerStr returns the display string for a single Euchre player.
// sittingOut appends a marker when this seat sits out a go-alone hand.
// playable が非 nil のとき、そのインデックスの札に "*" を付ける。
//
// **レフトボーア (同色の別スートの J) が切り札扱いになる**という分かりにくい
// ルールを含むので、CUI では自力でルールを再現するか、出せない札を選んで
// エラーを受け取るまで気づけなかった (#4781)。Web は同じ判定で合法な札に
// 枠線を付けている。
func euchrePlayerStr(player *domain.EuchrePlayer, i int, sittingOut bool, playable []int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("euchre.playerLine",
		"name", cuiPlayerName(player, i),
		"team", strconv.Itoa(player.GetTeam()),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	if sittingOut {
		b.WriteString(i18n.T("euchre.sittingOut"))
	}
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(euchreHandStr(player, playable) + "\n")
	}
	return b.String()
}

// euchreHandStr renders the hand as an indexed list, starring the cards that
// may legally be played right now.
func euchreHandStr(player *domain.EuchrePlayer, playable []int) string {
	if len(playable) == 0 {
		return cuiIndexedCardListStr(player)
	}
	mark := make(map[int]bool, len(playable))
	for _, idx := range playable {
		mark[idx] = true
	}
	parts := make([]string, player.GetCardsSize())
	for i := 0; i < player.GetCardsSize(); i++ {
		parts[i] = "[" + strconv.Itoa(i) + "]" + cuiCardStr(player.GetCard(i))
		if mark[i] {
			parts[i] += "*"
		}
	}
	return strings.Join(parts, "  ")
}

// euchrePlayableIndices returns the human's legal plays, or nil when it is not
// the human's turn to play a card.
//
// **手番でないときは印を付けない。**別のプレイヤーの合法手を人間の手札に
// 当てると、出せない札に印が付く。
func euchrePlayableIndices(e interfaces.EuchreGame) []int {
	if e.GetPhase() != domain.EuchrePhasePlay {
		return nil
	}
	idx := e.GetCurrentPlayerIdx()
	p := e.GetPlayer(idx)
	if p == nil || !p.GetIsHuman() {
		return nil
	}
	return e.GetValidPlayIndices(idx)
}

// euchreSittingOutIdx returns the seat that sits out during a go-alone hand —
// the lone player's same-team partner — or -1 when nobody sits out. Mirrors the
// web euchreSittingOutIdx helper so the CUI matches the graphical UI.
func euchreSittingOutIdx(e interfaces.EuchreGame) int {
	if !e.GetGoingAlone() {
		return -1
	}
	goer := e.GetGoingAlonePlayerIdx()
	gp := e.GetPlayer(goer)
	if gp == nil {
		return -1
	}
	for i := 0; i < e.GetPlayerCnt(); i++ {
		if i == goer {
			continue
		}
		if p := e.GetPlayer(i); p != nil && p.GetTeam() == gp.GetTeam() {
			return i
		}
	}
	return -1
}

// EuchreCuiPresenter renders the Euchre CUI view.
type EuchreCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *EuchreCuiPresenter) Output(e interfaces.EuchreGame, lastErr error) string {
	return buildCuiOutput(i18n.T("euchre.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("euchre.header",
			"round", strconv.Itoa(e.GetRoundNumber()),
			"trick", strconv.Itoa(e.GetTrickNumber())) + "\n")
		dealerIdx := e.GetDealerIdx()
		b.WriteString(i18n.Tf("euchre.dealer",
			"name", cuiPlayerName(e.GetPlayer(dealerIdx), dealerIdx)) + "\n")

		if trumpSuit := e.GetTrumpSuit(); trumpSuit > 0 {
			b.WriteString(i18n.Tf("euchre.trumpLine",
				"suit", cuiSuitName(trumpSuit),
				"team", strconv.Itoa(e.GetMakerTeam())) + "\n")
		} else {
			b.WriteString(i18n.T("euchre.trumpUndecided") + "\n")
		}

		if faceUpCard := e.GetFaceUpCard(); faceUpCard != nil {
			b.WriteString(i18n.Tf("euchre.faceUpCard", "card", cuiCardStr(faceUpCard)) + "\n")
		}

		if e.GetGoingAlone() {
			aloneIdx := e.GetGoingAlonePlayerIdx()
			b.WriteString(i18n.Tf("euchre.goingAlone",
				"name", cuiPlayerName(e.GetPlayer(aloneIdx), aloneIdx)) + "\n")
		}

		b.WriteString(i18n.Tf("euchre.teamScoreLine",
			"t0", strconv.Itoa(e.GetTeamScore(0)),
			"t1", strconv.Itoa(e.GetTeamScore(1))) + "\n")

		sittingOutIdx := euchreSittingOutIdx(e)
		playable := euchrePlayableIndices(e)
		for i := 0; i < e.GetPlayerCnt(); i++ {
			b.WriteString(euchrePlayerStr(e.GetPlayer(i), i, i == sittingOutIdx, playable))
		}

		b.WriteString("----------\n")

		// Current trick
		trick := e.GetCurrentTrick()
		cuiTrickBlock(b, trick,
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(e.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		// Game state
		if e.GetGameEndFlag() {
			banner := i18n.Tf("euchre.gameEnd", "team", strconv.Itoa(e.GetWinnerTeam()))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch e.GetPhase() {
		case domain.EuchrePhasePickUp:
			bidIdx := e.GetBidPlayerIdx()
			b.WriteString(i18n.Tf("euchre.promptPickup",
				"name", cuiPlayerName(e.GetPlayer(bidIdx), bidIdx)) + "\n")
			b.WriteString(i18n.T("euchre.promptPickupHelp") + "\n")
		case domain.EuchrePhaseCallTrump:
			bidIdx := e.GetBidPlayerIdx()
			b.WriteString(i18n.Tf("euchre.promptCallTrump",
				"name", cuiPlayerName(e.GetPlayer(bidIdx), bidIdx)) + "\n")
			b.WriteString(i18n.T("euchre.promptCallTrumpHelp") + "\n")
		case domain.EuchrePhaseDiscard:
			b.WriteString(i18n.T("euchre.promptDiscard") + "\n")
			b.WriteString(i18n.T("euchre.promptDiscardHelp") + "\n")
		case domain.EuchrePhasePlay:
			currentIdx := e.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("euchre.promptCurrentPlayer",
				"name", cuiPlayerName(e.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("euchre.promptPlayHelp") + "\n")
		case domain.EuchrePhaseTrickEnd:
			b.WriteString(i18n.T("euchre.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("euchre.promptTrickEndHelp") + "\n")
		case domain.EuchrePhaseRoundEnd:
			b.WriteString(i18n.T("euchre.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("euchre.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Euchre hint.
func (p *EuchreCuiPresenter) HintOutput(e interfaces.EuchreGame) string {
	hint := e.GetHint()
	if hint == nil {
		return i18n.T("euchre.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, euchreHintReasonKeys)
	// **判断の元になった数値も出す。** 真偽値と定型の理由キーだけでは、自分の手札が
	// CPU の基準にどれだけ近いのか分からない (#5509)。しきい値も一緒に言う。
	scoreLine := ""
	if hint.Score != nil {
		scoreLine = i18n.Tf("euchre.hintScore",
			"score", strconv.Itoa(*hint.Score),
			"orderUp", strconv.Itoa(domain.EuchreOrderUpScore),
			"goAlone", strconv.Itoa(domain.EuchreGoAloneScore)) + "\n"
	}
	if hint.OrderUp != nil {
		if *hint.OrderUp {
			key := "euchre.hintOrderUp"
			if hint.GoAlone != nil && *hint.GoAlone {
				key = "euchre.hintOrderUpAlone"
			}
			return color.Yellow(i18n.Tf(key, "reason", reason)) + "\n" + scoreLine
		}
		return color.Yellow(i18n.Tf("euchre.hintPass", "reason", reason)) + "\n" + scoreLine
	}
	if hint.Suit != nil {
		key := "euchre.hintCallSuit"
		if hint.GoAlone != nil && *hint.GoAlone {
			key = "euchre.hintCallSuitAlone"
		}
		return color.Yellow(i18n.Tf(key,
			"suit", cuiSuitName(*hint.Suit),
			"reason", reason)) + "\n" + scoreLine
	}
	if hint.CardIndex == nil {
		return i18n.T("euchre.hintNone") + "\n"
	}
	player := e.GetPlayer(0)
	card := player.GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("euchre.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", reason)) + "\n"
}

// euchreHintReasonKeys maps Euchre-specific hint-reason identifiers to their
// i18n keys. Reasons not listed here fall through to cui_common via
// hintReasonStr.
var euchreHintReasonKeys = map[string]string{
	"discard_weakest": "euchre.hintReasonDiscardWeakest",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *EuchreCuiPresenter) ActionLogOutput(e interfaces.EuchreGame) string {
	return actionLogOutputTextWithNames(e, func(idx int) string { return cuiPlayerName(e.GetPlayer(idx), idx) })
}
