//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// MississippiStudInteractorIF ミシシッピ・スタッドインタラクターインタフェース
type MississippiStudInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Bet アンティをベットしカードを配る
	Bet(amount int) string
	// Play 1x/2x/3x のストリートベットを置く
	Play(multiplier int) string
	// Fold フォールド
	Fold() string
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// MississippiStudInteractor ミシシッピ・スタッドインタラクタークラス
type MississippiStudInteractor struct {
	GameBase[interfaces.MississippiStudGame]
	cp presenter.MississippiStudPresenter
}

// NewMississippiStudInteractor コンストラクタ
func NewMississippiStudInteractor(g interfaces.MississippiStudGame, cp presenter.MississippiStudPresenter) *MississippiStudInteractor {
	mustNotNil("MississippiStudInteractor", map[string]any{"g": g, "cp": cp})
	return &MississippiStudInteractor{
		GameBase: GameBase[interfaces.MississippiStudGame]{Game: g},
		cp:       cp,
	}
}

// Reset ゲーム初期化
func (mi *MississippiStudInteractor) Reset() string {
	return runAndPresent(mi.Game, mi.cp, mi.Game.Reset)
}

// Bet アンティをベットしカードを配る
func (mi *MississippiStudInteractor) Bet(amount int) string {
	return execAndPresent(mi.Game, mi.cp, func() error { return mi.Game.Bet(amount) })
}

// Play ストリートベットを置く
func (mi *MississippiStudInteractor) Play(multiplier int) string {
	return execAndPresent(mi.Game, mi.cp, func() error { return mi.Game.Play(multiplier) })
}

// Fold フォールド
func (mi *MississippiStudInteractor) Fold() string {
	return execAndPresent(mi.Game, mi.cp, mi.Game.Fold)
}

// Hint ヒント取得
func (mi *MississippiStudInteractor) Hint() string {
	return mi.cp.HintOutput(mi.Game)
}

// ActionLog 棋譜を出力する
func (mi *MississippiStudInteractor) ActionLog() string {
	return mi.cp.ActionLogOutput(mi.Game)
}

// RestoreMississippiStudInteractor deserialises JSON into a MississippiStudInteractor.
func RestoreMississippiStudInteractor(data []byte, cp presenter.MississippiStudPresenter) (*MississippiStudInteractor, error) {
	return restoreAndBuild[domain.MississippiStud](data, func(g *domain.MississippiStud) *MississippiStudInteractor {
		return &MississippiStudInteractor{GameBase: GameBase[interfaces.MississippiStudGame]{Game: g}, cp: cp}
	})
}
