//go:build !js || !wasm || extra4

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// RussianBankWebPresenter ロシアンバンク (クラペット) のWebプレゼンタークラス。
type RussianBankWebPresenter struct{}

// Output ゲーム状態をJSON出力。
func (p *RussianBankWebPresenter) Output(g interfaces.RussianBankGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**RussianBank.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.RussianBankWebOutputHint{
			Zone:         int(hint.Zone),
			FromOpponent: hint.FromOpponent,
			Col:          hint.Col,
			ToFoundation: hint.ToFoundation,
			ToCol:        hint.ToCol,
		}
	}

	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築。
func (p *RussianBankWebPresenter) buildBase(g interfaces.RussianBankGame) *controller.RussianBankWebOutput {
	resObj := new(controller.RussianBankWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.CurrentPlayerIdx = g.GetCurrentPlayer()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinner()
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.CanCallStop = g.CanCallStop()
	resObj.CanUndo = g.CanUndo()
	resObj.MoveCount = g.GetMoveCount()
	resObj.Config = controller.RussianBankWebOutputConfig{CpuDifficulty: int(g.GetConfig().CpuDifficulty)}

	tableau := g.GetTableau()
	resObj.Tableau = make([][]*controller.WebOutputCard, len(tableau))
	for i, col := range tableau {
		resObj.Tableau[i] = cardsToOutputOrEmpty(col)
	}
	foundations := g.GetFoundations()
	resObj.Foundations = make([][]*controller.WebOutputCard, len(foundations))
	for i, f := range foundations {
		resObj.Foundations[i] = cardsToOutputOrEmpty(f)
	}
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築。手札と裏向きリザーブの中身は公開しない。
func (p *RussianBankWebPresenter) buildPlayersOutput(g interfaces.RussianBankGame) []*controller.RussianBankWebOutputPlayer {
	out := make([]*controller.RussianBankWebOutputPlayer, 0)
	for i, player := range g.GetPlayers() {
		if player == nil {
			continue
		}
		out = append(out, &controller.RussianBankWebOutputPlayer{
			ID:           i,
			IsHuman:      !player.IsCPU(),
			ReserveCount: player.ReserveSize(),
			ReserveTop:   cardToOutput(player.ReserveTop()),
			HandCount:    player.HandSize(),
			WasteCount:   player.WasteSize(),
			WasteTop:     cardToOutput(player.WasteTop()),
			StopPoints:   g.GetStopPoints(i),
		})
	}
	return out
}

// buildMessage 結果/フェーズメッセージを構築。
func (p *RussianBankWebPresenter) buildMessage(g interfaces.RussianBankGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	// **なぜボタンが増えたのかを言う。**CUI は同じ状態を黄色で明示しているのに、
	// Web は無言で Stop ボタンが現れるだけだった (#4817)。
	if g.CanCallStop() {
		return "", "russianbank.stopAvailable", nil
	}
	return "", "russianbank.playing", nil
}

// winnerMessage 勝者メッセージを構築する。
func (p *RussianBankWebPresenter) winnerMessage(g interfaces.RussianBankGame) (string, string, map[string]string) {
	winner := g.GetWinner()
	if winner < 0 {
		return "", "russianbank.result.draw", nil
	}
	if player := g.GetPlayer(winner); player != nil && !player.IsCPU() {
		return "", "russianbank.result.humanWin", nil
	}
	params := map[string]string{"player": fmt.Sprintf("%d", winner)}
	return "", "russianbank.result.cpuWin", params
}

// HintOutput ヒント情報をJSON出力する。
func (p *RussianBankWebPresenter) HintOutput(g interfaces.RussianBankGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.RussianBankWebOutputHint{
			Zone:         int(hint.Zone),
			FromOpponent: hint.FromOpponent,
			Col:          hint.Col,
			ToFoundation: hint.ToFoundation,
			ToCol:        hint.ToCol,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力。
func (p *RussianBankWebPresenter) ActionLogOutput(g interfaces.RussianBankGame) string {
	return actionLogOutputJSON(g)
}
