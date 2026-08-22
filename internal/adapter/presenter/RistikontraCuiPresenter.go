//go:build !js || !wasm || extra2

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// RistikontraCuiPresenter は Pişti CUI ビューを描画する。
type RistikontraCuiPresenter struct{}

// Output は現在のゲーム状態をアクティブロケールで描画する。
func (p *RistikontraCuiPresenter) Output(pg interfaces.RistikontraGame, lastErr error) string {
	return buildCuiOutput(i18n.T("ristikontra.helpTitle"), func(b *strings.Builder) {
		for i := 0; i < pg.GetPlayerCnt(); i++ {
			b.WriteString(ristikontraPlayerStr(pg, pg.GetPlayer(i), i))
		}
		// **対局中の優劣を数値で出す。**リスティコントラ賞と捕獲枚数を別々に出すだけ
		// では、複数プレイヤー分を毎回暗算することになる (#4892)。ゲーム終了後は
		// 下の最終スコアが出るので、ここでは出さない。
		if !pg.GetGameEndFlag() {
			prov := pg.GetProvisionalScores()
			leader := pg.GetProvisionalLeader()
			for i := 0; i < pg.GetPlayerCnt() && i < len(prov); i++ {
				pl := pg.GetPlayer(i)
				if pl == nil {
					continue
				}
				line := i18n.Tf("ristikontra.provisional",
					"name", cuiPlayerName(pl, i), "score", strconv.Itoa(prov[i]))
				if i == leader {
					line = color.Yellow(line)
				}
				b.WriteString(line + "\n")
			}
			// **カード点は含まないと断る。**確実な分だけの近似値なので。
			b.WriteString(i18n.T("ristikontra.provisionalNote") + "\n")
		}
		b.WriteString("----------\n")

		if top := pg.GetPileTop(); top != nil {
			b.WriteString(i18n.Tf("ristikontra.pileLine",
				"top", cuiCardSliceStr([]*domain.Card{top}),
				"count", strconv.Itoa(len(pg.GetPile()))) + "\n")
		} else {
			b.WriteString(i18n.T("ristikontra.pileEmpty") + "\n")
		}
		b.WriteString(i18n.Tf("ristikontra.deckLine",
			"count", strconv.Itoa(pg.GetRemainingDeck())) + "\n")

		cuiErrorBlock(b, lastErr)

		if pg.GetGameEndFlag() {
			b.WriteString(i18n.T("ristikontra.gameEnd") + "\n")
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
				b.WriteString(i18n.Tf("ristikontra.scoreEntry",
					"name", cuiPlayerName(pl, i),
					"score", strconv.Itoa(score)) + "\n")
			}
			return
		}
		currentTurn := pg.GetCurrentTurn()
		b.WriteString(i18n.Tf("ristikontra.promptCurrentTurn",
			"name", cuiPlayerName(pg.GetPlayer(currentTurn), currentTurn)) + "\n")
		b.WriteString(i18n.T("ristikontra.promptHelp") + "\n")
	})
}

// ristikontraPlayerStr は 1 プレイヤーの表示文字列を返す。
func ristikontraPlayerStr(pg interfaces.RistikontraGame, player *domain.RistikontraPlayer, i int) string {
	if player == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("ristikontra.playerLine",
		"name", cuiPlayerName(player, i),
		"hand", strconv.Itoa(player.GetCardsSize()),
		"captured", strconv.Itoa(player.CapturedCount()),
		"team", strconv.Itoa(i%2+1)) + "\n")
	if player.GetIsHuman() {
		// **場を取れるのは「場のトップと同ランク」か「ジャック(総取り)」** (#5672)。
		// Web は該当札にリングを付けているのに、CUI は素の一覧で、毎ターン場の
		// トップと手札を照合させていた。自分の手番でないときは出さない。
		var capturing []int
		if pg.IsHumanTurn() {
			top := pg.GetPileTop()
			for idx := 0; idx < player.GetCardsSize(); idx++ {
				c := player.GetCard(idx)
				if c == nil {
					continue
				}
				if c.GetValue() == domain.RistikontraJackValue || (top != nil && c.GetValue() == top.GetValue()) {
					capturing = append(capturing, idx)
				}
			}
		}
		b.WriteString(cuiIndexMarkedCardListStr(player, capturing, CuiLegalMark) + "\n")
		if len(capturing) > 0 {
			b.WriteString(i18n.T("ristikontra.captureLegend") + "\n")
		}
	}
	return b.String()
}

// ActionLogOutput は棋譜をテキストとして出力する。
func (p *RistikontraCuiPresenter) ActionLogOutput(pg interfaces.RistikontraGame) string {
	return actionLogOutputTextForSeats[*domain.RistikontraPlayer](pg)
}
