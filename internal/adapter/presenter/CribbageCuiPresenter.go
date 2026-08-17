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

// cribbagePegValue is a card's pegging value: face cards count as 10, aces as
// 1, all others their pip value (mirrors the domain's cribbageCardValue).
func cribbagePegValue(card *domain.Card) int {
	if v := card.GetValue(); v <= 10 {
		return v
	}
	return 10
}

// cribbagePlayerStr returns the display string for a single Cribbage player.
func cribbagePlayerStr(player *domain.CribbagePlayer, i int, dealerIdx int) string {
	var b strings.Builder
	dealerMark := ""
	if i == dealerIdx {
		dealerMark = i18n.T("cribbage.dealerMark")
	}
	b.WriteString(i18n.Tf("cribbage.playerLine",
		"name", cuiPlayerName(player, i),
		"dealerMark", dealerMark,
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize())) + "\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// CribbageCuiPresenter renders the Cribbage CUI view.
type CribbageCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *CribbageCuiPresenter) Output(g interfaces.CribbageGame, lastErr error) string {
	return buildCuiOutput(i18n.T("cribbage.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("cribbage.header",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"dealer", strconv.Itoa(g.GetDealerIdx())) + "\n")

		// Starter card
		if starter := g.GetStarter(); starter != nil {
			b.WriteString(i18n.Tf("cribbage.starterLine",
				"card", cuiCardStr(starter)) + "\n")
			// **スターターが J ならディーラーに 2 点** (`cutStarter` の His Heels)。
			// Web は専用バッジで出すのに CUI は黙っており、ディーラーの点が
			// 唐突に 2 増える理由が分からなかった (#4902)。
			if starter.GetValue() == domain.CribbageJackValue {
				b.WriteString(color.Yellow(i18n.T("cribbage.hisHeels")) + "\n")
			}
		}

		// Pegging info
		phase := g.GetPhase()
		if phase == domain.CribbagePhasePegging {
			b.WriteString(i18n.Tf("cribbage.peggingTotal",
				"count", strconv.Itoa(g.GetPegCount())) + "\n")
			pegCards := g.GetPegPlayedCards()
			if len(pegCards) > 0 {
				cardStrs := make([]string, len(pegCards))
				for i, c := range pegCards {
					cardStrs[i] = cuiCardStr(c)
				}
				b.WriteString(i18n.Tf("cribbage.peggingCards",
					"cards", strings.Join(cardStrs, ", ")) + "\n")
			}
		}

		// Players
		for i := range domain.CribbagePlayerCnt {
			player := g.GetPlayer(i)
			if player != nil {
				b.WriteString(cribbagePlayerStr(player, i, g.GetDealerIdx()))
			}
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			// **最終ラウンドの内訳を見せてから終わる。** ここで早期 return して
			// いたので writeShowDetails に到達せず、`n` を連打すると内訳を一度も
			// 見ないままバナーだけが出ていた (#5512)。Web は isGameEnd でも内訳表を
			// 出している。バナーは今までどおり最後に置く。
			p.writeShowDetails(b, g)
			winnerIdx := g.GetWinnerIdx()
			banner := i18n.Tf("cribbage.gameEnd",
				"name", cuiPlayerName(g.GetPlayer(winnerIdx), winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch phase {
		case domain.CribbagePhaseDiscard:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("cribbage.promptDiscard",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("cribbage.promptDiscardHelp") + "\n")
		case domain.CribbagePhaseCut:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("cribbage.promptCut",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			b.WriteString(i18n.T("cribbage.promptCutHelp") + "\n")
		case domain.CribbagePhasePegging:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("cribbage.promptPegging",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx)) + "\n")
			// List the hand indices playable under the 31 limit so the human
			// need not add pip values by hand; warn to declare go if none fit.
			if cur := g.GetPlayer(currentIdx); cur != nil && cur.GetIsHuman() {
				pegCount := g.GetPegCount()
				var legal []string
				for i := 0; i < cur.GetCardsSize(); i++ {
					if pegCount+cribbagePegValue(cur.GetCard(i)) <= domain.CribbagePegLimit {
						legal = append(legal, "["+strconv.Itoa(i)+"]")
					}
				}
				if len(legal) > 0 {
					b.WriteString(i18n.Tf("cribbage.legalPegLegend", "indices", strings.Join(legal, " ")) + "\n")
				} else {
					b.WriteString(color.Yellow(i18n.T("cribbage.mustGo")) + "\n")
				}
			}
			b.WriteString(i18n.T("cribbage.promptPeggingHelp") + "\n")
			b.WriteString(i18n.T("cribbage.promptPeggingGo") + "\n")
		case domain.CribbagePhaseShow:
			b.WriteString(i18n.T("cribbage.promptShow") + "\n")
			p.writeShowDetails(b, g)
			b.WriteString(i18n.T("cribbage.promptShowHelp") + "\n")
		case domain.CribbagePhaseRoundEnd:
			b.WriteString(i18n.T("cribbage.promptRoundEnd") + "\n")
			p.writeShowDetails(b, g)
			b.WriteString(i18n.T("cribbage.promptRoundEndHelp") + "\n")
		}
	})
}

// writeShowDetails prints score detail lines for the show phase.
func (p *CribbageCuiPresenter) writeShowDetails(b *strings.Builder, g interfaces.CribbageGame) {
	details := g.GetHandScoreDetails()
	labelKeys := [3]string{
		"cribbage.showLabelPone",
		"cribbage.showLabelDealer",
		"cribbage.showLabelCrib",
	}
	for i, d := range details {
		if d == nil {
			continue
		}
		b.WriteString(i18n.Tf("cribbage.showDetailLine",
			"label", i18n.T(labelKeys[i]),
			"total", strconv.Itoa(d.Total),
			"fifteens", strconv.Itoa(d.Fifteens),
			"pairs", strconv.Itoa(d.Pairs),
			"runs", strconv.Itoa(d.Runs),
			"flush", strconv.Itoa(d.Flush),
			"nobs", strconv.Itoa(d.Nobs)) + "\n")
	}
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *CribbageCuiPresenter) ActionLogOutput(g interfaces.CribbageGame) string {
	return actionLogOutputText(g)
}

// HintOutput ヒントを出力（ディスカード推奨2枚 or ペギング推奨1枚）
func (p *CribbageCuiPresenter) HintOutput(g interfaces.CribbageGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	switch hint.Type {
	case "discard":
		if len(hint.Indices) < 2 {
			return i18n.T("cuiHintNone") + "\n"
		}
		return i18n.Tf("cribbage.hintDiscard",
			"i", strconv.Itoa(hint.Indices[0]),
			"j", strconv.Itoa(hint.Indices[1])) + "\n"
	case "play":
		if len(hint.Indices) < 1 {
			return i18n.T("cuiHintNone") + "\n"
		}
		return i18n.Tf("cribbage.hintPlay", "i", strconv.Itoa(hint.Indices[0])) + "\n"
	default:
		return i18n.T("cuiHintNone") + "\n"
	}
}
