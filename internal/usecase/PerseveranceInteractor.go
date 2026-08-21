//go:build !js || !wasm || extra4

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// PerseveranceInteractorIF パーシビアランスインタラクターインタフェース
type PerseveranceInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// MoveTableauToTableau タブローからタブローにカードを移動
	MoveTableauToTableau(fromCol, cardIndex, toCol int) string
	// MoveTableauToFoundation タブローからファンデーションにカードを移動
	MoveTableauToFoundation(col int) string
	// Redeal 集めて配り直す (最大2回)
	Redeal() string
	// GiveUp ギブアップ
	GiveUp() string
	// Hint ヒント取得
	Hint() string
	// Targets 列 col の一番下の札を置ける先を一覧する
	Targets(col int) string
	// AutoComplete オートコンプリート
	AutoComplete() string
	// ActionLog 棋譜を出力する
	ActionLog() string
	// Undo アンドゥ
	Undo() string
	// UndoN n回連続アンドゥ
	UndoN(n int) string
}

// PerseveranceInteractor パーシビアランスインタラクタークラス
type PerseveranceInteractor struct {
	GameBase[interfaces.PerseveranceGame]
	bdp presenter.PerseverancePresenter
	solitaireActions[interfaces.PerseveranceGame]
}

// NewPerseveranceInteractor コンストラクタ
func NewPerseveranceInteractor(bd interfaces.PerseveranceGame, bdp presenter.PerseverancePresenter) *PerseveranceInteractor {
	mustNotNil("PerseveranceInteractor", map[string]any{"bd": bd, "bdp": bdp})
	return &PerseveranceInteractor{
		GameBase:         GameBase[interfaces.PerseveranceGame]{Game: bd},
		bdp:              bdp,
		solitaireActions: newSolitaireActions[interfaces.PerseveranceGame](bd, bdp),
	}
}

// Reset ゲーム初期化
func (bi *PerseveranceInteractor) Reset() string {
	return runAndPresent(bi.Game, bi.bdp, bi.Game.Reset)
}

// MoveTableauToTableau タブローからタブローにカードを移動
func (bi *PerseveranceInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(bi.Game, bi.bdp, func() error { return bi.Game.MoveTableauToTableau(fromCol, cardIndex, toCol) })
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (bi *PerseveranceInteractor) MoveTableauToFoundation(col int) string {
	return execAndPresent(bi.Game, bi.bdp, func() error { return bi.Game.MoveTableauToFoundation(col) })
}

// Redeal 集めて配り直す (最大2回)
func (bi *PerseveranceInteractor) Redeal() string {
	return execAndPresent(bi.Game, bi.bdp, bi.Game.Redeal)
}

// Hint ヒント取得
func (bi *PerseveranceInteractor) Hint() string {
	return bi.bdp.HintOutput(bi.Game)
}

// Targets 列 col の一番下の札を置ける先を一覧する
func (bi *PerseveranceInteractor) Targets(col int) string {
	return bi.bdp.TargetsOutput(bi.Game, col)
}

// ActionLog 棋譜を出力する
func (bi *PerseveranceInteractor) ActionLog() string {
	return bi.bdp.ActionLogOutput(bi.Game)
}

// RestorePerseveranceInteractor deserialises JSON into a PerseveranceInteractor.
func RestorePerseveranceInteractor(data []byte, bdp presenter.PerseverancePresenter) (*PerseveranceInteractor, error) {
	return restoreAndBuild[domain.Perseverance](data, func(g *domain.Perseverance) *PerseveranceInteractor {
		return &PerseveranceInteractor{
			GameBase:         GameBase[interfaces.PerseveranceGame]{Game: g},
			bdp:              bdp,
			solitaireActions: newSolitaireActions[interfaces.PerseveranceGame](g, bdp),
		}
	})
}
