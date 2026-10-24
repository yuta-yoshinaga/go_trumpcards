//go:build !js || !wasm || extra4

package presenter

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MrsMopWebPresenter ミセス・モップソリティアWebプレゼンタークラス
type MrsMopWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *MrsMopWebPresenter) Output(s interfaces.MrsMopGame, lastErr error) string {
	resObj := new(controller.MrsMopWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, s, int(s.GetPhase()))
	resObj.StockCount = s.GetStockCount()
	resObj.CompletedSuits = s.GetCompletedSuits()
	resObj.Score = s.GetScore()
	resObj.Difficulty = int(s.GetDifficulty())
	resObj.LegalTargets = buildMrsMopLegalTargets(s)
	resObj.RunStarts = buildMrsMopRunStarts(s)

	// タブロー
	tableau := s.GetTableau()
	resObj.Tableau = make([][]*controller.MrsMopWebOutputTableauCard, domain.MrsMopTableauCnt)
	for i := range domain.MrsMopTableauCnt {
		colCards := tableau[i]
		resObj.Tableau[i] = make([]*controller.MrsMopWebOutputTableauCard, len(colCards))
		for j, tc := range colCards {
			outTC := &controller.MrsMopWebOutputTableauCard{FaceUp: tc.FaceUp}
			if tc.FaceUp {
				outTC.Card = cardToOutput(tc.Card)
			}
			resObj.Tableau[i][j] = outTC
		}
	}

	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if s.GetPhase() == domain.MrsMopPhasePlaying && !s.IsStalemate() {
		if hint := s.GetHint(); hint != nil {
			resObj.Hint = &controller.MrsMopWebOutputHint{
				FromCol:   hint.FromCol,
				CardIndex: hint.CardIndex,
				ToCol:     hint.ToCol,
			}
		}
	}

	// メッセージ
	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		phase := s.GetPhase()
		switch phase {
		case domain.MrsMopPhasePlaying:
			if s.IsStalemate() {
				if resObj.UndoToEscape > 0 {
					resObj.MessageCode = "mrsmop.stalemateWithEscape"
					resObj.MessageParams = map[string]string{"count": fmt.Sprintf("%d", resObj.UndoToEscape)}
				} else {
					resObj.MessageCode = "mrsmop.stalemate"
				}
			} else {
				resObj.MessageCode = "mrsmop.playing"
			}
		case domain.MrsMopPhaseGameClear:
			resObj.Message = fmt.Sprintf("ゲームクリア！ 手数: %d スコア: %d", s.GetMoveCount(), s.GetScore())
			resObj.MessageCode = "mrsmop.gameClear"
			resObj.MessageParams = map[string]string{
				"moveCount": fmt.Sprintf("%d", s.GetMoveCount()),
				"score":     fmt.Sprintf("%d", s.GetScore()),
			}
		case domain.MrsMopPhaseGameOver:
			resObj.Message = "ゲームオーバー"
			resObj.MessageCode = "mrsmop.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *MrsMopWebPresenter) HintOutput(s interfaces.MrsMopGame) string {
	hint := s.GetHint()
	resObj := new(controller.MrsMopWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, s, int(s.GetPhase()))
	resObj.StockCount = s.GetStockCount()
	resObj.CompletedSuits = s.GetCompletedSuits()
	resObj.Score = s.GetScore()
	resObj.Difficulty = int(s.GetDifficulty())
	resObj.Tableau = make([][]*controller.MrsMopWebOutputTableauCard, 0)
	resObj.LegalTargets = buildMrsMopLegalTargets(s)
	resObj.RunStarts = buildMrsMopRunStarts(s)

	if hint != nil {
		resObj.Hint = &controller.MrsMopWebOutputHint{
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "mrsmop.hintAvailable"
	} else {
		resObj.MessageCode = "mrsmop.noHint"
	}

	return marshalOrError(resObj)
}

func buildMrsMopRunStarts(s interfaces.MrsMopGame) [][]int {
	out := make([][]int, domain.MrsMopTableauCnt)
	for col := range domain.MrsMopTableauCnt {
		out[col] = intSliceOrEmpty(s.RunStarts(col))
	}
	return out
}

// TargetsOutput は Web では通常の盤面をそのまま返す。
//
// 置ける先は legalTargets として盤面とともに返す。Web は選択時にこの値を使う。
func (p *MrsMopWebPresenter) TargetsOutput(s interfaces.MrsMopGame, _, _ int) string {
	return p.Output(s, nil)
}

func buildMrsMopLegalTargets(s interfaces.MrsMopGame) map[string]controller.MrsMopWebOutputTargets {
	out := make(map[string]controller.MrsMopWebOutputTargets)
	for col := 0; col < domain.MrsMopTableauCnt; col++ {
		start, _ := s.LegalTargets(col, -1)
		if start < 0 {
			continue
		}
		for i := start; i < len(s.GetTableau()[col]); i++ {
			_, cols := s.LegalTargets(col, i)
			if len(cols) > 0 {
				out["tableau-"+strconv.Itoa(col)+"-"+strconv.Itoa(i)] = controller.MrsMopWebOutputTargets{Tableau: cols, Foundation: []int{}}
			}
		}
	}
	return out
}

// ActionLogOutput 棋譜をJSON出力
func (p *MrsMopWebPresenter) ActionLogOutput(s interfaces.MrsMopGame) string {
	return actionLogOutputJSON(s)
}
