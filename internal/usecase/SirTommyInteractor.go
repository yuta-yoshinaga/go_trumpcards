//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SirTommyInteractorIF サー・トミーインタラクターインタフェース
type SirTommyInteractorIF interface {
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

// SirTommyInteractor サー・トミーインタラクタークラス
type SirTommyInteractor struct {
	GameBase[interfaces.SirTommyGame]
	solitaireActions[interfaces.SirTommyGame]
	p presenter.SirTommyPresenter
}

// NewSirTommyInteractor コンストラクタ
func NewSirTommyInteractor(g interfaces.SirTommyGame, p presenter.SirTommyPresenter) *SirTommyInteractor {
	mustNotNil("SirTommyInteractor", map[string]any{"g": g, "p": p})
	return &SirTommyInteractor{
		GameBase:         GameBase[interfaces.SirTommyGame]{Game: g},
		solitaireActions: newSolitaireActions[interfaces.SirTommyGame](g, p),
		p:                p,
	}
}

// Reset ゲーム初期化
func (si *SirTommyInteractor) Reset() string {
	return runAndPresent(si.Game, si.p, si.Game.Reset)
}

// PlayStockToFoundation ストック最上段をファンデーションに置く
func (si *SirTommyInteractor) PlayStockToFoundation(fIdx int) string {
	return execAndPresent(si.Game, si.p, func() error { return si.Game.PlayStockToFoundation(fIdx) })
}

// PlayStockToWaste ストック最上段を指定ウェイストパイルに置く
func (si *SirTommyInteractor) PlayStockToWaste(wasteIdx int) string {
	return execAndPresent(si.Game, si.p, func() error { return si.Game.PlayStockToWaste(wasteIdx) })
}

// PlayWasteToFoundation ウェイスト最上段をファンデーションに置く
func (si *SirTommyInteractor) PlayWasteToFoundation(wasteIdx, fIdx int) string {
	return execAndPresent(si.Game, si.p, func() error { return si.Game.PlayWasteToFoundation(wasteIdx, fIdx) })
}

// Hint ヒント取得
func (si *SirTommyInteractor) Hint() string {
	return si.p.HintOutput(si.Game)
}

// ActionLog 棋譜を出力する
func (si *SirTommyInteractor) ActionLog() string {
	return si.p.ActionLogOutput(si.Game)
}

// RestoreSirTommyInteractor deserialises JSON into a SirTommyInteractor.
func RestoreSirTommyInteractor(data []byte, p presenter.SirTommyPresenter) (*SirTommyInteractor, error) {
	return restoreAndBuild[domain.SirTommy](data, func(g *domain.SirTommy) *SirTommyInteractor {
		return &SirTommyInteractor{
			GameBase:         GameBase[interfaces.SirTommyGame]{Game: g},
			solitaireActions: newSolitaireActions[interfaces.SirTommyGame](g, p),
			p:                p,
		}
	})
}
