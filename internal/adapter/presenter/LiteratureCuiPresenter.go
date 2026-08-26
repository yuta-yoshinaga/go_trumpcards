//go:build !js || !wasm || extra4

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// literatureHalfSuitLabel はハーフスートの表示名を返す。
func literatureHalfSuitLabel(half int) string {
	cards := domain.LiteratureHalfSuitCards(half)
	if len(cards) == 0 {
		return "-"
	}
	suit := "-"
	switch cards[0].GetDesign() {
	case domain.CardDesignSpade:
		suit = "♠"
	case domain.CardDesignClover:
		suit = "♣"
	case domain.CardDesignHeart:
		suit = "♥"
	case domain.CardDesignDiamond:
		suit = "♦"
	}
	key := "literature.lowHalf"
	if half%2 == 1 {
		key = "literature.highHalf"
	}
	return suit + i18n.T(key)
}

// literatureStateLabel はハーフスートの帰属の表示名を返す。
func literatureStateLabel(st domain.LiteratureHalfSuitState) string {
	switch st {
	case domain.LiteratureHalfTeam0:
		return i18n.Tf("literature.ownedBy", "team", "0")
	case domain.LiteratureHalfTeam1:
		return i18n.Tf("literature.ownedBy", "team", "1")
	case domain.LiteratureHalfCancelled:
		// **無効。**どちらのものにもならない。
		return i18n.T("literature.cancelled")
	}
	return i18n.T("literature.open")
}

// literaturePlayerStr returns the display string for a single seat.
func literaturePlayerStr(g interfaces.LiteratureGame, i int) string {
	player := g.GetPlayer(i)
	if player == nil {
		return ""
	}
	// **終局まで誰の手札も見えない。**味方の手札が見えたら推理が成立しない。
	hand := i18n.Tf("literature.hiddenHand", "count", strconv.Itoa(player.GetCardsSize()))
	if player.GetIsHuman() || g.GetGameEndFlag() {
		var b strings.Builder
		for j := range player.GetCardsSize() {
			b.WriteString(cuiCardStr(player.GetCard(j)) + " ")
		}
		hand = strings.TrimSpace(b.String())
	}
	turn := ""
	if !g.GetGameEndFlag() && i == g.GetCurrentPlayerIdx() {
		turn = " " + i18n.T("literature.turnTag")
	}
	return i18n.Tf("literature.playerLine",
		"seat", strconv.Itoa(i),
		"name", cuiPlayerName(player, i),
		"team", strconv.Itoa(domain.LiteratureTeamOf(i)),
		"turn", turn,
		"count", strconv.Itoa(player.GetCardsSize()),
		"hand", hand) + "\n"
}

// LiteratureCuiPresenter renders the Literature CUI view.
type LiteratureCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *LiteratureCuiPresenter) Output(g interfaces.LiteratureGame, lastErr error) string {
	return buildCuiOutput(i18n.T("literature.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("literature.header",
			"t0", strconv.Itoa(g.LiteratureTeamHalfSuits(0)),
			"t1", strconv.Itoa(g.LiteratureTeamHalfSuits(1)),
			"cancelled", strconv.Itoa(g.LiteratureCancelledCount()),
			"open", strconv.Itoa(g.LiteratureOpenCount()),
			"need", strconv.Itoa(domain.LiteratureWinThreshold)) + "\n")
		b.WriteString(i18n.T("literature.thresholdNote") + "\n")

		// ハーフスートの一覧と帰属。
		for half := range domain.LiteratureHalfSuitCnt {
			b.WriteString(i18n.Tf("literature.halfSuitLine",
				"idx", strconv.Itoa(half),
				"name", literatureHalfSuitLabel(half),
				"state", literatureStateLabel(g.GetHalfSuitState(half))) + "\n")
		}

		for i := range g.GetPlayers() {
			b.WriteString(literaturePlayerStr(g, i))
		}

		// **要求の履歴は公開情報。**直近だけ出す。
		if asks := g.GetAsks(); len(asks) > 0 {
			b.WriteString(i18n.T("literature.recentAsks") + "\n")
			start := max(0, len(asks)-5)
			for _, a := range asks[start:] {
				if a == nil {
					continue
				}
				key := "literature.askMissLine"
				if a.Success {
					key = "literature.askHitLine"
				}
				b.WriteString("  " + i18n.Tf(key,
					"from", strconv.Itoa(a.From),
					"to", strconv.Itoa(a.To),
					"card", cuiCardStr(a.Card)) + "\n")
			}
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			key := "literature.gameEnd"
			if g.GetWinnerTeam() < 0 {
				key = "literature.gameDrawn"
			}
			banner := i18n.Tf(key, "team", strconv.Itoa(g.GetWinnerTeam()))
			b.WriteString(color.Green(banner) + "\n")
			return
		}

		// **直前の宣言の結末は 3 通り。**
		if c := g.GetLastClaim(); c != nil {
			key := "literature.claimWonLine"
			switch c.Outcome {
			case domain.LiteratureClaimCancelled:
				key = "literature.claimCancelledLine"
			case domain.LiteratureClaimLost:
				key = "literature.claimLostLine"
			}
			b.WriteString(i18n.Tf(key,
				"name", literatureHalfSuitLabel(c.HalfSuit),
				"team", strconv.Itoa(c.AwardedTeam)) + "\n")
		}

		idx := g.GetCurrentPlayerIdx()
		b.WriteString(i18n.Tf("literature.promptTurn", "name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
		b.WriteString(i18n.T("literature.askRules") + "\n")
		b.WriteString(i18n.T("literature.promptHelp") + "\n")
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *LiteratureCuiPresenter) ActionLogOutput(g interfaces.LiteratureGame) string {
	return actionLogOutputTextForSeats[*domain.LiteraturePlayer](g)
}
