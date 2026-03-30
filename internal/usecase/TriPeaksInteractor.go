package usecase

import (
	"encoding/json"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// TriPeaksInteractorIF トリピークスインタラクターインタフェース
type TriPeaksInteractorIF interface {
	// Reset ゲーム初期化
	Reset() string
	// Draw ストックからウェイストにカードを引く
	Draw() string
	// Remove タブローのカードを除去
	Remove(row, col int) string
	// GiveUp ギブアップ
	GiveUp() string
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
	// Undo アンドゥ
	Undo() string
}

// TriPeaksInteractor トリピークスインタラクタークラス
type TriPeaksInteractor struct {
	t  interfaces.TriPeaksGame
	tp presenter.TriPeaksPresenter
}

// NewTriPeaksInteractor コンストラクタ
func NewTriPeaksInteractor(t interfaces.TriPeaksGame, tp presenter.TriPeaksPresenter) *TriPeaksInteractor {
	mustNotNil("TriPeaksInteractor", map[string]any{"t": t, "tp": tp})
	return &TriPeaksInteractor{t: t, tp: tp}
}

// Reset ゲーム初期化
func (ti *TriPeaksInteractor) Reset() string {
	return runAndPresent(ti.t, ti.tp, ti.t.Reset)
}

// Draw ストックからウェイストにカードを引く
func (ti *TriPeaksInteractor) Draw() string {
	return execAndPresent(ti.t, ti.tp, ti.t.Draw)
}

// Remove タブローのカードを除去
func (ti *TriPeaksInteractor) Remove(row, col int) string {
	return execAndPresent(ti.t, ti.tp, func() error { return ti.t.Remove(row, col) })
}

// GiveUp ギブアップ
func (ti *TriPeaksInteractor) GiveUp() string {
	return runAndPresent(ti.t, ti.tp, ti.t.GiveUp)
}

// Hint ヒント取得
func (ti *TriPeaksInteractor) Hint() string {
	return ti.tp.HintOutput(ti.t)
}

// ActionLog 棋譜を出力する
func (ti *TriPeaksInteractor) ActionLog() string {
	return ti.tp.ActionLogOutput(ti.t)
}

// Undo アンドゥ
func (ti *TriPeaksInteractor) Undo() string {
	return execAndPresent(ti.t, ti.tp, ti.t.Undo)
}

// Snapshot serialises the game state to JSON for KV persistence.
func (ti *TriPeaksInteractor) Snapshot() ([]byte, error) {
	return json.Marshal(ti.t)
}

// RestoreTriPeaksInteractor deserialises JSON into a TriPeaksInteractor.
func RestoreTriPeaksInteractor(data []byte, tp presenter.TriPeaksPresenter) (*TriPeaksInteractor, error) {
	tripeaks, err := restoreGame[domain.TriPeaks](data)
	if err != nil {
		return nil, err
	}
	return &TriPeaksInteractor{t: tripeaks, tp: tp}, nil
}
