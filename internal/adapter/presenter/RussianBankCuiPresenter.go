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

// rbCardOrDash カードを表示文字列にする。nil は "-"。
func rbCardOrDash(c *domain.Card) string {
	if c == nil {
		return "-"
	}
	return cuiCardStr(c)
}

// rbTableauPiles はタブロー列を「下から順に並べ、トップだけ [] で囲む」形で表示する。
//
// **列は複数枚重なるのに、CUI はトップ 1 枚しか出していなかった** (#5677)。Web は
// #3574 で埋もれた札のランク・スートをカスケード表示している。
//
// 1 枚だけの列と空列は rbPileTops と同じ見た目のままにする (`[card]` / `[-]`) ──
// トップがどれかは常に [] で読める。
func rbTableauPiles(piles [][]*domain.Card) string {
	parts := make([]string, len(piles))
	for i, pile := range piles {
		if len(pile) == 0 {
			parts[i] = "[-]"
			continue
		}
		var sb strings.Builder
		for _, c := range pile[:len(pile)-1] {
			sb.WriteString(cuiCardStr(c) + " ")
		}
		sb.WriteString("[" + cuiCardStr(pile[len(pile)-1]) + "]")
		parts[i] = sb.String()
	}
	return strings.Join(parts, "  ")
}

// rbPileTops カード列をトップのみ簡潔表示する (ファウンデーション用)。
//
// ファウンデーションは A から順に積むので、下に何があるかはトップから決まる。
// タブローと違って埋もれた札を出す必要がない (#5677)。
func rbPileTops(piles [][]*domain.Card) string {
	parts := make([]string, len(piles))
	for i, pile := range piles {
		if len(pile) == 0 {
			parts[i] = "[-]"
		} else {
			parts[i] = "[" + cuiCardStr(pile[len(pile)-1]) + "]"
		}
	}
	return strings.Join(parts, " ")
}

// RussianBankCuiPresenter ロシアンバンク (クラペット) のCUIプレゼンタークラス。
type RussianBankCuiPresenter struct{}

// Output 現在のゲーム状態を描画する。
func (p *RussianBankCuiPresenter) Output(g interfaces.RussianBankGame, lastErr error) string {
	return buildCuiOutput(i18n.T("russianbank.helpTitle"), func(b *strings.Builder) {
		foundations := g.GetFoundations()
		b.WriteString(i18n.Tf("russianbank.foundations", "cards", rbPileTops(foundations[:])) + "\n")
		tableau := g.GetTableau()
		b.WriteString(i18n.Tf("russianbank.tableau", "cards", rbTableauPiles(tableau[:])) + "\n")

		for i, player := range g.GetPlayers() {
			if player == nil {
				continue
			}
			b.WriteString(i18n.Tf("russianbank.playerLine",
				"name", cuiPlayerName(player, i),
				"reserveTop", rbCardOrDash(player.ReserveTop()),
				"reserve", strconv.Itoa(player.ReserveSize()),
				"hand", strconv.Itoa(player.HandSize()),
				"wasteTop", rbCardOrDash(player.WasteTop()),
				"waste", strconv.Itoa(player.WasteSize()),
				"stop", strconv.Itoa(g.GetStopPoints(i))) + "\n")
		}

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			b.WriteString(color.Green(p.gameEndBanner(g)) + "\n")
			return
		}
		p.writePrompt(b, g)
	})
}

// gameEndBanner 結果バナーを描画する。
func (p *RussianBankCuiPresenter) gameEndBanner(g interfaces.RussianBankGame) string {
	winner := g.GetWinner()
	if winner < 0 {
		return i18n.T("russianbank.gameEndDraw")
	}
	if player := g.GetPlayer(winner); player != nil {
		return i18n.Tf("russianbank.gameEnd", "name", cuiPlayerName(player, winner))
	}
	return i18n.T("russianbank.gameEndDraw")
}

// writePrompt 手番に応じたプロンプトを描画する。
func (p *RussianBankCuiPresenter) writePrompt(b *strings.Builder, g interfaces.RussianBankGame) {
	if !g.IsHumanTurn() {
		b.WriteString(i18n.T("russianbank.cpuTurn") + "\n")
		return
	}
	if g.CanCallStop() {
		b.WriteString(color.Yellow(i18n.T("russianbank.stopAvailable")) + "\n")
	}
	b.WriteString(i18n.T("russianbank.promptHuman") + "\n")
	b.WriteString(i18n.T("russianbank.promptHelp") + "\n")
}

// HintOutput ヒントを描画する。
func (p *RussianBankCuiPresenter) HintOutput(g interfaces.RussianBankGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("russianbank.hintNone") + "\n"
	}
	dest := i18n.T("russianbank.destFoundation")
	if !hint.ToFoundation {
		dest = i18n.Tf("russianbank.destTableau", "col", strconv.Itoa(hint.ToCol))
	}
	return color.Yellow(i18n.Tf("russianbank.hintMove",
		"src", rbHintSourceName(hint),
		"dest", dest)) + "\n"
}

// rbHintSourceName ヒント移動元の表示名を返す。
func rbHintSourceName(h *domain.RussianBankHint) string {
	switch h.Zone {
	case domain.RussianBankZoneReserve:
		if h.FromOpponent {
			return i18n.T("russianbank.srcOppReserve")
		}
		return i18n.T("russianbank.srcReserve")
	case domain.RussianBankZoneWaste:
		if h.FromOpponent {
			return i18n.T("russianbank.srcOppWaste")
		}
		return i18n.T("russianbank.srcWaste")
	default:
		return i18n.Tf("russianbank.srcTableau", "col", strconv.Itoa(h.Col))
	}
}

// ActionLogOutput 棋譜をテキスト出力する。
func (p *RussianBankCuiPresenter) ActionLogOutput(g interfaces.RussianBankGame) string {
	return actionLogOutputTextForSeats[*domain.RussianBankPlayer](g)
}
