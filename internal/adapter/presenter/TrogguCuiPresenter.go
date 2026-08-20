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

// TrogguCuiPresenter renders the Troggu CUI view.
type TrogguCuiPresenter struct{}

// trogguIndexedHand 人間の手札をインデックス付きで表示する。
func trogguIndexedHand(p *domain.TrogguPlayer) string {
	parts := make([]string, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		parts[i] = "[" + strconv.Itoa(i) + "]" + frenchTarotCuiCardStr(p.GetCard(i))
	}
	return strings.Join(parts, "  ")
}

// trogguContractLabel 契約の i18n ラベルを返す。
// trogguAvailableBids は今の最高入札を上回れる契約のコマンド名を並べる。
//
// 空にはならない: ミゼールが最高でも「もう入札できない」ことを言えるように、
// 呼び出し側の文言はパスだけを案内する形にしてある。
func trogguAvailableBids(highest domain.TrogguBid) string {
	names := make([]string, 0, 4)
	for _, bid := range []domain.TrogguBid{
		domain.TrogguBidTrois, domain.TrogguBidSolo, domain.TrogguBidPiccolo, domain.TrogguBidMisere,
	} {
		if bid > highest {
			names = append(names, domain.TrogguBidName(bid))
		}
	}
	if len(names) == 0 {
		return i18n.T("troggu.noBidBeatsIt")
	}
	return strings.Join(names, "|")
}

func trogguContractLabel(bid domain.TrogguBid) string {
	return i18n.T("troggu.contract." + domain.TrogguBidName(bid))
}

// trogguPlayerStr 席 1 行ぶんの状態文字列を返す。
func trogguPlayerStr(g interfaces.TrogguGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	role := i18n.T("troggu.roleDefender")
	if idx == g.GetDeclarerIdx() {
		role = i18n.T("troggu.roleDeclarer")
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("troggu.playerLine",
		"name", cuiPlayerName(player, idx),
		"role", role,
		"cards", strconv.Itoa(player.GetCardsSize()),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"points", strconv.Itoa(g.GetCardPoints(idx)),
		"score", strconv.Itoa(g.GetPlayerScore(idx))) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(trogguIndexedHand(player) + "\n")
	}
	return b.String()
}

