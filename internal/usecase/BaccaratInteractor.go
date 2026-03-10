package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BaccaratInteractorIF バカラインタラクターインタフェース
type BaccaratInteractorIF interface {
	Reset() string
	Bet(amount, betType int) string
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
func (bi *BaccaratInteractor) Bet(amount, betType int) string {
	err := bi.b.Bet(amount, betType)
	return bi.bp.Output(bi.b, err)
}

// ActionLog 棋譜を出力する
func (bi *BaccaratInteractor) ActionLog() string {
	return bi.bp.ActionLogOutput(bi.b)
}
