//go:build !js || !wasm || extra2

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// PishtiCuiPresenter は Pişti CUI ビューを描画する。
type PishtiCuiPresenter struct{}

// Output は現在のゲーム状態をアクティブロケールで描画する。
func (p *PishtiCuiPresenter) Output(pg interfaces.PishtiGame, lastErr error) string {
	return buildCuiOutput(i18n.T("pishti.helpTitle"), func(b *strings.Builder) {
		for i := 0; i < pg.GetPlayerCnt(); i++ {
			b.WriteString(pishtiPlayerStr(pg.GetPlayer(i), i))
		}
		b.WriteString("----------\n")

		if top := pg.GetPileTop(); top != nil {
			b.WriteString(i18n.Tf("pishti.pileLine",
				"top", cuiCardSliceStr([]*domain.Card{top}),
				"count", strconv.Itoa(len(pg.GetPile()))) + "\n")
		} else {
			b.WriteString(i18n.T("pishti.pileEmpty") + "\n")
		}
		b.WriteString(i18n.Tf("pishti.deckLine",
			"count", strconv.Itoa(pg.GetRemainingDeck())) + "\n")

		cuiErrorBlock(b, lastErr)

		if pg.GetGameEndFlag() {
			b.WriteString(i18n.T("pishti.gameEnd") + "\n")
			scores := pg.GetFinalScores()
			for i := 0; i < pg.GetPlayerCnt(); i++ {
				pl := pg.GetPlayer(i)
				if pl == nil {
					continue
				}
				score := 0
				if i < len(scores) {
					score = scores[i]
				}
				b.WriteString(i18n.Tf("pishti.scoreEntry",
					"name", cuiPlayerName(pl, i),
					"score", strconv.Itoa(score)) + "\n")
			}
			return
		}
		currentTurn := pg.GetCurrentTurn()
		b.WriteString(i18n.Tf("pishti.promptCurrentTurn",
			"name", cuiPlayerName(pg.GetPlayer(currentTurn), currentTurn)) + "\n")
		b.WriteString(i18n.T("pishti.promptHelp") + "\n")
	})
}

// pishtiPlayerStr は 1 プレイヤーの表示文字列を返す。
func pishtiPlayerStr(player *domain.PishtiPlayer, i int) string {
	if player == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("pishti.playerLine",
		"name", cuiPlayerName(player, i),
		"hand", strconv.Itoa(player.GetCardsSize()),
		"captured", strconv.Itoa(player.CapturedCount()),
		"pisti", strconv.Itoa(player.GetPistiBonus())) + "\n")
	if player.GetIsHuman() {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// ActionLogOutput は棋譜をテキストとして出力する。
func (p *PishtiCuiPresenter) ActionLogOutput(pg interfaces.PishtiGame) string {
	return actionLogOutputText(pg)
}
