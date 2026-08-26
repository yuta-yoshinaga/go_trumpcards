//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// WhiteheadInteractorIF ホワイトヘッドインタラクターインタフェース
type WhiteheadInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定付きリセット
	ResetWithConfig(cfg domain.WhiteheadConfig) string
	// Draw ストックからウェイストにカードを引く
	Draw() string
	// MoveWasteToTableau ウェイストからタブローにカードを移動
	MoveWasteToTableau(col int) string
	// MoveWasteToFoundation ウェイストからファンデーションにカードを移動
	MoveWasteToFoundation() string
	// MoveTableauToTableau タブローからタブローにカードを移動
	MoveTableauToTableau(fromCol, cardIndex, toCol int) string
	// MoveTableauToFoundation タブローからファンデーションにカードを移動
	MoveTableauToFoundation(col int) string
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

// WhiteheadInteractor ホワイトヘッドインタラクタークラス
type WhiteheadInteractor struct {
	GameBase[interfaces.WhiteheadGame]
	kp presenter.WhiteheadPresenter
	solitaireActions[interfaces.WhiteheadGame]
}

// NewWhiteheadInteractor コンストラクタ
func NewWhiteheadInteractor(k interfaces.WhiteheadGame, kp presenter.WhiteheadPresenter) *WhiteheadInteractor {
	mustNotNil("WhiteheadInteractor", map[string]any{"k": k, "kp": kp})
	return &WhiteheadInteractor{
		GameBase:         GameBase[interfaces.WhiteheadGame]{Game: k},
		kp:               kp,
		solitaireActions: newSolitaireActions[interfaces.WhiteheadGame](k, kp),
	}
}

// Reset ゲーム初期化
func (ki *WhiteheadInteractor) Reset() string {
	return runAndPresent(ki.Game, ki.kp, ki.Game.Reset)
}

// ResetWithConfig 設定付きリセット
func (ki *WhiteheadInteractor) ResetWithConfig(cfg domain.WhiteheadConfig) string {
	return runAndPresent(ki.Game, ki.kp, func() { ki.Game.ResetWithConfig(cfg) })
}

// Draw ストックからウェイストにカードを引く
func (ki *WhiteheadInteractor) Draw() string {
	return execAndPresent(ki.Game, ki.kp, ki.Game.Draw)
}

// MoveWasteToTableau ウェイストからタブローにカードを移動
func (ki *WhiteheadInteractor) MoveWasteToTableau(col int) string {
	return execAndPresent(ki.Game, ki.kp, func() error { return ki.Game.MoveWasteToTableau(col) })
}

// MoveWasteToFoundation ウェイストからファンデーションにカードを移動
func (ki *WhiteheadInteractor) MoveWasteToFoundation() string {
	return execAndPresent(ki.Game, ki.kp, ki.Game.MoveWasteToFoundation)
}

// MoveTableauToTableau タブローからタブローにカードを移動
func (ki *WhiteheadInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(ki.Game, ki.kp, func() error { return ki.Game.MoveTableauToTableau(fromCol, cardIndex, toCol) })
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (ki *WhiteheadInteractor) MoveTableauToFoundation(col int) string {
	return execAndPresent(ki.Game, ki.kp, func() error { return ki.Game.MoveTableauToFoundation(col) })
}

// Hint ヒント取得
func (ki *WhiteheadInteractor) Hint() string {
	return ki.kp.HintOutput(ki.Game)
}

// ActionLog 棋譜を出力する
func (ki *WhiteheadInteractor) ActionLog() string {
	return ki.kp.ActionLogOutput(ki.Game)
}

// RestoreWhiteheadInteractor deserialises JSON into a WhiteheadInteractor.
func RestoreWhiteheadInteractor(data []byte, kp presenter.WhiteheadPresenter) (*WhiteheadInteractor, error) {
	return restoreAndBuild[domain.Whitehead](data, func(g *domain.Whitehead) *WhiteheadInteractor {
		return &WhiteheadInteractor{
			GameBase:         GameBase[interfaces.WhiteheadGame]{Game: g},
			kp:               kp,
			solitaireActions: newSolitaireActions[interfaces.WhiteheadGame](g, kp),
		}
	})
}
