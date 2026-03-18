package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// KlondikeInteractorIF クロンダイクインタラクターインタフェース
type KlondikeInteractorIF interface {
	Reset() string
	ResetWithConfig(cfg domain.KlondikeConfig) string
	Draw() string
	MoveWasteToTableau(col int) string
	MoveWasteToFoundation() string
	MoveTableauToTableau(fromCol, cardIndex, toCol int) string
	MoveTableauToFoundation(col int) string
	GiveUp() string
	Hint() string
	AutoComplete() string
	ActionLog() string
	Undo() string
}

// KlondikeInteractor クロンダイクインタラクタークラス
type KlondikeInteractor struct {
	k  interfaces.KlondikeGame
	kp presenter.KlondikePresenter
}

// NewKlondikeInteractor コンストラクタ
func NewKlondikeInteractor(k interfaces.KlondikeGame, kp presenter.KlondikePresenter) *KlondikeInteractor {
	mustNotNil("KlondikeInteractor", map[string]any{"k": k, "kp": kp})
	return &KlondikeInteractor{k: k, kp: kp}
}

// Reset ゲーム初期化
func (ki *KlondikeInteractor) Reset() string {
	ki.k.Reset()
	return ki.kp.Output(ki.k, nil)
}

// ResetWithConfig 設定付きリセット
func (ki *KlondikeInteractor) ResetWithConfig(cfg domain.KlondikeConfig) string {
	ki.k.ResetWithConfig(cfg)
	return ki.kp.Output(ki.k, nil)
}

// Draw ストックからウェイストにカードを引く
func (ki *KlondikeInteractor) Draw() string {
	err := ki.k.Draw()
	return ki.kp.Output(ki.k, err)
}

// MoveWasteToTableau ウェイストからタブローにカードを移動
func (ki *KlondikeInteractor) MoveWasteToTableau(col int) string {
	err := ki.k.MoveWasteToTableau(col)
	return ki.kp.Output(ki.k, err)
}

// MoveWasteToFoundation ウェイストからファンデーションにカードを移動
func (ki *KlondikeInteractor) MoveWasteToFoundation() string {
	err := ki.k.MoveWasteToFoundation()
	return ki.kp.Output(ki.k, err)
}

// MoveTableauToTableau タブローからタブローにカードを移動
func (ki *KlondikeInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	err := ki.k.MoveTableauToTableau(fromCol, cardIndex, toCol)
	return ki.kp.Output(ki.k, err)
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (ki *KlondikeInteractor) MoveTableauToFoundation(col int) string {
	err := ki.k.MoveTableauToFoundation(col)
	return ki.kp.Output(ki.k, err)
}

// GiveUp ギブアップ
func (ki *KlondikeInteractor) GiveUp() string {
	ki.k.GiveUp()
	return ki.kp.Output(ki.k, nil)
}

// Hint ヒント取得
func (ki *KlondikeInteractor) Hint() string {
	return ki.kp.HintOutput(ki.k)
}

// AutoComplete オートコンプリート
func (ki *KlondikeInteractor) AutoComplete() string {
	err := ki.k.AutoComplete()
	return ki.kp.Output(ki.k, err)
}

// ActionLog 棋譜を出力する
func (ki *KlondikeInteractor) ActionLog() string {
	return ki.kp.ActionLogOutput(ki.k)
}

// Undo アンドゥ
func (ki *KlondikeInteractor) Undo() string {
	err := ki.k.Undo()
	return ki.kp.Output(ki.k, err)
}
