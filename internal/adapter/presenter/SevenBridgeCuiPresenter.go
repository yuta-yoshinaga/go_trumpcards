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

// sevenBridgePlayerStr returns the display string for a single SevenBridge player.
func sevenBridgePlayerStr(player *domain.SevenBridgePlayer, i int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("sevenbridge.playerLine",
		"name", cuiPlayerName(player, i),
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize())) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	if player.GetMeldCount() > 0 {
		b.WriteString(i18n.T("sevenbridge.meldsHeader"))
		for mi := 0; mi < player.GetMeldCount(); mi++ {
			meld := player.GetMeld(mi)
			if mi > 0 {
				b.WriteString(" | ")
			}
			cardStrs := make([]string, len(meld))
			for ci, c := range meld {
				cardStrs[ci] = cuiCardStr(c)
			}
			b.WriteString(strings.Join(cardStrs, " "))
		}
		b.WriteString("\n")
	}
	return b.String()
}

// SevenBridgeCuiPresenter renders the Seven Bridge CUI view.
type SevenBridgeCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *SevenBridgeCuiPresenter) Output(g interfaces.SevenBridgeGame, lastErr error) string {
	return buildCuiOutput(i18n.T("sevenbridge.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("sevenbridge.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"stock", strconv.Itoa(g.GetDrawPileCount())) + "\n")

		if top := g.GetDiscardTop(); top != nil {
			b.WriteString(i18n.Tf("sevenbridge.discardLine", "card", cuiCardStr(top)) + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(sevenBridgePlayerStr(g.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerIdx := g.GetWinnerIdx()
			banner := i18n.Tf("sevenbridge.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.SevenBridgePhaseDraw:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("sevenbridge.promptDraw",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("sevenbridge.promptDrawHelpStock") + "\n")
			b.WriteString(i18n.T("sevenbridge.promptDrawHelpPon") + "\n")
			b.WriteString(i18n.T("sevenbridge.promptDrawHelpChi") + "\n")
		case domain.SevenBridgePhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("sevenbridge.promptPlay",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("sevenbridge.promptPlayHelpMeld") + "\n")
			b.WriteString(i18n.T("sevenbridge.promptPlayHelpLayoff") + "\n")
			b.WriteString(i18n.T("sevenbridge.promptPlayHelpDiscard") + "\n")
		case domain.SevenBridgePhaseRoundEnd:
			b.WriteString(i18n.T("sevenbridge.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("sevenbridge.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput suggests a meld (or a discard when no meld is available) for the
// human's play phase, reusing the shared domain suggestion logic.
func (p *SevenBridgeCuiPresenter) HintOutput(g interfaces.SevenBridgeGame) string {
	// **ドローフェーズこそ一番迷う。**ポン・チーの判断を無支援にしていた (#4904)。
	if g.GetPhase() == domain.SevenBridgePhaseDraw && g.IsHumanTurn() {
		return sevenBridgeDrawHint(g)
	}
	if g.GetPhase() != domain.SevenBridgePhasePlay || !g.IsHumanTurn() {
		return i18n.T("sevenbridge.hintNotYourTurn") + "\n"
	}
	idx := g.GetCurrentPlayerIdx()
	human := g.GetPlayer(idx)
	if human == nil {
		return i18n.T("sevenbridge.hintNotYourTurn") + "\n"
	}
	if meld := g.SuggestMeld(idx); len(meld) > 0 {
		idxStrs := make([]string, len(meld))
		cards := make([]string, len(meld))
		for i, hi := range meld {
			idxStrs[i] = strconv.Itoa(hi)
			cards[i] = cuiCardStr(human.GetCard(hi))
		}
		return i18n.Tf("sevenbridge.hintMeld",
			"idxs", strings.Join(idxStrs, ", "), "cards", strings.Join(cards, ",")) + "\n"
	}
	if d := g.SuggestDiscard(idx); d >= 0 && d < human.GetCardsSize() {
		return i18n.Tf("sevenbridge.hintDiscard",
			"idx", strconv.Itoa(d), "card", cuiCardStr(human.GetCard(d))) + "\n"
	}
	return i18n.T("sevenbridge.hintNoMove") + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *SevenBridgeCuiPresenter) ActionLogOutput(g interfaces.SevenBridgeGame) string {
	return actionLogOutputText(g)
}

// sevenBridgeDrawHint はドローフェーズの助言を返す。
//
// ポンとチーの判定は claim 経路と同じ `SuggestPon` / `SuggestChi` を通す。
// 別に書くと、案内した組が拒否されることになる。
func sevenBridgeDrawHint(g interfaces.SevenBridgeGame) string {
	idx := g.GetCurrentPlayerIdx()
	human := g.GetPlayer(idx)
	if human == nil {
		return i18n.T("sevenbridge.hintNotYourTurn") + "\n"
	}
	// idxs と cards は近隣の hintMeld / hintDiscard と同じく別々に渡す。
	label := func(is []int) (string, string) {
		idxStrs := make([]string, len(is))
		cards := make([]string, len(is))
		for i, hi := range is {
			idxStrs[i] = strconv.Itoa(hi)
			cards[i] = cuiCardStr(human.GetCard(hi))
		}
		return strings.Join(idxStrs, ", "), strings.Join(cards, ",")
	}
	// **ポンが先。**同ランク 3 枚は連番より確実に面子になる。
	if pon := g.SuggestPon(idx); len(pon) > 0 {
		idxs, cards := label(pon)
		return color.Yellow(i18n.Tf("sevenbridge.hintPon", "idxs", idxs, "cards", cards)) + "\n"
	}
	if chi := g.SuggestChi(idx); len(chi) > 0 {
		idxs, cards := label(chi)
		return color.Yellow(i18n.Tf("sevenbridge.hintChi", "idxs", idxs, "cards", cards)) + "\n"
	}
	return i18n.T("sevenbridge.hintDraw") + "\n"
}
