package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SixCardGolfWebPresenter SixCardGolf Webプレゼンタークラス
type SixCardGolfWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *SixCardGolfWebPresenter) Output(g interfaces.SixCardGolfGame, lastErr error) string {
	resObj := new(controller.SixCardGolfWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.DrawPileCount = g.GetDrawPileCount()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.FinalTurnTrigger = g.GetFinalTurnTrigger()
	resObj.DrawnFromDiscard = g.GetDrawnFromDiscard()
	resObj.CanFlip = g.GetCanFlip()

	cfg := g.GetConfig()
	resObj.TotalRounds = cfg.Rounds
	resObj.Config = controller.SixCardGolfWebOutputConfig{
		PlayerCount:   cfg.PlayerCount,
		CpuDifficulty: int(cfg.CpuDifficulty),
		Rounds:        cfg.Rounds,
	}

	top := g.GetDiscardTop()
	if top != nil {
		resObj.DiscardTop = cardToOutput(top)
	}

	drawn := g.GetDrawnCard()
	if drawn != nil {
		resObj.DrawnCard = cardToOutput(drawn)
	}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// buildPlayersOutput プレイヤー情報を構築
func (p *SixCardGolfWebPresenter) buildPlayersOutput(g interfaces.SixCardGolfGame) []*controller.SixCardGolfWebOutputPlayer {
	phase := g.GetPhase()
	revealAll := phase == domain.SixCardGolfPhaseRoundOver || phase == domain.SixCardGolfPhaseGameOver
	out := make([]*controller.SixCardGolfWebOutputPlayer, 0, g.GetPlayerCnt())
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		grid := make([]*controller.SixCardGolfWebOutputSlot, domain.SixCardGolfGridSize)
		for j := 0; j < domain.SixCardGolfGridSize; j++ {
			slot := player.Grid[j]
			s := &controller.SixCardGolfWebOutputSlot{FaceUp: slot.FaceUp}
			if slot.FaceUp || revealAll {
				s.Card = cardToOutput(slot.Card)
				if revealAll {
					s.FaceUp = true
				}
			}
			grid[j] = s
		}
		pObj := &controller.SixCardGolfWebOutputPlayer{
			ID:              i,
			IsHuman:         !player.IsCpu,
			Grid:            grid,
			RoundScore:      player.RoundScore,
			CumulativeScore: player.CumulativeScore,
			AllFaceUp:       player.AllFaceUp(),
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage メッセージ構築
func (p *SixCardGolfWebPresenter) buildMessage(g interfaces.SixCardGolfGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerIdx := g.GetWinnerIdx()
		player := g.GetPlayer(winnerIdx)
		isHuman := player != nil && !player.IsCpu
		return buildWinnerWebMessage("sixcardgolf", winnerIdx, isHuman)
	}
	switch g.GetPhase() {
	case domain.SixCardGolfPhaseSetup:
		return "", "sixcardgolf.setup", nil
	case domain.SixCardGolfPhasePlayerTurn:
		if g.GetCanFlip() {
			return "", "sixcardgolf.canFlip", nil
		}
		return "", "sixcardgolf.playerTurn", nil
	case domain.SixCardGolfPhaseDrawPending:
		return "", "sixcardgolf.drawPending", nil
	case domain.SixCardGolfPhaseRoundOver:
		return "", "sixcardgolf.roundOver", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜をJSON出力
func (p *SixCardGolfWebPresenter) ActionLogOutput(g interfaces.SixCardGolfGame) string {
	return actionLogOutputJSON(g)
}

// HintOutput はヒントを返す。Web ではクライアント側でヒントを算出するため、
// 状態出力にフォールバックする (CUI プレゼンターのみが専用ヒントを返す)。
func (p *SixCardGolfWebPresenter) HintOutput(g interfaces.SixCardGolfGame) string {
	return p.Output(g, nil)
}
