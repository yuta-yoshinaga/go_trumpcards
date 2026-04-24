package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// CalculationInteractorIF カルキュレーションインタラクターインタフェース
type CalculationInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// PlayStockToFoundation ストック最上段をファンデーションに置く
	PlayStockToFoundation(fIdx int) string
	// PlayStockToWaste ストック最上段を指定ウェイストパイルに置く
	PlayStockToWaste(wasteIdx int) string
	// PlayWasteToFoundation ウェイスト最上段をファンデーションに置く
	PlayWasteToFoundation(wasteIdx, fIdx int) string
	// GiveUp ギブアップ
	GiveUp() string
	// Undo アンドゥ
	Undo() string
	// UndoN n回連続アンドゥ
	UndoN(n int) string
	// AutoComplete オートコンプリート
	AutoComplete() string
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// CalculationInteractor カルキュレーションインタラクタークラス
type CalculationInteractor struct {
	GameBase[interfaces.CalculationGame]
	solitaireActions[interfaces.CalculationGame]
	p presenter.CalculationPresenter
}

// NewCalculationInteractor コンストラクタ
func NewCalculationInteractor(g interfaces.CalculationGame, p presenter.CalculationPresenter) *CalculationInteractor {
	mustNotNil("CalculationInteractor", map[string]any{"g": g, "p": p})
	return &CalculationInteractor{
		GameBase:         GameBase[interfaces.CalculationGame]{Game: g},
		solitaireActions: newSolitaireActions[interfaces.CalculationGame](g, p),
		p:                p,
	}
}

// Reset ゲーム初期化
func (ci *CalculationInteractor) Reset() string {
	return runAndPresent(ci.Game, ci.p, ci.Game.Reset)
}

// PlayStockToFoundation ストック最上段をファンデーションに置く
func (ci *CalculationInteractor) PlayStockToFoundation(fIdx int) string {
	return execAndPresent(ci.Game, ci.p, func() error { return ci.Game.PlayStockToFoundation(fIdx) })
}

// PlayStockToWaste ストック最上段を指定ウェイストパイルに置く
func (ci *CalculationInteractor) PlayStockToWaste(wasteIdx int) string {
	return execAndPresent(ci.Game, ci.p, func() error { return ci.Game.PlayStockToWaste(wasteIdx) })
}

// PlayWasteToFoundation ウェイスト最上段をファンデーションに置く
func (ci *CalculationInteractor) PlayWasteToFoundation(wasteIdx, fIdx int) string {
	return execAndPresent(ci.Game, ci.p, func() error { return ci.Game.PlayWasteToFoundation(wasteIdx, fIdx) })
}

// Hint ヒント取得
func (ci *CalculationInteractor) Hint() string {
	return ci.p.HintOutput(ci.Game)
}

// ActionLog 棋譜を出力する
func (ci *CalculationInteractor) ActionLog() string {
	return ci.p.ActionLogOutput(ci.Game)
}

// RestoreCalculationInteractor deserialises JSON into a CalculationInteractor.
func RestoreCalculationInteractor(data []byte, p presenter.CalculationPresenter) (*CalculationInteractor, error) {
	return restoreAndBuild[domain.Calculation](data, func(g *domain.Calculation) *CalculationInteractor {
		return &CalculationInteractor{
			GameBase:         GameBase[interfaces.CalculationGame]{Game: g},
			solitaireActions: newSolitaireActions[interfaces.CalculationGame](g, p),
			p:                p,
		}
	})
}
