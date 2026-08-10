//go:build !js || !wasm || extra2

package presenter

import (
	"encoding/json"
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CribbageSquaresWebPresenter はクリベッジ・スクエアズの Web プレゼンター。
type CribbageSquaresWebPresenter struct{}

// Output はゲーム状態を JSON で出力する。
func (pr *CribbageSquaresWebPresenter) Output(p interfaces.CribbageSquaresGame, lastErr error) string {
	resObj := new(controller.CribbageSquaresWebOutput)
	resObj.Phase = int(p.GetPhase())
	resObj.PlacedCount = p.GetPlacedCount()
	resObj.CanUndo = p.CanUndo()
	resObj.CurrentCard = cardToOutput(p.GetCurrentCard())

	board := p.GetBoard()
	resObj.Board = make([][]*controller.CribbageSquaresWebOutputCard, domain.CribbageSquaresGridSize)
	for r := range domain.CribbageSquaresGridSize {
		resObj.Board[r] = make([]*controller.CribbageSquaresWebOutputCard, domain.CribbageSquaresGridSize)
		for c := range domain.CribbageSquaresGridSize {
			cell := &controller.CribbageSquaresWebOutputCard{}
			if board[r][c] != nil {
				cell.Card = cardToOutput(board[r][c])
			}
			resObj.Board[r][c] = cell
		}
	}

	resObj.Starter = cardToOutput(p.GetStarter())
	resObj.WinScore = domain.CribbageSquaresWinScore

	resObj.RowScores = make([]int, domain.CribbageSquaresGridSize)
	resObj.ColScores = make([]int, domain.CribbageSquaresGridSize)
	resObj.RowDetails = make([]*controller.CribbageSquaresWebOutputScore, domain.CribbageSquaresGridSize)
	resObj.ColDetails = make([]*controller.CribbageSquaresWebOutputScore, domain.CribbageSquaresGridSize)
	for i := range domain.CribbageSquaresGridSize {
		resObj.RowDetails[i] = cribbageSquaresScoreOutput(p.RowDetail(i))
		resObj.ColDetails[i] = cribbageSquaresScoreOutput(p.ColDetail(i))
		resObj.RowScores[i] = resObj.RowDetails[i].Total
		resObj.ColScores[i] = resObj.ColDetails[i].Total
		resObj.TotalScore += resObj.RowScores[i] + resObj.ColScores[i]
	}
	resObj.IsWin = p.IsWin()

	// **受動ヒントは Output() でも埋める。**command:"hint" 専用のレスポンスは
	// ページの state にマージされないので、ここで埋めないと state.hint が
	// 永久に undefined になる (#4483)。
	resObj.Hint = cribbageSquaresWebHint(p)

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		switch p.GetPhase() {
		case domain.CribbageSquaresPhasePlaying:
			resObj.MessageCode = "cribbagesquares.playing"
		case domain.CribbageSquaresPhaseComplete:
			resObj.Message = fmt.Sprintf("ゲーム終了 合計得点: %d", resObj.TotalScore)
			// 61 点に届いたかで文言を分ける。合計だけ出しても、それが良いのか
			// 悪いのか初見では分からない。
			resObj.MessageCode = "cribbagesquares.lose"
			if resObj.IsWin {
				resObj.MessageCode = "cribbagesquares.win"
			}
			resObj.MessageParams = map[string]string{
				"totalScore": fmt.Sprintf("%d", resObj.TotalScore),
				"winScore":   fmt.Sprintf("%d", domain.CribbageSquaresWinScore),
			}
		}
	}

	return marshalOrError(resObj)
}

// HintOutput はヒントを JSON で出力する。Web GUI はクライアント側の
// ヒューリスティックヒントを使うため、ここでは JSON 化した状態を返して
// CribbageSquaresPresenter インタフェースを満たす (CUI のみが利用する)。
func (pr *CribbageSquaresWebPresenter) HintOutput(p interfaces.CribbageSquaresGame) string {
	out := pr.Output(p, nil)
	// **押したときだけ出す印を付ける。**ページは isRequestedHint で
	// 「今まさに要求されたヒント」かを見分ける。Output() が常に載せる hint と
	// 区別が付かないと、ボタンを押していないのに助言が出続ける。
	var resObj controller.CribbageSquaresWebOutput
	if err := json.Unmarshal([]byte(out), &resObj); err != nil {
		return out
	}
	if resObj.Hint != nil {
		resObj.MessageCode = "cribbagesquares.hintAvailable"
	} else {
		resObj.MessageCode = "cribbagesquares.noHint"
	}
	return marshalOrError(&resObj)
}

// cribbageSquaresScoreOutput はクリベッジの得点内訳を JSON 用の形に移す。
func cribbageSquaresScoreOutput(d domain.CribbageScoreDetail) *controller.CribbageSquaresWebOutputScore {
	return &controller.CribbageSquaresWebOutputScore{
		Fifteens: d.Fifteens,
		Pairs:    d.Pairs,
		Runs:     d.Runs,
		Flush:    d.Flush,
		Nobs:     d.Nobs,
		Total:    d.Total,
	}
}

// cribbageSquaresWebHint はドメインのヒントを JSON 用の形に移す。
func cribbageSquaresWebHint(p interfaces.CribbageSquaresGame) *controller.CribbageSquaresWebOutputHint {
	hint := p.GetHint()
	if hint == nil {
		return nil
	}
	return &controller.CribbageSquaresWebOutputHint{
		Row: hint.Row, Col: hint.Col, Score: hint.Score, Synergy: hint.Synergy,
	}
}

// ActionLogOutput は棋譜を JSON で出力する。
func (pr *CribbageSquaresWebPresenter) ActionLogOutput(p interfaces.CribbageSquaresGame) string {
	return actionLogOutputJSON(p)
}
