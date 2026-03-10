package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MemoryWebPresenter 神経衰弱Webプレゼンタークラス
type MemoryWebPresenter struct{}

// NewMemoryWebPresenter コンストラクタ
func NewMemoryWebPresenter() *MemoryWebPresenter {
	return &MemoryWebPresenter{}
}

// Output ゲーム状態をJSON出力
func (p *MemoryWebPresenter) Output(m interfaces.MemoryGame, lastErr error) string {
	resObj := new(controller.MemoryWebOutput)
	resObj.Players = make([]*controller.MemoryWebOutputPlayer, 0)
	resObj.Phase = int(m.GetPhase())
	resObj.CurrentPlayerIdx = m.GetCurrentPlayerIdx()
	resObj.FirstFlipPos = m.GetFirstFlipPos()
	resObj.SecondFlipPos = m.GetSecondFlipPos()
	resObj.LastMatchResult = m.GetLastMatchResult()
	resObj.GameEndFlag = m.GetGameEndFlag()
	resObj.WinnerIdx = m.GetWinnerIdx()
	resObj.TurnNumber = m.GetTurnNumber()

	// 設定
	cfg := m.GetConfig()
	resObj.Config = controller.MemoryWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
	}

	// ボード
	board := m.GetBoard()
	resObj.Board = make([]*controller.MemoryWebOutputBoardCard, domain.MemoryBoardSize)
	for i := 0; i < domain.MemoryBoardSize; i++ {
		bc := board[i]
		outCard := &controller.MemoryWebOutputBoardCard{
			FaceUp: bc.FaceUp,
			Taken:  bc.Taken,
		}
		if bc.FaceUp && !bc.Taken {
			outCard.Card = cardToOutput(bc.Card)
		}
		resObj.Board[i] = outCard
	}

	// プレイヤー情報
	for i := 0; i < m.GetPlayerCnt(); i++ {
		player := m.GetPlayer(i)
		pObj := &controller.MemoryWebOutputPlayer{
			ID:        i,
			IsHuman:   player.GetIsHuman(),
			PairCount: player.GetPairCount(),
		}
		resObj.Players = append(resObj.Players, pObj)
	}

	// メッセージ
	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if m.GetGameEndFlag() {
		resObj.Message = p.buildResultMessage(m)
		winnerIdx := m.GetWinnerIdx()
		player := m.GetPlayer(winnerIdx)
		if player != nil && player.GetIsHuman() {
			resObj.MessageCode = "memory.result.humanWin"
		} else {
			resObj.MessageCode = "memory.result.cpuWin"
			resObj.MessageParams = map[string]string{"cpuId": fmt.Sprintf("%d", winnerIdx)}
		}
	} else {
		phase := m.GetPhase()
		switch phase {
		case domain.MemoryPhaseFlip1:
			resObj.MessageCode = "memory.flip1"
		case domain.MemoryPhaseFlip2:
			resObj.MessageCode = "memory.flip2"
		case domain.MemoryPhaseResult:
			if m.GetLastMatchResult() {
				resObj.MessageCode = "memory.matched"
			} else {
				resObj.MessageCode = "memory.mismatched"
			}
		}
	}

	return marshalOrError(resObj)
}

// buildResultMessage ゲーム終了メッセージを生成
func (p *MemoryWebPresenter) buildResultMessage(m interfaces.MemoryGame) string {
	winnerIdx := m.GetWinnerIdx()
	player := m.GetPlayer(winnerIdx)
	if player == nil {
		return fmt.Sprintf("ゲーム終了！ CPU %dの勝ち！", winnerIdx)
	}
	var name string
	if player.GetIsHuman() {
		name = "あなた"
	} else {
		name = fmt.Sprintf("CPU %d", winnerIdx)
	}
	return fmt.Sprintf("ゲーム終了！ %sの勝ち！", name)
}

// ActionLogOutput 棋譜をJSON出力
func (p *MemoryWebPresenter) ActionLogOutput(m interfaces.MemoryGame) string {
	if !m.GetGameEndFlag() {
		return actionLogToJSON(nil)
	}
	return actionLogToJSON(m.GetActionLog())
}
