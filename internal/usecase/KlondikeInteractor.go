package usecase

import (
	"encoding/json"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// KlondikeInteractorIF クロンダイクインタラクターインタフェース
type KlondikeInteractorIF interface {
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定付きリセット
	ResetWithConfig(cfg domain.KlondikeConfig) string
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
	return runAndPresent(ki.k, ki.kp, ki.k.Reset)
}

// ResetWithConfig 設定付きリセット
func (ki *KlondikeInteractor) ResetWithConfig(cfg domain.KlondikeConfig) string {
	return runAndPresent(ki.k, ki.kp, func() { ki.k.ResetWithConfig(cfg) })
}

// Draw ストックからウェイストにカードを引く
func (ki *KlondikeInteractor) Draw() string {
	return execAndPresent(ki.k, ki.kp, ki.k.Draw)
}

// MoveWasteToTableau ウェイストからタブローにカードを移動
func (ki *KlondikeInteractor) MoveWasteToTableau(col int) string {
	return execAndPresent(ki.k, ki.kp, func() error { return ki.k.MoveWasteToTableau(col) })
}

// MoveWasteToFoundation ウェイストからファンデーションにカードを移動
func (ki *KlondikeInteractor) MoveWasteToFoundation() string {
	return execAndPresent(ki.k, ki.kp, ki.k.MoveWasteToFoundation)
}

// MoveTableauToTableau タブローからタブローにカードを移動
func (ki *KlondikeInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	return execAndPresent(ki.k, ki.kp, func() error { return ki.k.MoveTableauToTableau(fromCol, cardIndex, toCol) })
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (ki *KlondikeInteractor) MoveTableauToFoundation(col int) string {
	return execAndPresent(ki.k, ki.kp, func() error { return ki.k.MoveTableauToFoundation(col) })
}

// GiveUp ギブアップ
func (ki *KlondikeInteractor) GiveUp() string {
	return runAndPresent(ki.k, ki.kp, ki.k.GiveUp)
}

// Hint ヒント取得
func (ki *KlondikeInteractor) Hint() string {
	return ki.kp.HintOutput(ki.k)
}

// AutoComplete オートコンプリート
func (ki *KlondikeInteractor) AutoComplete() string {
	return execAndPresent(ki.k, ki.kp, ki.k.AutoComplete)
}

// ActionLog 棋譜を出力する
func (ki *KlondikeInteractor) ActionLog() string {
	return ki.kp.ActionLogOutput(ki.k)
}

// Undo アンドゥ
func (ki *KlondikeInteractor) Undo() string {
	return execAndPresent(ki.k, ki.kp, ki.k.Undo)
}

// Snapshot serialises the game state to JSON for KV persistence.
func (ki *KlondikeInteractor) Snapshot() ([]byte, error) {
	return json.Marshal(ki.k)
}

// RestoreKlondikeInteractor deserialises JSON into a KlondikeInteractor.
func RestoreKlondikeInteractor(data []byte, kp presenter.KlondikePresenter) (*KlondikeInteractor, error) {
	kl, err := restoreGame[domain.Klondike](data)
	if err != nil {
		return nil, err
	}
	return &KlondikeInteractor{k: kl, kp: kp}, nil
}
