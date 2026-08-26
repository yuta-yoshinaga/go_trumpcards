//go:build !js || !wasm || extra4

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SevenTwentySevenWebPresenter はセブン・トゥエンティセブン (SevenTwentySeven) の Web プレゼンタークラス。
type SevenTwentySevenWebPresenter struct{}

// Output はゲーム状態を JSON 出力する。
func (p *SevenTwentySevenWebPresenter) Output(g interfaces.SevenTwentySevenGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **SevenTwentySeven.GetHint() は宣言フェーズかつ席 0（NewSevenTwentySevenPlayer(i == 0) で常に人間）に限る。**
	// 他ゲームがそうだから、で済ませない —— Pinochle は見ていなかった (#4585)。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.SevenTwentySevenWebOutputHint{
			Draw:   hint.Draw,
			Reason: hint.Reason,
		}
	}

	return marshalOrError(resObj)
}

// buildBase は基本フィールドを埋めた出力オブジェクトを生成する。
func (p *SevenTwentySevenWebPresenter) buildBase(g interfaces.SevenTwentySevenGame) *controller.SevenTwentySevenWebOutput {
	resObj := new(controller.SevenTwentySevenWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.Pot = g.GetPot()
	resObj.CarryPot = g.GetCarryPot()
	resObj.CarryCount = g.GetCarryCount()
	resObj.Ante = g.GetAnte()
	resObj.Chips = g.GetChips()
	resObj.LowWinner = g.GetLowWinner()
	resObj.HighWinner = g.GetHighWinner()
	resObj.DrawRound = g.GetDrawRound()
	resObj.MatchWinnerIdx = g.GetMatchWinnerIdx()
	resObj.Result = int(g.GetResult())
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.Players = p.buildPlayersOutput(g)

	cfg := g.GetConfig()
	resObj.Config = controller.SevenTwentySevenWebOutputConfig{
		PlayerCount:   cfg.PlayerCount,
		Ante:          cfg.Ante,
		StartingChips: cfg.StartingChips,
		TargetRounds:  cfg.TargetRounds,
	}
	return resObj
}

// buildPlayersOutput はプレイヤー情報を構築する。人間は常に手札公開。結果フェーズでは
// 「イン」で残った (非脱落) プレイヤーの手も公開する。
func (p *SevenTwentySevenWebPresenter) buildPlayersOutput(g interfaces.SevenTwentySevenGame) []*controller.SevenTwentySevenWebOutputPlayer {
	out := make([]*controller.SevenTwentySevenWebOutputPlayer, 0)
	reveal := g.GetPhase() == domain.SevenTwentySevenPhaseResult
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		showCards := player.GetIsHuman() || (reveal && !player.GetOut())
		out = append(out, &controller.SevenTwentySevenWebOutputPlayer{
			ID:        i,
			IsHuman:   player.GetIsHuman(),
			Chips:     player.GetChips(),
			Standing:  player.GetStanding(),
			Out:       player.GetOut(),
			RoundBet:  player.GetRoundBet(),
			CardCount: player.GetCardsSize(),
			Cards:     playerCardsToOutput(player, showCards),
			// **両側の得点を必ず返す。** 片方だけでは、いま何を狙えるのかが
			// ページから読めない。相手の得点は手札が見えている場合のみ。
			LowScore:  sevenTwentySevenSideScore(g, i, domain.SevenTwentySevenSideLow, showCards),
			HighScore: sevenTwentySevenSideScore(g, i, domain.SevenTwentySevenSideHigh, showCards),
			WonLow:    i == g.GetLowWinner(),
			WonHigh:   i == g.GetHighWinner(),
		})
	}
	return out
}

// buildMessage はゲーム結果メッセージを構築する。
func (p *SevenTwentySevenWebPresenter) buildMessage(g interfaces.SevenTwentySevenGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.SevenTwentySevenPhaseDraw:
		return "", "seventwentyseven.drawPhase", nil
	case domain.SevenTwentySevenPhaseResult:
		return p.roundEndMessage(g)
	}
	return "", "", nil
}

// roundEndMessage はラウンド終了時のメッセージを構築する。
func (p *SevenTwentySevenWebPresenter) roundEndMessage(g interfaces.SevenTwentySevenGame) (string, string, map[string]string) {
	low, high := g.GetLowWinner(), g.GetHighWinner()
	if low < 0 && high < 0 {
		return "Everyone busted both ways; the pot carries over.", "seventwentyseven.roundEndCarry", nil
	}
	if low >= 0 && low == high {
		if low == 0 {
			return "You scooped both halves!", "seventwentyseven.roundEndHumanScoop", nil
		}
		params := map[string]string{"player": fmt.Sprintf("%d", low)}
		return fmt.Sprintf("CPU %d scooped both halves.", low), "seventwentyseven.roundEndCpuScoop", params
	}
	switch g.GetResult() {
	case domain.SevenTwentySevenResultWin:
		return "You take a share of the pot.", "seventwentyseven.roundEndHumanWin", nil
	default:
		return "You took neither side.", "seventwentyseven.roundEndHumanLose", nil
	}
}

// winnerMessage は試合終了メッセージを構築する。
func (p *SevenTwentySevenWebPresenter) winnerMessage(g interfaces.SevenTwentySevenGame) (string, string, map[string]string) {
	winner := g.GetMatchWinnerIdx()
	pl := g.GetPlayer(winner)
	if pl != nil && pl.GetIsHuman() {
		return "Game over! You win!", "seventwentyseven.result.humanWin", nil
	}
	params := map[string]string{"player": fmt.Sprintf("%d", winner)}
	return fmt.Sprintf("Game over! CPU %d wins!", winner), "seventwentyseven.result.cpuWin", params
}

// HintOutput はヒント情報を JSON 出力する。
func (p *SevenTwentySevenWebPresenter) HintOutput(g interfaces.SevenTwentySevenGame) string {
	resObj := p.buildBase(g)
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.SevenTwentySevenWebOutputHint{
			Draw:   hint.Draw,
			Reason: hint.Reason,
		}
	}
	// **「頼んだヒントか」を CLI が見分けられるようにする。**このゲーム群の
	// `hintAvailable` は画面のラベルとして既に使われているので、別キーを出す (#4483)。
	if g.GetHint() != nil {
		resObj.MessageCode = "seventwentyseven.hintRequested"
	} else {
		resObj.MessageCode = "seventwentyseven.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput は棋譜を JSON 出力する。
func (p *SevenTwentySevenWebPresenter) ActionLogOutput(g interfaces.SevenTwentySevenGame) string {
	return actionLogOutputJSON(g)
}

// sevenTwentySevenSideScore は side 側の得点を表示用の文字列で返す。
// 手札が見えていない相手には空文字（得点は手札そのものなので、隠すなら両方隠す）。
func sevenTwentySevenSideScore(g interfaces.SevenTwentySevenGame, idx, side int, visible bool) string {
	if !visible {
		return ""
	}
	v, ok := g.GetScore(idx, side)
	if !ok {
		return "-"
	}
	return domain.SevenTwentySevenFormat(v)
}
