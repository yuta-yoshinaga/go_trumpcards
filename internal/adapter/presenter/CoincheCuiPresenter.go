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

// coinchePlayerStr returns the display string for a single Coinche player.
func coinchePlayerStr(player *domain.CoinchePlayer, i int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("coinche.playerLine",
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

// CoincheCuiPresenter renders the Coinche CUI view.
type CoincheCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *CoincheCuiPresenter) Output(b interfaces.CoincheGame, lastErr error) string {
	return buildCuiOutput(i18n.T("coinche.helpTitle"), func(out *strings.Builder) {
		out.WriteString(i18n.Tf("coinche.header",
			"round", strconv.Itoa(b.GetRoundNumber()),
			"trick", strconv.Itoa(b.GetTrickNumber())) + "\n")
		// **最終トリックは特別だと、点数計算を見る前に知らせる。**Web は
		// バッジを点滅させているのに、CUI は自分の手番が最後かどうかも
		// ボーナスの存在も言っていなかった (#5592)。点数は設定から取る ──
		// 訳文に 10 と書くと、設定を変えたとき案内だけが嘘になる。
		if dd := b.GetConfig().DixDeDer; dd > 0 && b.GetTrickNumber() == domain.CoincheHandSize {
			out.WriteString(color.Yellow(i18n.Tf("coinche.dixDeDerNotice",
				"points", strconv.Itoa(dd))) + "\n")
		}

		dealerIdx := b.GetDealerIdx()
		out.WriteString(i18n.Tf("coinche.dealer",
			"name", cuiPlayerName(b.GetPlayer(dealerIdx), dealerIdx)) + "\n")

		if trumpSuit := b.GetTrumpSuit(); trumpSuit > 0 {
			out.WriteString(i18n.Tf("coinche.trumpLine",
				"suit", cuiSuitName(trumpSuit),
				"team", strconv.Itoa(b.GetMakerTeam())) + "\n")
		} else {
			out.WriteString(i18n.T("coinche.trumpUndecided") + "\n")
		}

		// **契約と倍率は精算そのもの。** 出さないと、同じカード点でも勝ち
		// 負けが変わる理由が盤面から読めない。
		if pts := b.GetContractPoints(); pts > 0 {
			out.WriteString(i18n.Tf("coinche.contractLine",
				"points", strconv.Itoa(pts),
				"team", strconv.Itoa(b.GetMakerTeam()),
				"mult", strconv.Itoa(b.GetMultiplier())) + "\n")
		}

		out.WriteString(i18n.Tf("coinche.teamScoreLine",
			"t0", strconv.Itoa(b.GetTeamScore(0)),
			"t1", strconv.Itoa(b.GetTeamScore(1))) + "\n")

		// **20 点規模のボーナスに気づけない。**Web は専用バッジと読み上げまで
		// 用意しているのに、CUI は累計点しか出しておらず、Belote/Rebelote が
		// 成立したこと自体が伝わっていなかった (#4913)。
		for team := range domain.CoincheTeamCnt {
			if bonus := b.GetRoundBeloteBonus(team); bonus > 0 {
				out.WriteString(i18n.Tf("coinche.beloteBonusLine",
					"team", strconv.Itoa(team),
					"points", strconv.Itoa(bonus)) + "\n")
			}
		}

		for i := 0; i < b.GetPlayerCnt(); i++ {
			out.WriteString(coinchePlayerStr(b.GetPlayer(i), i))
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
			banner := i18n.Tf("coinche.gameEnd", "team", strconv.Itoa(b.GetWinnerTeam()))
			out.WriteString(color.Green(banner) + "\n")
			return
		}
		switch b.GetPhase() {
		case domain.CoinchePhaseBid:
			bidIdx := b.GetBidPlayerIdx()
			out.WriteString(i18n.Tf("coinche.promptBid",
				"name", cuiPlayerName(b.GetPlayer(bidIdx), bidIdx)) + "\n")
			// **打てる点だけを案内する。** 全部並べると、打てば必ず拒否
			// される値を勧めることになる。
			if pts := b.GetBiddablePoints(); len(pts) > 0 {
				labels := make([]string, len(pts))
				for i, v := range pts {
					labels[i] = strconv.Itoa(v)
				}
				out.WriteString(i18n.Tf("coinche.biddablePoints",
					"points", strings.Join(labels, " / ")) + "\n")
			}
			out.WriteString(i18n.T("coinche.promptBidHelp") + "\n")
		case domain.CoinchePhaseDouble:
			idx := b.GetCurrentPlayerIdx()
			out.WriteString(i18n.Tf("coinche.promptDouble",
				"name", cuiPlayerName(b.GetPlayer(idx), idx)) + "\n")
			out.WriteString(i18n.T("coinche.promptDoubleHelp") + "\n")
		case domain.CoinchePhasePlay:
			currentIdx := b.GetCurrentPlayerIdx()
			out.WriteString(i18n.Tf("coinche.promptCurrentPlayer",
				"name", cuiPlayerName(b.GetPlayer(currentIdx), currentIdx)) + "\n")
			out.WriteString(i18n.T("coinche.promptPlayHelp") + "\n")
		case domain.CoinchePhaseTrickEnd:
			out.WriteString(i18n.T("coinche.promptTrickEnd") + "\n")
			out.WriteString(i18n.T("coinche.promptTrickEndHelp") + "\n")
		case domain.CoinchePhaseRoundEnd:
			out.WriteString(i18n.T("coinche.promptRoundEnd") + "\n")
			out.WriteString(i18n.T("coinche.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Coinche hint.
func (p *CoincheCuiPresenter) HintOutput(b interfaces.CoincheGame) string {
	hint := b.GetHint()
	if hint == nil {
		return i18n.T("coinche.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, nil)
	if hint.Bid != nil && hint.Suit != nil {
		return color.Yellow(i18n.Tf("coinche.hintBid",
			"points", strconv.Itoa(*hint.Bid),
			"suit", cuiSuitName(*hint.Suit),
			"reason", reason)) + "\n"
	}
	if hint.CardIndex == nil && hint.Reason != "" {
		return color.Yellow(i18n.Tf("coinche.hintPass", "reason", reason)) + "\n"
	}
	if hint.CardIndex == nil {
		return i18n.T("coinche.hintNone") + "\n"
	}
	player := b.GetPlayer(0)
	card := player.GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("coinche.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", reason)) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *CoincheCuiPresenter) ActionLogOutput(b interfaces.CoincheGame) string {
	return actionLogOutputTextForSeats[*domain.CoinchePlayer](b)
}
