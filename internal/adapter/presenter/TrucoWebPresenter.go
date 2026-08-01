package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// TrucoWebPresenter トゥルコWebプレゼンタークラス
type TrucoWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *TrucoWebPresenter) Output(g interfaces.TrucoGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**Truco.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.TrucoWebOutputHint{
			Action:    hint.Action,
			CardIndex: hint.CardIndex,
			Reason:    hint.Reason,
		}
	}

	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *TrucoWebPresenter) buildBase(g interfaces.TrucoGame) *controller.TrucoWebOutput {
	resObj := new(controller.TrucoWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.HandNumber = g.GetHandNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.ResponderIdx = g.GetResponderIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.ManoIdx = g.GetManoIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.HandStake = g.GetHandStake()
	resObj.AcceptedLevel = g.GetAcceptedLevel()
	resObj.PendingLevel = g.GetPendingLevel()
	resObj.TrucoCallerIdx = g.GetTrucoCallerIdx()
	resObj.CanDeclareTruco = g.CanDeclareTruco()
	resObj.MatchTarget = g.GetMatchTarget()
	resObj.HandWinnerIdx = g.GetHandWinnerIdx()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()

	results := g.GetTrickResults()
	if results == nil {
		results = make([]int, 0)
	}
	resObj.TrickResults = results

	resObj.MatchPoints = []int{g.GetPlayerMatchPoints(0), g.GetPlayerMatchPoints(1)}

	cfg := g.GetConfig()
	resObj.Config = controller.TrucoWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		MatchTarget:   g.GetMatchTarget(),
	}

	resObj.CurrentTrick = trickCardsToOutput(g.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *TrucoWebPresenter) buildPlayersOutput(g interfaces.TrucoGame) []*controller.TrucoWebOutputPlayer {
	out := make([]*controller.TrucoWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		out = append(out, &controller.TrucoWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			TrickCount: player.GetTrickCount(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *TrucoWebPresenter) buildMessage(g interfaces.TrucoGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		p0 := g.GetPlayerMatchPoints(0)
		p1 := g.GetPlayerMatchPoints(1)
		params := map[string]string{"p0": fmt.Sprintf("%d", p0), "p1": fmt.Sprintf("%d", p1)}
		if g.GetWinnerIdx() == 0 {
			return fmt.Sprintf("マッチ終了！ あなたの勝利です (%d-%d)！", p0, p1), "truco.result.p0Win", params
		}
		return fmt.Sprintf("マッチ終了！ CPUの勝利です (%d-%d)。", p0, p1), "truco.result.p1Win", params
	}
	switch g.GetPhase() {
	case domain.TrucoPhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "truco.playPhase.lead", nil
		}
		return "", "truco.playPhase.follow", nil
	case domain.TrucoPhaseRespond:
		return "", "truco.respondPhase", map[string]string{"level": fmt.Sprintf("%d", g.GetPendingLevel())}
	case domain.TrucoPhaseTrickEnd:
		return "", "truco.trickEnd", nil
	case domain.TrucoPhaseHandEnd:
		params := map[string]string{
			"stake": fmt.Sprintf("%d", g.GetHandStake()),
		}
		if g.GetHandWinnerIdx() == 0 {
			return "", "truco.handEnd.p0", params
		}
		return "", "truco.handEnd.p1", params
	}
	return "", "", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *TrucoWebPresenter) HintOutput(g interfaces.TrucoGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.TrucoWebOutputHint{
			Action:    hint.Action,
			CardIndex: hint.CardIndex,
			Reason:    hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *TrucoWebPresenter) ActionLogOutput(g interfaces.TrucoGame) string {
	return actionLogOutputJSON(g)
}
