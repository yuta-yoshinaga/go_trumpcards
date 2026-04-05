package usecase

import (
	"encoding/json"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// GolfInteractorIF ゴルフソリティアインタラクターインタフェース
type GolfInteractorIF interface {
	// Reset ゲーム初期化
	Reset() string
	// Draw ストックからウェイストにカードを引く
	Draw() string
	// Remove タブローのカードを除去
	Remove(col int) string
	// GiveUp ギブアップ
	GiveUp() string
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
	// Undo アンドゥ
	Undo() string
	// UndoN n回連続アンドゥ
	UndoN(n int) string
}

// GolfInteractor ゴルフソリティアインタラクタークラス
type GolfInteractor struct {
	g  interfaces.GolfGame
	gp presenter.GolfPresenter
}

// NewGolfInteractor コンストラクタ
func NewGolfInteractor(g interfaces.GolfGame, gp presenter.GolfPresenter) *GolfInteractor {
	mustNotNil("GolfInteractor", map[string]any{"g": g, "gp": gp})
	return &GolfInteractor{g: g, gp: gp}
}

// Reset ゲーム初期化
func (gi *GolfInteractor) Reset() string {
	return runAndPresent(gi.g, gi.gp, gi.g.Reset)
}

// Draw ストックからウェイストにカードを引く
func (gi *GolfInteractor) Draw() string {
	return execAndPresent(gi.g, gi.gp, gi.g.Draw)
}

// Remove タブローのカードを除去
func (gi *GolfInteractor) Remove(col int) string {
	return execAndPresent(gi.g, gi.gp, func() error { return gi.g.Remove(col) })
}

// GiveUp ギブアップ
func (gi *GolfInteractor) GiveUp() string {
	return runAndPresent(gi.g, gi.gp, gi.g.GiveUp)
}

// Hint ヒント取得
func (gi *GolfInteractor) Hint() string {
	return gi.gp.HintOutput(gi.g)
}

// ActionLog 棋譜を出力する
func (gi *GolfInteractor) ActionLog() string {
	return gi.gp.ActionLogOutput(gi.g)
}

// Undo アンドゥ
func (gi *GolfInteractor) Undo() string {
	return execAndPresent(gi.g, gi.gp, gi.g.Undo)
}

// UndoN n回連続アンドゥ
func (gi *GolfInteractor) UndoN(n int) string {
	return execAndPresent(gi.g, gi.gp, func() error { return gi.g.UndoN(n) })
}

// Snapshot serialises the game state to JSON for KV persistence.
func (gi *GolfInteractor) Snapshot() ([]byte, error) {
	return json.Marshal(gi.g)
}

// RestoreGolfInteractor deserialises JSON into a GolfInteractor.
func RestoreGolfInteractor(data []byte, gp presenter.GolfPresenter) (*GolfInteractor, error) {
	golf, err := restoreGame[domain.Golf](data)
	if err != nil {
		return nil, err
	}
	return &GolfInteractor{g: golf, gp: gp}, nil
}
