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

// belotePlayerStr returns the display string for a single Belote player.
func belotePlayerStr(player *domain.BelotePlayer, i int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("belote.playerLine",
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

// BeloteCuiPresenter renders the Belote CUI view.
type BeloteCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *BeloteCuiPresenter) Output(b interfaces.BeloteGame, lastErr error) string {
	return buildCuiOutput(i18n.T("belote.helpTitle"), func(out *strings.Builder) {
		out.WriteString(i18n.Tf("belote.header",
			"round", strconv.Itoa(b.GetRoundNumber()),
			"trick", strconv.Itoa(b.GetTrickNumber())) + "\n")
		// **最終トリックは特別だと、点数計算を見る前に知らせる。**Web は
		// バッジを点滅させているのに、CUI は自分の手番が最後かどうかも
		// ボーナスの存在も言っていなかった (#5592)。点数は設定から取る ──
		// 訳文に 10 と書くと、設定を変えたとき案内だけが嘘になる。
		if dd := b.GetConfig().DixDeDer; dd > 0 && b.GetTrickNumber() == domain.BeloteHandSize {
			out.WriteString(color.Yellow(i18n.Tf("belote.dixDeDerNotice",
				"points", strconv.Itoa(dd))) + "\n")
		}

		dealerIdx := b.GetDealerIdx()
		out.WriteString(i18n.Tf("belote.dealer",
			"name", cuiPlayerName(b.GetPlayer(dealerIdx), dealerIdx)) + "\n")

		if trumpSuit := b.GetTrumpSuit(); trumpSuit > 0 {
			out.WriteString(i18n.Tf("belote.trumpLine",
				"suit", cuiSuitName(trumpSuit),
				"team", strconv.Itoa(b.GetMakerTeam())) + "\n")
		} else {
			out.WriteString(i18n.T("belote.trumpUndecided") + "\n")
		}

		if faceUpCard := b.GetFaceUpCard(); faceUpCard != nil {
			out.WriteString(i18n.Tf("belote.faceUpCard", "card", cuiCardStr(faceUpCard)) + "\n")
		}

		out.WriteString(i18n.Tf("belote.teamScoreLine",
			"t0", strconv.Itoa(b.GetTeamScore(0)),
			"t1", strconv.Itoa(b.GetTeamScore(1))) + "\n")

		// **20 点規模のボーナスに気づけない。**Web は専用バッジと読み上げまで
		// 用意しているのに、CUI は累計点しか出しておらず、Belote/Rebelote が
		// 成立したこと自体が伝わっていなかった (#4913)。
		for team := range domain.BeloteTeamCnt {
			if bonus := b.GetRoundBeloteBonus(team); bonus > 0 {
				out.WriteString(i18n.Tf("belote.beloteBonusLine",
					"team", strconv.Itoa(team),
					"points", strconv.Itoa(bonus)) + "\n")
			}
		}

		for i := 0; i < b.GetPlayerCnt(); i++ {
			out.WriteString(belotePlayerStr(b.GetPlayer(i), i))
		}

		out.WriteString("----------\n")

		trick := b.GetCurrentTrick()
		cuiTrickBlock(out, trick,
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(b.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(out, lastErr)

		if b.GetGameEndFlag() {
			banner := i18n.Tf("belote.gameEnd", "team", strconv.Itoa(b.GetWinnerTeam()))
			out.WriteString(color.Green(banner) + "\n")
			return
		}
		switch b.GetPhase() {
		case domain.BelotePhaseBidPickUp:
			bidIdx := b.GetBidPlayerIdx()
			out.WriteString(i18n.Tf("belote.promptPickup",
				"name", cuiPlayerName(b.GetPlayer(bidIdx), bidIdx)) + "\n")
			out.WriteString(i18n.T("belote.promptPickupHelp") + "\n")
		case domain.BelotePhaseBidCallTrump:
			bidIdx := b.GetBidPlayerIdx()
			out.WriteString(i18n.Tf("belote.promptCallTrump",
				"name", cuiPlayerName(b.GetPlayer(bidIdx), bidIdx)) + "\n")
			out.WriteString(i18n.T("belote.promptCallTrumpHelp") + "\n")
		case domain.BelotePhasePlay:
			currentIdx := b.GetCurrentPlayerIdx()
			out.WriteString(i18n.Tf("belote.promptCurrentPlayer",
				"name", cuiPlayerName(b.GetPlayer(currentIdx), currentIdx)) + "\n")
			out.WriteString(i18n.T("belote.promptPlayHelp") + "\n")
		case domain.BelotePhaseTrickEnd:
			out.WriteString(i18n.T("belote.promptTrickEnd") + "\n")
			out.WriteString(i18n.T("belote.promptTrickEndHelp") + "\n")
		case domain.BelotePhaseRoundEnd:
			out.WriteString(i18n.T("belote.promptRoundEnd") + "\n")
			out.WriteString(i18n.T("belote.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Belote hint.
func (p *BeloteCuiPresenter) HintOutput(b interfaces.BeloteGame) string {
	hint := b.GetHint()
	if hint == nil {
		return i18n.T("belote.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, nil)
	if hint.OrderUp != nil {
		if *hint.OrderUp {
			return color.Yellow(i18n.Tf("belote.hintOrderUp", "reason", reason)) + "\n"
		}
		return color.Yellow(i18n.Tf("belote.hintPass", "reason", reason)) + "\n"
	}
	if hint.Suit != nil {
		return color.Yellow(i18n.Tf("belote.hintCallSuit",
			"suit", cuiSuitName(*hint.Suit),
			"reason", reason)) + "\n"
	}
	if hint.CardIndex == nil {
		return i18n.T("belote.hintNone") + "\n"
	}
	player := b.GetPlayer(0)
	card := player.GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("belote.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", reason)) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *BeloteCuiPresenter) ActionLogOutput(b interfaces.BeloteGame) string {
	return actionLogOutputTextWithNames(b, func(idx int) string { return cuiPlayerName(b.GetPlayer(idx), idx) })
}