// Output renders the current game state for the active locale.
func (p *TrogguCuiPresenter) Output(g interfaces.TrogguGame, lastErr error) string {
	return buildCuiOutput(i18n.T("troggu.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("troggu.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"rounds", strconv.Itoa(g.GetConfig().TargetDeals),
			"trick", strconv.Itoa(g.GetTrickNumber()),
			"contract", trogguContractLabel(g.GetContract())) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(trogguPlayerStr(g, i))
		}
		b.WriteString("----------\n")

		cuiTrickBlock(b, g.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return frenchTarotCuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winner := g.GetWinnerPlayer()
			if winner < 0 {
				b.WriteString(color.Green(i18n.T("troggu.gameDraw")) + "\n")
				return
			}
			b.WriteString(color.Green(i18n.Tf("troggu.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(winner), winner))) + "\n")
			return
		}
		p.writePrompt(b, g)
	})
}

// writePrompt 現在のフェーズに応じたプロンプトを書き込む。
func (p *TrogguCuiPresenter) writePrompt(b *strings.Builder, g interfaces.TrogguGame) {
	switch g.GetPhase() {
	case domain.TrogguPhaseBid:
		idx := g.GetBidPlayerIdx()
		b.WriteString(i18n.Tf("troggu.promptBid",
			"name", cuiPlayerName(g.GetPlayer(idx), idx),
			"high", trogguContractLabel(g.GetHighestBid())) + "\n")
		// **超えられない契約は挙げない。**bid <= highestBid はドメインが却下する
		// ので、並べたままだと「打てるのに弾かれる」ように見える (#5808)。
		b.WriteString(i18n.Tf("troggu.promptBidHelp", "bids", trogguAvailableBids(g.GetHighestBid())) + "\n")
	case domain.TrogguPhasePlay:
		idx := g.GetCurrentPlayerIdx()
		b.WriteString(i18n.Tf("troggu.promptPlay",
			"name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
		b.WriteString(i18n.T("troggu.promptPlayHelp") + "\n")
	case domain.TrogguPhaseTrickEnd:
		winner := g.GetLastTrickWinner()
		b.WriteString(i18n.Tf("troggu.promptTrickEnd",
			"name", cuiPlayerName(g.GetPlayer(winner), winner)) + "\n")
		b.WriteString(i18n.T("troggu.promptTrickEndHelp") + "\n")
	case domain.TrogguPhaseRoundEnd:
		b.WriteString(p.roundResultStr(g) + "\n")
		b.WriteString(i18n.T("troggu.promptRoundEndHelp") + "\n")
	}
}

// roundResultStr ディール結果の説明文を返す。
//
// **契約ごとに見るものが違う。** ソロだけがカードポイント、他はトリック数なので、
// 同じ文型に押し込むと「3 点で成功」のような読めない結果表示になる。
func (p *TrogguCuiPresenter) roundResultStr(g interfaces.TrogguGame) string {
	bd := g.GetBreakdown()
	if bd == nil {
		return i18n.T("troggu.roundThrownIn")
	}
	declarer := g.GetDeclarerIdx()
	key := "troggu.roundLossPoints"
	if bd.TargetIsTricks {
		key = "troggu.roundLossTricks"
	}
	if bd.Won {
		key = "troggu.roundWinPoints"
		if bd.TargetIsTricks {
			key = "troggu.roundWinTricks"
		}
	}
	got := strconv.Itoa(bd.DeclarerPoints)
	if bd.TargetIsTricks {
		got = strconv.Itoa(bd.DeclarerTricks)
	}
	return i18n.Tf(key,
		"name", cuiPlayerName(g.GetPlayer(declarer), declarer),
		"contract", trogguContractLabel(bd.Contract),
		"got", got,
		"target", strconv.Itoa(bd.Target))
}

// HintOutput emits the current Troggu hint.
func (p *TrogguCuiPresenter) HintOutput(g interfaces.TrogguGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("troggu.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, trogguHintReasonKeys)
	switch {
	case hint.Bid != nil:
		return color.Yellow(i18n.Tf("troggu.hintCard",
			"cards", trogguContractLabel(domain.TrogguBid(*hint.Bid)),
			"reason", reason)) + "\n"
	case hint.CardIndex != nil:
		card := "-"
		if player := g.GetPlayer(g.GetCurrentPlayerIdx()); player != nil &&
			*hint.CardIndex >= 0 && *hint.CardIndex < player.GetCardsSize() {
			card = "[" + strconv.Itoa(*hint.CardIndex) + "]" +
				frenchTarotCuiCardStr(player.GetCard(*hint.CardIndex))
		}
		return color.Yellow(i18n.Tf("troggu.hintCard", "cards", card, "reason", reason)) + "\n"
	default:
		return color.Yellow(i18n.Tf("troggu.hintCard", "cards", "-", "reason", reason)) + "\n"
	}
}

// trogguHintReasonKeys maps hint-reason identifiers to i18n keys.
var trogguHintReasonKeys = map[string]string{
	"pass_weak_hand": "troggu.hintReasonPass",
	"bid_trois":      "troggu.hintReasonBidTrois",
	"bid_solo":       "troggu.hintReasonBidSolo",
	"bid_piccolo":    "troggu.hintReasonBidPiccolo",
	"bid_misere":     "troggu.hintReasonBidMisere",
	"play_win":       "troggu.hintReasonPlayWin",
	"play_duck":      "troggu.hintReasonPlayDuck",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *TrogguCuiPresenter) ActionLogOutput(g interfaces.TrogguGame) string {
	return actionLogOutputText(g)
}
