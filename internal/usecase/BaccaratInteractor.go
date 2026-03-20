package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BaccaratInteractorIF バカラインタラクターインタフェース
type BaccaratInteractorIF interface {
	// Reset ゲーム初期化
	Reset() string
	// Bet ベット
	Bet(amount, betType, ppBet, bpBet int) string
	// ClearHistory 罫線履歴クリア
	ClearHistory() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// BaccaratInteractor バカラインタラクタークラス
type BaccaratInteractor struct {
	b  interfaces.BaccaratGame
	bp presenter.BaccaratPresenter
}

// NewBaccaratInteractor コンストラクタ
func NewBaccaratInteractor(b interfaces.BaccaratGame, bp presenter.BaccaratPresenter) *BaccaratInteractor {
	mustNotNil("BaccaratInteractor", map[string]any{"b": b, "bp": bp})
	return &BaccaratInteractor{
		b:  b,
		bp: bp,
	}
}

// Reset ゲーム初期化
func (bi *BaccaratInteractor) Reset() string {
	bi.b.Reset()
	return bi.bp.Output(bi.b, nil)
}

// Bet ベット
func (bi *BaccaratInteractor) Bet(amount, betType, ppBet, bpBet int) string {
	err := bi.b.Bet(amount, betType, ppBet, bpBet)
	return bi.bp.Output(bi.b, err)
}

// ClearHistory 罫線履歴クリア
func (bi *BaccaratInteractor) ClearHistory() string {
	bi.b.ClearHistory()
	return bi.bp.Output(bi.b, nil)
}

// ActionLog 棋譜を出力する
func (bi *BaccaratInteractor) ActionLog() string {
	return bi.bp.ActionLogOutput(bi.b)
}
