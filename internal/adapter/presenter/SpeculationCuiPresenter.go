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

// speculationDisplayRound は画面に出すラウンド番号。
//
// **決着後は「今終わった回」を出す。** roundNo は決着で 1 進むので、そのまま
// +1 すると round 1 の結果画面に「ラウンド 2 / 5」と出て、まだ始めていない回の
// 結果を見ているように読める。
func speculationDisplayRound(c interfaces.SpeculationGame) int {
	n := c.GetRoundNo()
	switch c.GetPhase() {
	case domain.SpeculationPhaseResult, domain.SpeculationPhaseGameEnd:
		return n
	default:
		return n + 1
	}
}

// speculationSuitName は切り札スートの記号。**まだ決まっていなければ "-"** ——
// 0 は ♠ の正当な値なので、未決定を 0 と同じに描くとスペードが切り札に見える。
func speculationSuitName(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return "♠"
	case domain.CardDesignClover:
		return "♣"
	case domain.CardDesignHeart:
		return "♥"
	case domain.CardDesignDiamond:
		return "♦"
	}
	return "-"
}

// SpeculationCuiPresenter スペキュレーションCUIプレゼンタークラス
type SpeculationCuiPresenter struct {
}

// Output ゲーム状態を出力
func (cp *SpeculationCuiPresenter) Output(c interfaces.SpeculationGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	sb.WriteString(i18n.Tf("speculation.roundLine",
		"round", strconv.Itoa(speculationDisplayRound(c)),
		"total", strconv.Itoa(c.GetConfig().Rounds)) + "\n")
	sb.WriteString(i18n.Tf("speculation.potLine", "pot", strconv.Itoa(c.GetPot())) + "\n")
	sb.WriteString(i18n.Tf("speculation.trumpLine", "trump", speculationSuitName(c.GetTrumpSuit())) + "\n")
	sb.WriteString(i18n.Tf("speculation.phaseLine", "phase", cp.phaseStr(c.GetPhase())) + "\n")

	cp.writeSeats(&sb, c)

	if c.GetPhase() == domain.SpeculationPhaseAuction {
		cp.writeOffer(&sb, c)
	}
	sb.WriteString("----------\n")

	if lastErr != nil {
		sb.WriteString(i18n.MarkErrorLine(color.Red(lastErr.Error())) + "\n")
	}
	if c.GetPhase() == domain.SpeculationPhaseResult || c.GetGameEndFlag() {
		cp.writeResult(&sb, c)
	}
	return sb.String()
}

// writeSeats は各席の公開情報を出す。
//
// **伏せ札の中身は出さない。** 枚数だけ ── 中身が見えていたら、いくら出すべきかが
// 盤面から丸見えになり、競りが競りでなくなる。
func (cp *SpeculationCuiPresenter) writeSeats(sb *strings.Builder, c interfaces.SpeculationGame) {
	best := c.GetBestSeat()
	turn := c.GetTurnSeat()
	for seat, p := range c.GetPlayers() {
		line := i18n.Tf("speculation.seatLine",
			"name", p.GetName(),
			"chips", strconv.Itoa(p.GetChips()),
			"hidden", strconv.Itoa(p.GetHiddenCount()))
		if b := p.GetBest(); b != nil {
			line += " " + i18n.Tf("speculation.holdsLine", "card", cuiCardStr(b))
		}
		switch seat {
		case best:
			line = color.Yellow(line)
		case turn:
			line = color.Bold(line)
		}
		sb.WriteString(line + "\n")
	}
}

// writeOffer は競りの申し出を出す。
func (cp *SpeculationCuiPresenter) writeOffer(sb *strings.Builder, c interfaces.SpeculationGame) {
	from, to := c.GetOfferFrom(), c.GetOfferTo()
	players := c.GetPlayers()
	if from < 0 || to < 0 || from >= len(players) || to >= len(players) {
		return
	}
	key := "speculation.offerToYou"
	if to != 0 {
		key = "speculation.offerFromYou"
	}
	sb.WriteString(color.Yellow(i18n.Tf(key,
		"buyer", players[from].GetName(),
		"owner", players[to].GetName(),
		"amount", strconv.Itoa(c.GetOfferAmount()))) + "\n")
}

// writeResult はラウンドの決着を出す。
func (cp *SpeculationCuiPresenter) writeResult(sb *strings.Builder, c interfaces.SpeculationGame) {
	winner := c.GetWinnerSeat()
	players := c.GetPlayers()
	switch {
	case winner < 0 || winner >= len(players):
		// **流局は「誰かが負けた」ではない。** 切り札が 1 枚も出なければ
		// 参加料は戻る。
		sb.WriteString(i18n.T("speculation.voidRound") + "\n")
	case winner == 0:
		sb.WriteString(color.Green(i18n.T("speculation.youWin")) + "\n")
	default:
		sb.WriteString(color.Red(i18n.Tf("speculation.seatWins", "name", players[winner].GetName())) + "\n")
	}
	if c.GetGameEndFlag() {
		sb.WriteString(i18n.Tf("speculation.finalChips",
			"chips", strconv.Itoa(players[0].GetChips())) + "\n")
	}
	sb.WriteString("----------\n")
}

// ActionLogOutput 棋譜をテキスト出力
func (cp *SpeculationCuiPresenter) ActionLogOutput(c interfaces.SpeculationGame) string {
	return actionLogOutputText(c)
}

// HintOutput は次の一手を助言する。
//
// **競りでは「札の強さ」ではなく「残り枚数」で決める。** まだ大量に伏せ札が
// 残っているうちは、いくら強い札でも上を出される。
func (cp *SpeculationCuiPresenter) HintOutput(c interfaces.SpeculationGame) string {
	switch c.GetPhase() {
	case domain.SpeculationPhaseAuction:
		return color.Yellow(i18n.T(speculationAuctionHintKey(c))) + "\n"
	case domain.SpeculationPhaseFlip:
		return color.Yellow(i18n.T("speculation.hintFlip")) + "\n"
	default:
		return i18n.T("speculation.hintNone") + "\n"
	}
}

// speculationAuctionHintKey は競り中の助言キーを返す。
func speculationAuctionHintKey(c interfaces.SpeculationGame) string {
	remaining := 0
	for _, p := range c.GetPlayers() {
		remaining += p.GetHiddenCount()
	}
	if c.GetOfferTo() == 0 {
		// 人間が売り手。残りが多いなら売ってしまってよい。
		if remaining > len(c.GetPlayers()) {
			return "speculation.hintSell"
		}
		return "speculation.hintHold"
	}
	// 人間が買い手。
	if remaining <= len(c.GetPlayers()) {
		return "speculation.hintBuy"
	}
	return "speculation.hintPass"
}

// phaseStr フェーズ文字列
func (cp *SpeculationCuiPresenter) phaseStr(phase domain.SpeculationPhase) string {
	switch phase {
	case domain.SpeculationPhaseFlip:
		return i18n.T("speculation.phaseFlip")
	case domain.SpeculationPhaseAuction:
		return i18n.T("speculation.phaseAuction")
	case domain.SpeculationPhaseResult:
		return i18n.T("speculation.phaseResult")
	case domain.SpeculationPhaseGameEnd:
		return i18n.T("speculation.phaseGameEnd")
	default:
		return i18n.T("speculation.phaseUnknown")
	}
}
