//go:build !js || !wasm || casino

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// musPhaseLabel returns the i18n display label for a mus betting phase index.
func musPhaseLabel(ri int) string {
	switch ri {
	case 0:
		return i18n.T("mus.roundGrande")
	case 1:
		return i18n.T("mus.roundChica")
	case 2:
		return i18n.T("mus.roundPares")
	case 3:
		return i18n.T("mus.roundJuego")
	default:
		return "?"
	}
}

// musActionLabel returns the i18n display label for a mus bet action int.
func musActionLabel(action int) string {
	switch action {
	case domain.MusActionPaso:
		return i18n.T("mus.actionPaso")
	case domain.MusActionEnvido:
		return i18n.T("mus.actionEnvido")
	case domain.MusActionOrdago:
		return i18n.T("mus.actionOrdago")
	case domain.MusActionQuiero:
		return i18n.T("mus.actionQuiero")
	case domain.MusActionNoQuiero:
		return i18n.T("mus.actionNoQuiero")
	default:
		return "?"
	}
}

// musResultKindLabel returns the i18n display label for a mus result kind int,
// so the round-result line shows a name rather than a raw MusResult* constant.
func musResultKindLabel(kind int) string {
	switch kind {
	case domain.MusResultDeferred:
		return i18n.T("mus.resultDeferred")
	case domain.MusResultAccepted:
		return i18n.T("mus.resultAccepted")
	case domain.MusResultAwarded:
		return i18n.T("mus.resultAwarded")
	case domain.MusResultOrdago:
		return i18n.T("mus.resultOrdago")
	default:
		return "?"
	}
}

// musParesLabel は Pares 分類のロケール文字列を返す。
func musParesLabel(category int) string {
	switch category {
	case domain.MusParesDuples:
		return i18n.T("mus.paresDuples")
	case domain.MusParesMedias:
		return i18n.T("mus.paresMedias")
	case domain.MusParesPar:
		return i18n.T("mus.paresPar")
	default:
		return i18n.T("mus.paresNone")
	}
}

// musJuegoLabel は Juego のロケール文字列を返す。
//
// ちょうど 31 は Juego の中で最強なので別扱いにする -- Web の juegoBest と同じ。
func musJuegoLabel(s *domain.MusHandSummary) string {
	switch {
	case s.Points == domain.MusJuegoThreshold:
		return i18n.T("mus.juegoBest")
	case s.HasJuego:
		return i18n.Tf("mus.juegoYes", "points", strconv.Itoa(s.Points))
	default:
		return i18n.Tf("mus.juegoPunto", "points", strconv.Itoa(s.Points))
	}
}

