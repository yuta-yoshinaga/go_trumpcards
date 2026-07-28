//go:build !js || !wasm || extra3

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CribbageWebPresenter クリベッジWebプレゼンタークラス
type CribbageWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *CribbageWebPresenter) Output(g interfaces.CribbageGame, lastErr error) string {
	resObj := new(controller.CribbageWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.PegCount = g.GetPegCount()
	resObj.ShowPhaseStep = g.GetShowPhaseStep()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()

	// スターターカード
	starter := g.GetStarter()
	if starter != nil {
		resObj.Starter = cardToOutput(starter)
	}

	// クリブ
	crib := g.GetCrib()
	resObj.Crib = make([]*controller.WebOutputCard, 0, len(crib))
	phase := g.GetPhase()
	showCrib := phase == domain.CribbagePhaseShow || phase == domain.CribbagePhaseRoundEnd || phase == domain.CribbagePhaseGameEnd
	if showCrib {
		for _, card := range crib {
			resObj.Crib = append(resObj.Crib, cardToOutput(card))
		}
	}

	// ペギングで出されたカード
	pegCards := g.GetPegPlayedCards()
	resObj.PegPlayedCards = make([]*controller.WebOutputCard, 0, len(pegCards))
	for _, card := range pegCards {
		resObj.PegPlayedCards = append(resObj.PegPlayedCards, cardToOutput(card))
	}

	// ハンドスコア詳細
	details := g.GetHandScoreDetails()
	for i, d := range details {
		if d != nil {
			resObj.HandScoreDetails[i] = &controller.CribbageWebOutputScoreDetail{
				Fifteens: d.Fifteens,
				Pairs:    d.Pairs,
				Runs:     d.Runs,
				Flush:    d.Flush,
				Nobs:     d.Nobs,
				Total:    d.Total,
			}
		}
	}

	// 設定
	cfg := g.GetConfig()
	resObj.Config = controller.CribbageWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PointLimit:    cfg.PointLimit,
	}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// buildPlayersOutput プレイヤー情報を構築
func (p *CribbageWebPresenter) buildPlayersOutput(g interfaces.CribbageGame) []*controller.CribbageWebOutputPlayer {
	out := make([]*controller.CribbageWebOutputPlayer, 0)
	for i := range domain.CribbagePlayerCnt {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		showCards := player.GetIsHuman()
		// ショー・ラウンド終了・ゲーム終了時はCPUの手札も表示
		phase := g.GetPhase()
		if phase == domain.CribbagePhaseShow || phase == domain.CribbagePhaseRoundEnd || phase == domain.CribbagePhaseGameEnd {
			showCards = true
		}
		pObj := &controller.CribbageWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, showCards),
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *CribbageWebPresenter) buildMessage(g interfaces.CribbageGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerIdx := g.GetWinnerIdx()
		player := g.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("cribbage", winnerIdx, isHuman)
	}
	switch g.GetPhase() {
	case domain.CribbagePhaseDiscard:
		return "", "cribbage.discardPhase", nil
	case domain.CribbagePhaseCut:
		return "", "cribbage.cutPhase", nil
	case domain.CribbagePhasePegging:
		return "", "cribbage.peggingPhase", nil
	case domain.CribbagePhaseShow:
		return "", "cribbage.showPhase", nil
	case domain.CribbagePhaseRoundEnd:
		return "", "cribbage.roundEnd", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜をJSON出力
func (p *CribbageWebPresenter) ActionLogOutput(g interfaces.CribbageGame) string {
	return actionLogOutputJSON(g)
}

// HintOutput ヒント出力。Web GUI は useGameHint のフロントエンドヒントを使うため、
// サーバーヒントは状態出力をそのまま返す（Web ルートからは呼ばれない）。
func (p *CribbageWebPresenter) HintOutput(g interfaces.CribbageGame) string {
	return p.Output(g, nil)
}
