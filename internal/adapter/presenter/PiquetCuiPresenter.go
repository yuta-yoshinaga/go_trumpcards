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

// PiquetCuiPresenter Piquet CUIプレゼンター
type PiquetCuiPresenter struct{}

// Output 現在のゲーム状態を描画する
func (p *PiquetCuiPresenter) Output(g interfaces.PiquetGame, lastErr error) string {
	return buildCuiOutput(i18n.T("piquet.helpTitle"), func(b *strings.Builder) {
		// ヘッダー
		b.WriteString(i18n.Tf("piquet.dealHeader",
			"deal", strconv.Itoa(g.GetDealNumber()),
			"total", strconv.Itoa(g.GetDealsPerPartie())) + "\n")

		// プレイヤー (Elder / Younger)
		piquetWritePlayerSection(b, g, g.GetElderIdx(), i18n.T("piquet.roleElder"))
		piquetWritePlayerSection(b, g, g.GetYoungerIdx(), i18n.T("piquet.roleYounger"))
		b.WriteString("----------\n")

		switch g.GetPhase() {
		case domain.PiquetPhaseExchange:
			piquetWriteExchangeView(b, g)
		case domain.PiquetPhaseDeclaration:
			piquetWriteDeclarationView(b, g)
		case domain.PiquetPhasePlay:
			piquetWritePlayView(b, g)
		case domain.PiquetPhaseScore:
			piquetWriteRoundEndView(b, g)
		case domain.PiquetPhaseGameEnd:
			piquetWriteGameEndView(b, g)
		}

		cuiErrorBlock(b, lastErr)
	})
}

func piquetWritePlayerSection(b *strings.Builder, g interfaces.PiquetGame, idx int, role string) {
	pl := g.GetPlayer(idx)
	if pl == nil {
		return
	}
	tag := role
	if pl.GetIsHuman() {
		tag += " (" + i18n.T("piquet.you") + ")"
	} else {
		tag += " (CPU)"
	}
	b.WriteString(i18n.Tf("piquet.playerLine",
		"role", tag,
		"hand", strconv.Itoa(pl.GetCardsSize()),
		"tricks", strconv.Itoa(g.GetTricksWon(idx)),
		"round", strconv.Itoa(pl.GetRoundScore()),
		"match", strconv.Itoa(pl.GetMatchScore())) + "\n")
	if pl.GetIsHuman() {
		b.WriteString("  ")
		for i := 0; i < pl.GetCardsSize(); i++ {
			if i > 0 {
				b.WriteString(" ")
			}
			b.WriteString("[" + strconv.Itoa(i) + "]" + cuiCardStr(pl.GetCard(i)))
		}
		b.WriteString("\n")
	}
	if g.GetCarteBlanche(idx) {
		b.WriteString("  " + color.Green(i18n.T("piquet.carteBlanche")) + "\n")
	}
}

func piquetWriteExchangeView(b *strings.Builder, g interfaces.PiquetGame) {
	b.WriteString(i18n.T("piquet.exchangeHeader") + "\n")
	switch g.GetExchangeTurn() {
	case domain.PiquetExchangeTurnElder:
		b.WriteString(i18n.T("piquet.elderToExchange") + "\n")
	case domain.PiquetExchangeTurnYounger:
		b.WriteString(i18n.T("piquet.youngerToExchange") + "\n")
	}
}

func piquetWriteDeclarationView(b *strings.Builder, g interfaces.PiquetGame) {
	b.WriteString(i18n.T("piquet.declarationHeader") + "\n")
	for _, r := range g.GetDeclResults() {
		b.WriteString("  - " + piquetFormatDeclResult(r) + "\n")
	}
	b.WriteString("  " + i18n.Tf("piquet.nextDeclStage",
		"kind", piquetKindLabel(g.GetDeclStage())) + "\n")
}

func piquetWritePlayView(b *strings.Builder, g interfaces.PiquetGame) {
	b.WriteString(i18n.T("piquet.playHeader") + "\n")
	if len(g.GetCurrentTrick()) > 0 {
		b.WriteString("  ")
		for _, tc := range g.GetCurrentTrick() {
			b.WriteString(strconv.Itoa(tc.PlayerIdx) + ":" + cuiCardStr(tc.Card) + "  ")
		}
		b.WriteString("\n")
	}
	b.WriteString("  " + i18n.Tf("piquet.trickNumber",
		"trick", strconv.Itoa(g.GetTrickNumber()+1)) + "\n")
}

func piquetWriteRoundEndView(b *strings.Builder, g interfaces.PiquetGame) {
	b.WriteString(color.Green(i18n.T("piquet.roundEnd")) + "\n")
	for _, r := range g.GetDeclResults() {
		b.WriteString("  - " + piquetFormatDeclResult(r) + "\n")
	}
}

func piquetWriteGameEndView(b *strings.Builder, g interfaces.PiquetGame) {
	winner := g.GetWinnerIdx()
	if winner == -1 {
		b.WriteString(color.Yellow(i18n.T("piquet.partieDraw")) + "\n")
	} else {
		b.WriteString(color.Green(i18n.Tf("piquet.partieWinner",
			"idx", strconv.Itoa(winner))) + "\n")
	}
}

func piquetFormatDeclResult(r *domain.PiquetDeclarationResult) string {
	if r == nil {
		return ""
	}
	kind := piquetKindLabel(r.Kind)
	if r.Winner < 0 || r.Score == 0 {
		return i18n.Tf("piquet.declTied", "kind", kind)
	}
	return i18n.Tf("piquet.declWinner",
		"kind", kind,
		"idx", strconv.Itoa(r.ScoredBy),
		"score", strconv.Itoa(r.Score))
}

func piquetKindLabel(k domain.PiquetDeclarationKind) string {
	switch k {
	case domain.PiquetDeclKindPoint:
		return i18n.T("piquet.declKindPoint")
	case domain.PiquetDeclKindSequence:
		return i18n.T("piquet.declKindSequence")
	case domain.PiquetDeclKindSet:
		return i18n.T("piquet.declKindSet")
	}
	return ""
}

// HintOutput ヒントを出力
func (p *PiquetCuiPresenter) HintOutput(g interfaces.PiquetGame) string {
	hint := g.GetHint(g.GetCurrentPlayerIdx())
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	if hint.CardIndex != nil {
		return i18n.Tf("piquet.hintPlay", "idx", strconv.Itoa(*hint.CardIndex)) + "\n"
	}
	if len(hint.DiscardIndices) > 0 {
		out := make([]string, len(hint.DiscardIndices))
		for i, idx := range hint.DiscardIndices {
			out[i] = strconv.Itoa(idx)
		}
		return i18n.Tf("piquet.hintDiscard", "idxs", strings.Join(out, ",")) + "\n"
	}
	return i18n.T("cuiHintNone") + "\n"
}

// ActionLogOutput アクションログを出力
func (p *PiquetCuiPresenter) ActionLogOutput(g interfaces.PiquetGame) string {
	return actionLogToText(g.GetActionLog())
}
