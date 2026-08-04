//go:build !js || !wasm || extra3

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// MaoWebPresenter マオWebプレゼンタークラス
type MaoWebPresenter struct{}

// Output ゲーム状態をJSON出力する。
// 隠しルール本体 (トリガー・宣言語) はクライアントに公開せず、3回正解で
// 解放されたハーフヒント (GetRuleHintKey) のみを出力に含める。
func (p *MaoWebPresenter) Output(g interfaces.MaoGame, lastErr error) string {
	resObj := new(controller.MaoWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.DrawPileCount = g.GetDrawPileCount()
	resObj.ChosenSuit = g.GetChosenSuit()
	resObj.PenaltyDrawCount = g.GetPenaltyDrawCount()
	resObj.Direction = g.GetDirection()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()

	// 隠しルール関連: ルール本体は決して出力しない。
	resObj.AwaitingWord = g.GetAwaitingWord()
	resObj.CorrectCount = g.GetPlayerCorrectCount()
	resObj.HintUnlocked = g.GetHintUnlocked()
	// **コードも一緒に返す。**Web サーバの i18n 言語はプロセス全体で 1 つなので、
	// 文言だけ返すとブラウザが英語でも日本語のまま届く。翻訳はフロントで行い、
	// RuleHint はそれ以外のクライアント向けのフォールバックとして残す (#4917)。
	resObj.RuleHintCode = g.GetRuleHintKey() // 未解放なら空文字
	if resObj.RuleHintCode != "" {
		resObj.RuleHint = i18n.T("mao." + resObj.RuleHintCode)
	}
	resObj.RulePenalty = g.GetRulePenaltyFlag()

	top := g.GetDiscardTop()
	if top != nil {
		resObj.DiscardTop = cardToOutput(top)
	}

	cfg := g.GetConfig()
	resObj.Config = controller.MaoWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PointLimit:    cfg.PointLimit,
	}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// buildPlayersOutput プレイヤー情報を構築
func (p *MaoWebPresenter) buildPlayersOutput(g interfaces.MaoGame) []*controller.MaoWebOutputPlayer {
	out := make([]*controller.MaoWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		pObj := &controller.MaoWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, player.GetIsHuman()),
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
			HasDeclared:     player.GetHasDeclared(),
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *MaoWebPresenter) buildMessage(g interfaces.MaoGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerIdx := g.GetWinnerIdx()
		player := g.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("mao", winnerIdx, isHuman)
	}
	if g.GetRulePenaltyFlag() {
		return "", "mao.rulePenalty", nil
	}
	if g.GetAwaitingWord() {
		return "", "mao.awaitingWord", nil
	}
	switch g.GetPhase() {
	case domain.MaoPhasePlay:
		return "", "mao.playPhase", nil
	case domain.MaoPhaseChooseSuit:
		return "", "mao.chooseSuitPhase", nil
	case domain.MaoPhaseMustDeclare:
		return "", "mao.mustDeclarePhase", nil
	case domain.MaoPhaseRoundEnd:
		return "", "mao.roundEnd", nil
	case domain.MaoPhaseGameEnd:
		// Handled by the GetGameEndFlag() branch above; explicit for clarity.
		return "", "", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜をJSON出力
func (p *MaoWebPresenter) ActionLogOutput(g interfaces.MaoGame) string {
	return actionLogOutputJSON(g)
}
