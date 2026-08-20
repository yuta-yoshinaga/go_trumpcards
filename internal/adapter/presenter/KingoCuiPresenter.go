//go:build !js || !wasm || extra

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// KingoCuiPresenter キンゴCUIプレゼンタークラス
type KingoCuiPresenter struct{}

// Output ゲーム状態を出力
func (cp *KingoCuiPresenter) Output(c interfaces.KingoGame, lastErr error) string {
	return buildCuiOutput(i18n.T("kingo.outputTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("kingo.roundLine",
			"round", strconv.Itoa(c.GetRoundNumber()),
			"total", strconv.Itoa(c.GetConfig().Rounds)) + "\n")
		sb.WriteString(i18n.Tf("kingo.bankerLine",
			"name", c.GetPlayers()[c.GetBankerSeat()].GetName()) + "\n")
		cp.writeSeats(sb, c)
		cp.writePrompt(sb, c)
		cuiErrorBlock(sb, lastErr)
		cp.writeResult(sb, c)
	})
}

// writeSeats は席ごとの状態を書き出す。
//
// **配る前は誰の手札も無い。** 張りは完全に情報の無い賭けなので、
// 途中で見えるものは張り額だけ。
func (cp *KingoCuiPresenter) writeSeats(sb *strings.Builder, c interfaces.KingoGame) {
	sb.WriteString("----------\n")
	shown := c.GetPhase() != domain.KingoPhaseBet
	for i, p := range c.GetPlayers() {
		mark := " "
		if i == c.GetBankerSeat() {
			mark = "*"
		}
		cards, rank := "-", ""
		if shown && len(p.GetCards()) > 0 {
			cards = kingoCardsStr(p.GetCards())
			rank = i18n.T("kingo.rank." + domain.KingoRankName(p.GetRank()))
			// **同じ「嵐」でも K 3 枚と A 3 枚では強さの実感が違う** (#5783)。
			// 役が付いた席には、そろえた数字まで出す。
			if p.GetRank() > domain.KingoRankNone {
				rank = i18n.Tf("kingo.rankWithValue",
					"rank", rank,
					"value", cuiRankLabel(domain.KingoMatchedValue(p.GetCards())))
			}
		}
		bet := "-"
		if p.GetBet() > 0 {
			bet = strconv.Itoa(p.GetBet())
		}
		sb.WriteString(i18n.Tf("kingo.seatLine",
			"mark", mark,
			"name", p.GetName(),
			"chips", strconv.Itoa(p.GetChips()),
			"bet", bet,
			"cards", cards,
			"rank", rank) + "\n")
	}
}

// writePrompt はいま人間に求められている操作を書き出す。
func (cp *KingoCuiPresenter) writePrompt(sb *strings.Builder, c interfaces.KingoGame) {
	if c.GetGameEndFlag() || c.GetPhase() != domain.KingoPhaseBet {
		return
	}
	if c.IsHumanBanker() {
		sb.WriteString(i18n.T("kingo.dealPrompt") + "\n")
		return
	}
	sb.WriteString(i18n.Tf("kingo.betPrompt",
		"min", strconv.Itoa(c.GetConfig().MinBet)) + "\n")
}

// writeResult は決着を書き出す。
func (cp *KingoCuiPresenter) writeResult(sb *strings.Builder, c interfaces.KingoGame) {
	if c.GetPhase() == domain.KingoPhaseBet {
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
		if r.WonAmount == 0 {
			continue
		}
		line := i18n.Tf("kingo.wonLine",
			"name", players[i].GetName(),
			"amount", strconv.Itoa(r.WonAmount))
		if r.WonAmount > 0 {
			line = color.Green(line)
		}
		sb.WriteString(line + "\n")
	}
	if c.GetGameEndFlag() {
		sb.WriteString(i18n.Tf("kingo.gameEndLine",
			"name", players[c.WinnerSeat()].GetName()) + "\n")
	}
}

// ActionLogOutput 棋譜をテキスト出力
func (cp *KingoCuiPresenter) ActionLogOutput(c interfaces.KingoGame) string {
	return actionLogOutputText(c)
}

// HintOutput ヒントをテキスト出力
func (cp *KingoCuiPresenter) HintOutput(c interfaces.KingoGame) string {
	h := c.GetHint()
	if h == nil {
		return i18n.T("kingo.hintNone")
	}
	return i18n.Tf("kingo.hint",
		"action", i18n.T("kingo.action."+h.Action),
		"reason", i18n.T("kingo.reason."+h.Reason))
}

// kingoCardsStr は札の並びを 1 行の文字列にする。
func kingoCardsStr(cards []*domain.Card) string {
	parts := make([]string, 0, len(cards))
	for _, c := range cards {
		parts = append(parts, cuiCardStr(c))
	}
	return strings.Join(parts, " ")
}
