//go:build !js || !wasm || extra5

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// tongitsPlayerStr returns the display string for a single Tongits player.
func tongitsPlayerStr(player *domain.TongitsPlayer, i int) string {
	var b strings.Builder
	b.WriteString(i18n.Tf("tongits.playerLine",
		"name", cuiPlayerName(player, i),
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize())) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// tongitsMeldIsSet reports whether a meld is a set (same rank) rather than a run.
func tongitsMeldIsSet(meld []*domain.Card) bool {
	if len(meld) == 0 {
		return false
	}
	rank := meld[0].GetValue()
	for _, c := range meld {
		if c.GetValue() != rank {
			return false
		}
	}
	return true
}

// writeTongitsPlayerMelds は 1 人分の公開メルドを書き出す。
//
// Tongits のメルドは**出した時点で全員に見える**ので、宣言した 1 人だけを特別扱いする
// クローン元 (Tonk) の描画は成り立たない。誰のメルドかを添えて、全員分を同じ形で出す。
func writeTongitsPlayerMelds(b *strings.Builder, name string, melds [][]*domain.Card) {
	if len(melds) == 0 {
		return
	}
	b.WriteString(color.Bold(i18n.Tf("tongits.playerMeldsHeader", "name", name)) + "\n")
	for i, meld := range melds {
		typeLabel := i18n.T("tongits.meldRun")
		if tongitsMeldIsSet(meld) {
			typeLabel = i18n.T("tongits.meldSet")
		}
		b.WriteString(i18n.Tf("tongits.playerMeldLine",
			"idx", strconv.Itoa(i+1),
			"type", typeLabel,
			"cards", cuiCardSliceStr(meld)) + "\n")
	}
}

func tongitsHandCards(player *domain.TongitsPlayer) []*domain.Card {
	cards := make([]*domain.Card, player.GetCardsSize())
	for i := range cards {
		cards[i] = player.GetCard(i)
	}
	return cards
}

// TongitsCuiPresenter renders the Tongits CUI view.
type TongitsCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *TongitsCuiPresenter) Output(g interfaces.TongitsGame, lastErr error) string {
	return buildCuiOutput(i18n.T("tongits.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("tongits.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"stock", strconv.Itoa(g.GetDrawPileCount())) + "\n")

		if top := g.GetDiscardTop(); top != nil {
			b.WriteString(i18n.Tf("tongits.discardLine", "card", cuiCardStr(top)) + "\n")
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(tongitsPlayerStr(g.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerIdx := g.GetWinnerIdx()
			banner := i18n.Tf("tongits.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.TongitsPhaseDraw:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("tongits.promptDraw",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("tongits.promptDrawHelpStock") + "\n")
			b.WriteString(i18n.T("tongits.promptDrawHelpDiscard") + "\n")
		case domain.TongitsPhaseDiscard:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("tongits.promptDiscard",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			// 残り点は challenge の勝敗そのものなので、自分の手番のときだけ出す。
			// CPU の手番に出しても人間は行動できない。
			if cur := g.GetPlayer(currentIdx); cur.GetIsHuman() {
				b.WriteString(i18n.Tf("tongits.currentRemainingPoints",
					"value", strconv.Itoa(tongitsHandPoints(cur))) + "\n")
			}
			b.WriteString(i18n.T("tongits.promptMeldHelp") + "\n")
			b.WriteString(i18n.T("tongits.promptSapawHelp") + "\n")
			b.WriteString(i18n.T("tongits.promptDiscardHelp") + "\n")
			b.WriteString(i18n.T("tongits.promptChallengeHelp") + "\n")
		case domain.TongitsPhaseRoundEnd:
			if g.GetIsTongits() {
				b.WriteString(i18n.T("tongits.promptDealtTongits") + "\n")
			}
			// 決着の根拠を見せる。Tongits では melds は最初から場に公開されて
			// いるので、ノッカーだけを特別扱いする理由が無い ── 全員のメルドと、
			// challenge の判定材料である残り点を並べる。
			for i := 0; i < g.GetPlayerCnt(); i++ {
				cp := g.GetPlayer(i)
				writeTongitsPlayerMelds(b, cuiPlayerName(cp, i), cp.GetMelds())
				if !cp.GetIsHuman() && cp.GetCardsSize() > 0 {
					b.WriteString(i18n.Tf("tongits.revealedHand",
						"name", cuiPlayerName(cp, i),
						"cards", cuiCardSliceStr(tongitsHandCards(cp))) + "\n")
				}
				b.WriteString(i18n.Tf("tongits.revealedPoints",
					"name", cuiPlayerName(cp, i),
					"value", strconv.Itoa(tongitsHandPoints(cp))) + "\n")
			}
			b.WriteString(i18n.T("tongits.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("tongits.promptRoundEndHelp") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *TongitsCuiPresenter) ActionLogOutput(g interfaces.TongitsGame) string {
	return actionLogOutputTextForSeats[*domain.TongitsPlayer](g)
}

// tongitsHandPoints は手札の残り点を返す。challenge の勝敗はこの少なさで決まる。
func tongitsHandPoints(p *domain.TongitsPlayer) int {
	total := 0
	for i := 0; i < p.GetCardsSize(); i++ {
		total += domain.TongitsCardValue(p.GetCard(i))
	}
	return total
}
