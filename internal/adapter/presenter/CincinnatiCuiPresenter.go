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

// CincinnatiCuiPresenter シンシナティCUIプレゼンタークラス
type CincinnatiCuiPresenter struct{}

// Output ゲーム状態を出力
func (cp *CincinnatiCuiPresenter) Output(c interfaces.CincinnatiGame, lastErr error) string {
	return buildCuiOutput(i18n.T("cincinnati.outputTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("cincinnati.phaseLine", "phase", cp.phaseStr(c.GetPhase())) + "\n")
		sb.WriteString(i18n.Tf("cincinnati.handLine",
			"hand", strconv.Itoa(c.GetHandNumber()),
			"pot", strconv.Itoa(c.GetPot())) + "\n")
		cp.writeCommunity(sb, c)
		cp.writeSeats(sb, c)
		cp.writeBetLine(sb, c)
		cuiErrorBlock(sb, lastErr)
		cp.writeResult(sb, c)
	})
}

// writeCommunity は表向きのコミュニティと残り枚数を書き出す。
//
// **残り何枚めくれるかを出す。** 1 枚ずつ 5 回という進行がこのゲームの形なので、
// あと何回ベットがあるのかが読めないと降りどころが決められない。
func (cp *CincinnatiCuiPresenter) writeCommunity(sb *strings.Builder, c interfaces.CincinnatiGame) {
	sb.WriteString("----------\n")
	cards := c.GetCommunityCards()
	shown := "-"
	if len(cards) > 0 {
		shown = cincinnatiCardsStr(cards)
	}
	sb.WriteString(i18n.Tf("cincinnati.communityLine",
		"cards", shown,
		"revealed", strconv.Itoa(c.GetRevealedCount()),
		"total", strconv.Itoa(domain.CincinnatiCommunityCards)) + "\n")
}

// writeSeats は席ごとの状態を書き出す。
//
// **CPU の手札は伏せたまま。** 5 枚もあるので見えたら勝負にならない。
func (cp *CincinnatiCuiPresenter) writeSeats(sb *strings.Builder, c interfaces.CincinnatiGame) {
	showdown := c.GetPhase() == domain.CincinnatiPhaseShowdown || c.GetGameEndFlag()
	for i, p := range c.GetPlayers() {
		mark := " "
		if i == c.GetTurnSeat() && c.GetPhase() == domain.CincinnatiPhaseBetting {
			mark = "*"
		}
		cards := i18n.T("cincinnati.hidden")
		if p.GetIsHuman() || showdown {
			cards = cincinnatiCardsStr(p.GetCards())
		}
		state := ""
		if p.GetFolded() {
			state = i18n.T("cincinnati.foldedMark")
		} else if p.GetAllIn() {
			state = i18n.T("cincinnati.allInMark")
		}
		sb.WriteString(i18n.Tf("cincinnati.seatLine",
			"mark", mark,
			"name", p.GetName(),
			"state", state,
			"chips", strconv.Itoa(p.GetChips()),
			"bet", strconv.Itoa(p.GetCurrentBet()),
			"cards", cards) + "\n")
	}
}

// writeBetLine はいま人間が払うべき額を書き出す。
func (cp *CincinnatiCuiPresenter) writeBetLine(sb *strings.Builder, c interfaces.CincinnatiGame) {
	if c.GetPhase() != domain.CincinnatiPhaseBetting {
		return
	}
	if toCall := c.GetToCall(); toCall > 0 {
		sb.WriteString(i18n.Tf("cincinnati.toCallLine", "amount", strconv.Itoa(toCall)) + "\n")
		return
	}
	sb.WriteString(i18n.T("cincinnati.canCheckLine") + "\n")
}

// writeResult は決着を書き出す。
func (cp *CincinnatiCuiPresenter) writeResult(sb *strings.Builder, c interfaces.CincinnatiGame) {
	if c.GetPhase() != domain.CincinnatiPhaseShowdown && !c.GetGameEndFlag() {
		return
	}
	results := c.GetResults()
	players := c.GetPlayers()
	if len(results) == 0 {
		return
	}
	sb.WriteString("----------\n")
	for i, r := range results {
		if i >= len(players) {
			break
		}
		if r.WonAmount <= 0 {
			continue
		}
		// **なぜその配当になったのかを言う** (#5780)。5 枚の手札だけで成立する
		// 役も普通にあるので、金額だけでは読めない。
		// 役名は評価器が付けたランクから引く。範囲外は番号のまま出す
		// ——Web 側の handLabel と同じ振る舞いで、黙って消えないようにする。
		rank := players[i].GetHandRank()
		hand := strconv.Itoa(rank)
		if rank >= 0 && rank < len(domain.PokerHandNames) {
			hand = domain.PokerHandNames[rank]
		}
		sb.WriteString(color.Green(i18n.Tf("cincinnati.wonLine",
			"name", players[i].GetName(),
			"amount", strconv.Itoa(r.WonAmount),
			"hand", hand)) + "\n")
	}
	if c.GetGameEndFlag() {
		sb.WriteString(i18n.Tf("cincinnati.gameEndLine",
			"name", players[c.WinnerSeat()].GetName()) + "\n")
	}
}

// ActionLogOutput 棋譜をテキスト出力
func (cp *CincinnatiCuiPresenter) ActionLogOutput(c interfaces.CincinnatiGame) string {
	return actionLogOutputText(c)
}

// HintOutput ヒントをテキスト出力
func (cp *CincinnatiCuiPresenter) HintOutput(c interfaces.CincinnatiGame) string {
	h := c.GetHint()
	if h == nil {
		return i18n.T("cincinnati.hintNone")
	}
	return i18n.Tf("cincinnati.hint",
		"action", i18n.T("cincinnati.action."+h.Action),
		"reason", i18n.T("cincinnati.reason."+h.Reason))
}

// phaseStr フェーズ文字列
func (cp *CincinnatiCuiPresenter) phaseStr(phase domain.CincinnatiPhase) string {
	switch phase {
	case domain.CincinnatiPhaseDeal:
		return i18n.T("cincinnati.phaseDeal")
	case domain.CincinnatiPhaseBetting:
		return i18n.T("cincinnati.phaseBetting")
	case domain.CincinnatiPhaseShowdown:
		return i18n.T("cincinnati.phaseShowdown")
	case domain.CincinnatiPhaseGameEnd:
		return i18n.T("cincinnati.phaseGameEnd")
	default:
		return i18n.T("cincinnati.phaseUnknown")
	}
}

// cincinnatiCardsStr は札の並びを 1 行の文字列にする。
func cincinnatiCardsStr(cards []*domain.Card) string {
	parts := make([]string, 0, len(cards))
	for _, c := range cards {
		parts = append(parts, cuiCardStr(c))
	}
	return strings.Join(parts, " ")
}