// musPlayerStr returns the display string for a single Mus player.
func musPlayerStr(g interfaces.MusGame, i int) string {
	player := g.GetPlayer(i)
	if player == nil {
		return ""
	}
	team := domain.MusTeamOf(i)
	amarrakos := g.GetAmarrakos()
	var b strings.Builder
	b.WriteString(i18n.Tf("mus.playerLine",
		"name", cuiPlayerName(player, i),
		"team", strconv.Itoa(team),
		"score", strconv.Itoa(amarrakos[team]),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	// **4 つの賭けラウンドはそれぞれ手札の別の側面を見ている** (#5640)。Web は
	// mus-hand-summary で 4 項目を常時出しているのに、CUI は札を並べるだけで、
	// Mus 独自のランク付け (A/K が高位) を暗算させていた。評価はドメインが持つ。
	if player.GetIsHuman() {
		if s := g.GetHandSummary(i); s != nil {
			b.WriteString(i18n.Tf("mus.summaryLine",
				"label", i18n.T("mus.summaryLabel"),
				"grande", strconv.Itoa(s.HighestRank),
				"chica", strconv.Itoa(s.LowestRank),
				"pares", musParesLabel(s.ParesCategory),
				"juego", musJuegoLabel(s)) + "\n")
		}
	}
	return b.String()
}

// MusCuiPresenter renders the Mus CUI view.
type MusCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *MusCuiPresenter) Output(g interfaces.MusGame, lastErr error) string {
	return buildCuiOutput(i18n.T("mus.helpTitle"), func(b *strings.Builder) {
		amarrakos := g.GetAmarrakos()
		b.WriteString(i18n.Tf("mus.roundLine",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"t0", strconv.Itoa(amarrakos[0]),
			"t1", strconv.Itoa(amarrakos[1]),
		) + "\n")

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(musPlayerStr(g, i))
		}

		b.WriteString("----------\n")

		// Show betting round results if in a betting or later phase.
		phase := g.GetPhase()
		if phase >= domain.MusPhaseGrande {
			for ri := 0; ri < domain.MusRoundCnt; ri++ {
				r := g.GetResult(ri)
				if r.Kind != domain.MusResultPending {
					b.WriteString(i18n.Tf("mus.resultLine",
						"round", musPhaseLabel(ri),
						"kind", musResultKindLabel(r.Kind),
						"stake", strconv.Itoa(r.Stake),
					) + "\n")
				}
			}
		}

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winnerTeam := g.GetWinnerTeam()
			banner := i18n.Tf("mus.gameEnd", "team", strconv.Itoa(winnerTeam))
			b.WriteString(color.Green(banner) + "\n")
			return
		}

		switch phase {
		case domain.MusPhaseMus:
			b.WriteString(i18n.T("mus.promptMus") + "\n")
			b.WriteString(i18n.T("mus.promptMusHelp") + "\n")
		case domain.MusPhaseDiscard:
			b.WriteString(i18n.T("mus.promptDiscard") + "\n")
			b.WriteString(i18n.T("mus.promptDiscardHelp") + "\n")
		case domain.MusPhaseGrande, domain.MusPhaseChica, domain.MusPhasePares, domain.MusPhaseJuego:
			b.WriteString(i18n.Tf("mus.promptBet",
				"round", musPhaseLabel(int(phase)-2),
				"pending", strconv.Itoa(g.GetPendingStake()),
			) + "\n")
			b.WriteString(i18n.T("mus.promptBetHelp") + "\n")
		case domain.MusPhaseShowdown:
			b.WriteString(i18n.T("mus.promptShowdown") + "\n")
		case domain.MusPhaseRoundEnd:
			b.WriteString(i18n.T("mus.promptRoundEnd") + "\n")
			b.WriteString(i18n.T("mus.promptRoundEndHelp") + "\n")
		}
	})
}

// HintOutput emits the current Mus hint.
func (p *MusCuiPresenter) HintOutput(g interfaces.MusGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("mus.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, musHintReasonKeys)
	phase := g.GetPhase()
	switch phase {
	case domain.MusPhaseMus:
		action := "cut"
		if hint.Mus {
			action = "mus"
		}
		return color.Yellow(i18n.Tf("mus.hintMus", "action", action, "reason", reason)) + "\n"
	case domain.MusPhaseDiscard:
		if len(hint.Indices) == 0 {
			return color.Yellow(i18n.Tf("mus.hintDiscardNone", "reason", reason)) + "\n"
		}
		human := g.GetPlayer(g.GetDiscardTurn())
		cards := make([]string, len(hint.Indices))
		for j, idx := range hint.Indices {
			if human != nil {
				cards[j] = "[" + strconv.Itoa(idx) + "]" + cuiCardStr(human.GetCard(idx))
			} else {
				cards[j] = strconv.Itoa(idx)
			}
		}
		return color.Yellow(i18n.Tf("mus.hintDiscard",
			"cards", strings.Join(cards, ", "),
			"reason", reason)) + "\n"
	default:
		return color.Yellow(i18n.Tf("mus.hintBet",
			"action", musActionLabel(hint.Action),
			"reason", reason)) + "\n"
	}
}

// musHintReasonKeys maps Mus-specific hint-reason identifiers to i18n keys.
var musHintReasonKeys = map[string]string{
	"mus_exchange":  "mus.hintReasonMusExchange",
	"mus_cut":       "mus.hintReasonMusCut",
	"discard_low":   "mus.hintReasonDiscardLow",
	"bet_paso":      "mus.hintReasonBetPaso",
	"bet_envido":    "mus.hintReasonBetEnvido",
	"bet_ordago":    "mus.hintReasonBetOrdago",
	"bet_quiero":    "mus.hintReasonBetQuiero",
	"bet_no_quiero": "mus.hintReasonBetNoQuiero",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *MusCuiPresenter) ActionLogOutput(g interfaces.MusGame) string {
	return actionLogOutputTextForSeats[*domain.MusPlayer](g)
}
