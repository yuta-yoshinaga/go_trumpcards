//go:build !js || !wasm || extra4

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// MrsMopInteractorIF ミセス・モップソリティアインタラクターインタフェース
type MrsMopInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定付きリセット
	ResetWithConfig(cfg domain.MrsMopConfig) string
	// MoveTableauToTableau タブロー間でカードを移動
	MoveTableauToTableau(fromCol, cardIndex, toCol int) string
	// GiveUp ギブアップ
	GiveUp() string
	// Hint ヒント取得
	Hint() string
	// AutoComplete オートコンプリート
	AutoComplete() string
	// ActionLog 棋譜を出力する
	ActionLog() string
	// Undo アンドゥ
	Undo() string
	// UndoN n回連続アンドゥ
	UndoN(n int) string
}

// MrsMopInteractor ミセス・モップソリティアインタラクタークラス
type MrsMopInteractor struct {
	GameBase[interfaces.MrsMopGame]
	sp presenter.MrsMopPresenter
	solitaireActions[interfaces.MrsMopGame]
}

// NewMrsMopInteractor コンストラクタ
func NewMrsMopInteractor(s interfaces.MrsMopGame, sp presenter.MrsMopPresenter) *MrsMopInteractor {
	mustNotNil("MrsMopInteractor", map[string]any{"s": s, "sp": sp})
	return &MrsMopInteractor{
		GameBase:         GameBase[interfaces.MrsMopGame]{Game: s},
		sp:               sp,
		solitaireActions: newSolitaireActions[interfaces.MrsMopGame](s, sp),
	}
}

// Reset ゲーム初期化
func (si *MrsMopInteractor) Reset() string {
	return runAndPresent(si.Game, si.sp, si.Game.Reset)
}

// ResetWithConfig 設定付きリセット
func (si *MrsMopInteractor) ResetWithConfig(cfg domain.MrsMopConfig) string {
	return runAndPresent(si.Game, si.sp, func() { si.Game.ResetWithConfig(cfg) })
}

// MoveTableauToTableau タブロー間でカードを移動
func (si *MrsMopInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(si.Game, si.sp, func() error { return si.Game.MoveTableauToTableau(fromCol, cardIndex, toCol) })
}

// Hint ヒント取得
func (si *MrsMopInteractor) Hint() string {
	return si.sp.HintOutput(si.Game)
}

// ActionLog 棋譜を出力する
func (si *MrsMopInteractor) ActionLog() string {
	return si.sp.ActionLogOutput(si.Game)
}

// RestoreMrsMopInteractor deserialises JSON into a MrsMopInteractor.
func RestoreMrsMopInteractor(data []byte, sp presenter.MrsMopPresenter) (*MrsMopInteractor, error) {
	return restoreAndBuild[domain.MrsMop](data, func(g *domain.MrsMop) *MrsMopInteractor {
		return &MrsMopInteractor{
			GameBase:         GameBase[interfaces.MrsMopGame]{Game: g},
			sp:               sp,
			solitaireActions: newSolitaireActions[interfaces.MrsMopGame](g, sp),
		}
	})
}
