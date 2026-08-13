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

// BanLuckCuiPresenter バンラックCUIプレゼンタークラス
type BanLuckCuiPresenter struct{}

// Output ゲーム状態を出力
func (cp *BanLuckCuiPresenter) Output(c interfaces.BanLuckGame, lastErr error) string {
	return buildCuiOutput(i18n.T("banluck.outputTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("banluck.phaseLine", "phase", cp.phaseStr(c.GetPhase())) + "\n")
		sb.WriteString(i18n.Tf("banluck.roundLine",
			"round", strconv.Itoa(c.GetRoundNumber()),
			"total", strconv.Itoa(c.GetConfig().Rounds)) + "\n")
		sb.WriteString(i18n.Tf("banluck.bankerLine",
			"seat", cp.seatName(c, c.GetBankerSeat())) + "\n")
		if c.MustHit() {
			// **義務は必ず名指しする。** 「なぜ止まれないのか」が読めないと、
			// 拒否されたことだけが伝わって規則が伝わらない。
			sb.WriteString(color.Yellow(i18n.Tf("banluck.mustHitLine",
				"min", strconv.Itoa(domain.BanLuckBankerMustHitUnder))) + "\n")
		}
		cp.writeSeats(sb, c)
		cuiErrorBlock(sb, lastErr)
		cp.writeResult(sb, c)
	})
}

// seatName は席の表示名を返す。
func (cp *BanLuckCuiPresenter) seatName(c interfaces.BanLuckGame, i int) string {
	players := c.GetPlayers()
	if i < 0 || i >= len(players) || players[i] == nil {
		return "?"
	}
	return players[i].GetName()
}

// writeSeats は席ごとの手札と状態を書き出す。
func (cp *BanLuckCuiPresenter) writeSeats(sb *strings.Builder, c interfaces.BanLuckGame) {
	players := c.GetPlayers()
	hands := c.GetHands()
	if len(players) == 0 {
		return
	}
	sb.WriteString("----------\n")
	for i, p := range players {
		mark := " "
		if i == c.GetTurnSeat() && c.GetPhase() == domain.BanLuckPhasePlay {
			mark = "*"
		}
		role := ""
		if i == c.GetBankerSeat() {
			role = i18n.T("banluck.bankerMark")
		}
		cards, score := "", ""
		if i < len(hands) && hands[i] != nil {
			cards = banLuckCardsStr(hands[i].GetCards())
			score = strconv.Itoa(hands[i].GetScore())
		}
		sb.WriteString(i18n.Tf("banluck.seatLine",
			"mark", mark,
			"name", p.GetName(),
			"role", role,
			"chips", strconv.Itoa(p.GetChips()),
			"bet", strconv.Itoa(p.GetBet()),
			"cards", cards,
			"score", score) + "\n")
	}
}

// writeResult は決着を書き出す。
func (cp *BanLuckCuiPresenter) writeResult(sb *strings.Builder, c interfaces.BanLuckGame) {
	phase := c.GetPhase()
	if phase != domain.BanLuckPhaseRoundEnd && phase != domain.BanLuckPhaseGameEnd {
		return
	}
	results := c.GetResults()
	players := c.GetPlayers()
	sb.WriteString("----------\n")
	for i, r := range results {
		if i >= len(players) {
			break
		}
		line := i18n.Tf("banluck.resultLine",
			"name", players[i].GetName(),
			"rank", i18n.T("banluck.rank."+domain.BanLuckRankName(r.Rank)),
			"delta", strconv.Itoa(r.Delta))
		switch {
		case r.Delta > 0:
			sb.WriteString(color.Green(line) + "\n")
		case r.Delta < 0:
			sb.WriteString(color.Red(line) + "\n")
		default:
			sb.WriteString(line + "\n")
		}
	}
	if phase == domain.BanLuckPhaseGameEnd {
		sb.WriteString(i18n.Tf("banluck.winnerLine",
			"name", cp.seatName(c, c.WinnerSeat())) + "\n")
	}
}

// ActionLogOutput 棋譜をテキスト出力
func (cp *BanLuckCuiPresenter) ActionLogOutput(c interfaces.BanLuckGame) string {
	return actionLogOutputText(c)
}

// HintOutput ヒントをテキスト出力
func (cp *BanLuckCuiPresenter) HintOutput(c interfaces.BanLuckGame) string {
	h := c.GetHint()
	if h == nil {
		return i18n.T("banluck.hintNone")
	}
	return i18n.Tf("banluck.hint",
		"action", i18n.T("banluck.action."+h.Action),
		"reason", i18n.T("banluck.reason."+h.Reason))
}

// phaseStr フェーズ文字列
func (cp *BanLuckCuiPresenter) phaseStr(phase domain.BanLuckPhase) string {
	switch phase {
	case domain.BanLuckPhaseBet:
		return i18n.T("banluck.phaseBet")
	case domain.BanLuckPhasePlay:
		return i18n.T("banluck.phasePlay")
	case domain.BanLuckPhaseRoundEnd:
		return i18n.T("banluck.phaseRoundEnd")
	case domain.BanLuckPhaseGameEnd:
		return i18n.T("banluck.phaseGameEnd")
	default:
		return i18n.T("banluck.phaseUnknown")
	}
}

// banLuckCardsStr は札の並びを 1 行の文字列にする。
func banLuckCardsStr(cards []*domain.Card) string {
	parts := make([]string, 0, len(cards))
	for _, c := range cards {
		parts = append(parts, cuiCardStr(c))
	}
	return strings.Join(parts, " ")
}
