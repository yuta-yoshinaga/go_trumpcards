//go:build !js || !wasm || solo

package presenter

import (
	"sort"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MemoryWebPresenter 神経衰弱Webプレゼンタークラス
type MemoryWebPresenter struct{}

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
		// Normalized so a snapshot saved before ADR-0035 (PairCount 0) reports the
		// full deck rather than 0, which the UI would otherwise render as a blank
		// setting.
		PairCount: cfg.NormalizedPairCount(),
	}

	// ボード
	board := m.GetBoard()
	// 盤面の長さはペア数設定で変わる (ADR-0035)。定数で回すと 52 未満の盤面で
	// index out of range になる。
	resObj.Board = make([]*controller.MemoryWebOutputBoardCard, len(board))
	for i := 0; i < len(board); i++ {
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
			Pairs:     memoryCapturedPairs(player),
		}
		resObj.Players = append(resObj.Players, pObj)
	}

	// メッセージ
	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if m.GetGameEndFlag() {
		winnerIdx := m.GetWinnerIdx()
		player := m.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		resObj.Message, resObj.MessageCode, resObj.MessageParams = buildWinnerWebMessage(
			"memory", winnerIdx, isHuman)
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

// ActionLogOutput 棋譜をJSON出力
func (p *MemoryWebPresenter) ActionLogOutput(m interfaces.MemoryGame) string {
	return actionLogOutputJSON(m)
}

// memoryCapturedPairs はプレイヤーが獲得した各ペアの代表カード（各ペアの1枚目）を
// ランク昇順で返す。取得ペアのミニカード表示（issue #3028）のためのデータ。
func memoryCapturedPairs(player *domain.MemoryPlayer) []*controller.WebOutputCard {
	if player == nil {
		return make([]*controller.WebOutputCard, 0)
	}
	pairs := player.GetPairs()
	out := make([]*controller.WebOutputCard, 0, len(pairs))
	for _, pair := range pairs {
		if pair[0] == nil {
			continue
		}
		out = append(out, cardToOutput(pair[0]))
	}
	sort.SliceStable(out, func(a, b int) bool {
		return out[a].Value < out[b].Value
	})
	return out
}
