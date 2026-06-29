//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SultanInteractorIF スルタンインタラクターインタフェース
type SultanInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Draw ストックからウェイストにカードを引く
	Draw() string
	// Redeal ウェイストを集めて新しいストックを作る
	Redeal() string
	// MoveDivanToFoundation ディヴァンからファンデーションにカードを移動
	MoveDivanToFoundation(divanIdx int) string
	// MoveWasteToFoundation ウェイストからファンデーションにカードを移動
	MoveWasteToFoundation() string
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

// SultanInteractor スルタンインタラクタークラス
type SultanInteractor struct {
	GameBase[interfaces.SultanGame]
	sup presenter.SultanPresenter
	solitaireActions[interfaces.SultanGame]
}

// NewSultanInteractor コンストラクタ
func NewSultanInteractor(su interfaces.SultanGame, sup presenter.SultanPresenter) *SultanInteractor {
	mustNotNil("SultanInteractor", map[string]any{"su": su, "sup": sup})
	return &SultanInteractor{
		GameBase:         GameBase[interfaces.SultanGame]{Game: su},
		sup:              sup,
		solitaireActions: newSolitaireActions[interfaces.SultanGame](su, sup),
	}
}

// Reset ゲーム初期化
func (si *SultanInteractor) Reset() string {
	return runAndPresent(si.Game, si.sup, si.Game.Reset)
}

// Draw ストックからウェイストにカードを引く
func (si *SultanInteractor) Draw() string {
	return execAndPresent(si.Game, si.sup, si.Game.Draw)
}

// Redeal ウェイストを集めて新しいストックを作る
func (si *SultanInteractor) Redeal() string {
	return execAndPresent(si.Game, si.sup, si.Game.Redeal)
}

// MoveDivanToFoundation ディヴァンからファンデーションにカードを移動
func (si *SultanInteractor) MoveDivanToFoundation(divanIdx int) string {
	return execAndPresent(si.Game, si.sup, func() error { return si.Game.MoveDivanToFoundation(divanIdx) })
}

// MoveWasteToFoundation ウェイストからファンデーションにカードを移動
func (si *SultanInteractor) MoveWasteToFoundation() string {
	return execAndPresent(si.Game, si.sup, si.Game.MoveWasteToFoundation)
}

// Hint ヒント取得
func (si *SultanInteractor) Hint() string {
	return si.sup.HintOutput(si.Game)
}

// ActionLog 棋譜を出力する
func (si *SultanInteractor) ActionLog() string {
	return si.sup.ActionLogOutput(si.Game)
}

// RestoreSultanInteractor deserialises JSON into a SultanInteractor.
func RestoreSultanInteractor(data []byte, sup presenter.SultanPresenter) (*SultanInteractor, error) {
	return restoreAndBuild[domain.Sultan](data, func(g *domain.Sultan) *SultanInteractor {
		return &SultanInteractor{
			GameBase:         GameBase[interfaces.SultanGame]{Game: g},
			sup:              sup,
			solitaireActions: newSolitaireActions[interfaces.SultanGame](g, sup),
		}
	})
}
