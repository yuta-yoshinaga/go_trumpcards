package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
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
	GameBase[interfaces.BaccaratGame]
	bp presenter.BaccaratPresenter
}

// NewBaccaratInteractor コンストラクタ
func NewBaccaratInteractor(b interfaces.BaccaratGame, bp presenter.BaccaratPresenter) *BaccaratInteractor {
	mustNotNil("BaccaratInteractor", map[string]any{"b": b, "bp": bp})
	return &BaccaratInteractor{
		GameBase: GameBase[interfaces.BaccaratGame]{Game: b},
		bp:       bp,
	}
}

// Reset ゲーム初期化
func (bi *BaccaratInteractor) Reset() string {
	return runAndPresent(bi.Game, bi.bp, bi.Game.Reset)
}

// Bet ベット
func (bi *BaccaratInteractor) Bet(amount, betType, ppBet, bpBet int) string {
	return execAndPresent(bi.Game, bi.bp, func() error { return bi.Game.Bet(amount, betType, ppBet, bpBet) })
}

// ClearHistory 罫線履歴クリア
func (bi *BaccaratInteractor) ClearHistory() string {
	return runAndPresent(bi.Game, bi.bp, bi.Game.ClearHistory)
}

// ActionLog 棋譜を出力する
func (bi *BaccaratInteractor) ActionLog() string {
	return bi.bp.ActionLogOutput(bi.Game)
}

// RestoreBaccaratInteractor deserialises JSON into a BaccaratInteractor.
func RestoreBaccaratInteractor(data []byte, bp presenter.BaccaratPresenter) (*BaccaratInteractor, error) {
	return restoreAndBuild[domain.Baccarat](data, func(g *domain.Baccarat) *BaccaratInteractor {
		return &BaccaratInteractor{GameBase: GameBase[interfaces.BaccaratGame]{Game: g}, bp: bp}
	})
}
