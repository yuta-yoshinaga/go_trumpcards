package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// TrashInteractorIF トラッシュインタラクターインタフェース
type TrashInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Draw 山札から1枚引いて連鎖を解決する
	Draw() string
	// PlaceWild ワイルドを指定位置に配置する
	PlaceWild(pos int) string
	// CpuStep CPUのターンを1ステップ進める
	CpuStep() string
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// TrashInteractor トラッシュインタラクタークラス
type TrashInteractor struct {
	GameBase[interfaces.TrashGame]
	tp presenter.TrashPresenter
}

// NewTrashInteractor コンストラクタ
func NewTrashInteractor(t interfaces.TrashGame, tp presenter.TrashPresenter) *TrashInteractor {
	mustNotNil("TrashInteractor", map[string]any{"t": t, "tp": tp})
	return &TrashInteractor{GameBase: GameBase[interfaces.TrashGame]{Game: t}, tp: tp}
}

// Reset ゲーム初期化
func (ti *TrashInteractor) Reset() string {
	return runAndPresent(ti.Game, ti.tp, ti.Game.Reset)
}

// Draw 山札から1枚引く
func (ti *TrashInteractor) Draw() string {
	return execAndPresent(ti.Game, ti.tp, ti.Game.Draw)
}

// PlaceWild ワイルドを指定位置に配置
func (ti *TrashInteractor) PlaceWild(pos int) string {
	return execAndPresent(ti.Game, ti.tp, func() error { return ti.Game.PlaceWild(pos) })
}

// CpuStep CPUターンを1ステップ進める
func (ti *TrashInteractor) CpuStep() string {
	return execAndPresent(ti.Game, ti.tp, ti.Game.CpuStep)
}

// Hint ヒント取得
func (ti *TrashInteractor) Hint() string {
	return ti.tp.HintOutput(ti.Game)
}

// ActionLog 棋譜を出力する
func (ti *TrashInteractor) ActionLog() string {
	return ti.tp.ActionLogOutput(ti.Game)
}

// RestoreTrashInteractor deserialises JSON into a TrashInteractor.
func RestoreTrashInteractor(data []byte, tp presenter.TrashPresenter) (*TrashInteractor, error) {
	return restoreAndBuild[domain.Trash](data, func(g *domain.Trash) *TrashInteractor {
		return &TrashInteractor{GameBase: GameBase[interfaces.TrashGame]{Game: g}, tp: tp}
	})
}
