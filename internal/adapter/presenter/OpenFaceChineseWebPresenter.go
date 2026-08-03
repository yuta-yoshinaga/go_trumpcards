//go:build !js || !wasm || casino

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// OpenFaceChineseWebPresenter オープンフェイス・チャイニーズポーカー (OFC) のWebプレゼンタークラス
type OpenFaceChineseWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *OpenFaceChineseWebPresenter) Output(g interfaces.OpenFaceChineseGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**OpenFaceChinese.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.OpenFaceChineseWebOutputHint{
			Row:    hint.Row,
			Reason: hint.Reason,
		}
	}

	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *OpenFaceChineseWebPresenter) buildBase(g interfaces.OpenFaceChineseGame) *controller.OpenFaceChineseWebOutput {
	resObj := new(controller.OpenFaceChineseWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.IsHumanTurn = g.IsHumanTurn()
	if g.IsHumanTurn() {
		resObj.CurrentCard = cardToOutput(g.GetCurrentCard())
	}

	cfg := g.GetConfig()
	resObj.Config = controller.OpenFaceChineseWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PlayerCount:   cfg.PlayerCount,
		TargetRounds:  cfg.TargetRounds,
	}

	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *OpenFaceChineseWebPresenter) buildPlayersOutput(g interfaces.OpenFaceChineseGame) []*controller.OpenFaceChineseWebOutputPlayer {
	out := make([]*controller.OpenFaceChineseWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		out = append(out, &controller.OpenFaceChineseWebOutputPlayer{
			ID:          i,
			IsHuman:     player.GetIsHuman(),
			Front:       cardsToOutputOrEmpty(player.GetFront()),
			Middle:      cardsToOutputOrEmpty(player.GetMiddle()),
			Back:        cardsToOutputOrEmpty(player.GetBack()),
			Pending:     p.pendingOutput(player),
			RoundScore:  player.GetRoundScore(),
			Royalty:     player.GetRoyalty(),
			Fouled:      player.GetFouled(),
			Fantasyland: player.GetFantasyland(),
			TotalScore:  player.GetTotalScore(),
		})
	}
	return out
}

// pendingOutput 人間の保留カードのみ可視化する（CPU の保留は隠す）。
func (p *OpenFaceChineseWebPresenter) pendingOutput(player *domain.OpenFaceChinesePlayer) []*controller.WebOutputCard {
	if !player.GetIsHuman() {
		return make([]*controller.WebOutputCard, 0)
	}
	return cardsToOutputOrEmpty(player.GetPending())
}

// buildMessage ゲーム結果メッセージを構築
func (p *OpenFaceChineseWebPresenter) buildMessage(g interfaces.OpenFaceChineseGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.OpenFaceChinesePhasePlacing:
		return "", "openfacechinese.placing", nil
	case domain.OpenFaceChinesePhaseRoundEnd:
		return "", "openfacechinese.roundEnd", nil
	}
	return "", "", nil
}

// winnerMessage 勝者メッセージを構築する
func (p *OpenFaceChineseWebPresenter) winnerMessage(g interfaces.OpenFaceChineseGame) (string, string, map[string]string) {
	winner := g.GetWinnerIdx()
	if winner < 0 {
		return "", "openfacechinese.result.draw", nil
	}
	if player := g.GetPlayer(winner); player != nil && player.GetIsHuman() {
		return "", "openfacechinese.result.humanWin", nil
	}
	params := map[string]string{"player": fmt.Sprintf("%d", winner)}
	return "", "openfacechinese.result.cpuWin", params
}

// HintOutput ヒント情報をJSON出力する
func (p *OpenFaceChineseWebPresenter) HintOutput(g interfaces.OpenFaceChineseGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.OpenFaceChineseWebOutputHint{
			Row:    hint.Row,
			Reason: hint.Reason,
		}
	}
	// **「頼んだヒントか」をフロントが見分けられるようにする。**ページは
	// `isRequestedHint` でこのコードを見てからバナーを出すので (#4605)、
	// 付けないと押しても何も出ない。`hintAvailable` は画面のラベルとして
	// 既に使われているため別キーにする (#4483)。
	if hint != nil {
		resObj.MessageCode = "openfacechinese.hintRequested"
	} else {
		resObj.MessageCode = "openfacechinese.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *OpenFaceChineseWebPresenter) ActionLogOutput(g interfaces.OpenFaceChineseGame) string {
	return actionLogOutputJSON(g)
}
