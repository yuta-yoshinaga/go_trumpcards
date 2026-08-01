//go:build !js || !wasm || solo

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SeahavenTowersWebPresenter シーヘイブンタワーズWebプレゼンタークラス
type SeahavenTowersWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *SeahavenTowersWebPresenter) Output(s interfaces.SeahavenTowersGame, lastErr error) string {
	resObj := new(controller.SeahavenTowersWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, s, int(s.GetPhase()))

	// タブロー
	tableau := s.GetTableau()
	resObj.Tableau = make([][]*controller.WebOutputCard, domain.SeahavenTowersTableauCnt)
	for i := 0; i < domain.SeahavenTowersTableauCnt; i++ {
		colCards := tableau[i]
		resObj.Tableau[i] = make([]*controller.WebOutputCard, len(colCards))
		for j, c := range colCards {
			resObj.Tableau[i][j] = cardToOutput(c)
		}
	}

	// リザーブセル
	cells := s.GetFreeCells()
	resObj.ReservedCells = make([]*controller.WebOutputCard, domain.SeahavenTowersCellCnt)
	for i := 0; i < domain.SeahavenTowersCellCnt; i++ {
		resObj.ReservedCells[i] = cardToOutput(cells[i])
	}

	// ファンデーション
	foundation := s.GetFoundation()
	resObj.Foundation = make([][]*controller.WebOutputCard, domain.SeahavenTowersFoundationCnt)
	for i := 0; i < domain.SeahavenTowersFoundationCnt; i++ {
		pile := foundation[i]
		resObj.Foundation[i] = make([]*controller.WebOutputCard, len(pile))
		for j, c := range pile {
			resObj.Foundation[i][j] = cardToOutput(c)
		}
	}

	// メッセージ
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	if s.GetPhase() == domain.SeahavenTowersPhasePlaying && !s.IsStalemate() {
		if hint := s.GetHint(); hint != nil {
			resObj.Hint = &controller.SeahavenTowersWebOutputHint{
				FromZone:  hint.FromZone,
				FromCol:   hint.FromCol,
				CardIndex: hint.CardIndex,
				ToZone:    hint.ToZone,
				ToCol:     hint.ToCol,
			}
		}
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		phase := s.GetPhase()
		switch phase {
		case domain.SeahavenTowersPhasePlaying:
			if s.IsStalemate() {
				if resObj.UndoToEscape > 0 {
					resObj.MessageCode = "seahaventowers.stalemateWithEscape"
					resObj.MessageParams = map[string]string{"count": fmt.Sprintf("%d", resObj.UndoToEscape)}
				} else {
					resObj.MessageCode = "seahaventowers.stalemate"
				}
			} else {
				resObj.MessageCode = "seahaventowers.playing"
			}
		case domain.SeahavenTowersPhaseGameClear:
			resObj.MessageCode = "seahaventowers.gameClear"
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", s.GetMoveCount())}
		case domain.SeahavenTowersPhaseGameOver:
			resObj.MessageCode = "seahaventowers.gameOver"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *SeahavenTowersWebPresenter) HintOutput(s interfaces.SeahavenTowersGame) string {
	hint := s.GetHint()
	resObj := new(controller.SeahavenTowersWebOutput)
	populateSolitaireBase(&resObj.SolitaireWebOutputBase, s, int(s.GetPhase()))
	resObj.ReservedCells = make([]*controller.WebOutputCard, 0)
	resObj.Tableau = make([][]*controller.WebOutputCard, 0)
	resObj.Foundation = make([][]*controller.WebOutputCard, 0)

	if hint != nil {
		resObj.Hint = &controller.SeahavenTowersWebOutputHint{
			FromZone:  hint.FromZone,
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToZone:    hint.ToZone,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "seahaventowers.hintAvailable"
	} else {
		resObj.MessageCode = "seahaventowers.noHint"
	}

	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *SeahavenTowersWebPresenter) ActionLogOutput(s interfaces.SeahavenTowersGame) string {
	return actionLogOutputJSON(s)
}
