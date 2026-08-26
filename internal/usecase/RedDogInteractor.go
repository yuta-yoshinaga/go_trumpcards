//go:build !js || !wasm || extra4

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// RedDogInteractorIF レッドドッグインタラクターインタフェース
type RedDogInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Bet アンテをベット
	Bet(amount int) string
	// Raise レイズして3枚目を引く
	Raise(amount int) string
	// Stay レイズせず3枚目を引く
	Stay() string
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// RedDogInteractor レッドドッグインタラクタークラス
type RedDogInteractor struct {
	GameBase[interfaces.RedDogGame]
	cp presenter.RedDogPresenter
}

// NewRedDogInteractor コンストラクタ
func NewRedDogInteractor(rd interfaces.RedDogGame, cp presenter.RedDogPresenter) *RedDogInteractor {
	mustNotNil("RedDogInteractor", map[string]any{"rd": rd, "cp": cp})
	return &RedDogInteractor{
		GameBase: GameBase[interfaces.RedDogGame]{Game: rd},
		cp:       cp,
	}
}

// Reset ゲーム初期化
func (ri *RedDogInteractor) Reset() string {
	return runAndPresent(ri.Game, ri.cp, ri.Game.Reset)
}

// Bet アンテをベットしカードを配る。Bet 直後に ResolveInitial を呼んで初手結果まで進める。
func (ri *RedDogInteractor) Bet(amount int) string {
	return execAndPresent(ri.Game, ri.cp, func() error {
		if err := ri.Game.Bet(amount); err != nil {
			return err
		}
		ri.Game.ResolveInitial()
		return nil
	})
}

// Raise レイズして3枚目を引く
func (ri *RedDogInteractor) Raise(amount int) string {
	return execAndPresent(ri.Game, ri.cp, func() error { return ri.Game.Raise(amount) })
}

// Stay レイズせず3枚目を引く
func (ri *RedDogInteractor) Stay() string {
	return execAndPresent(ri.Game, ri.cp, ri.Game.Stay)
}

// Hint ヒント取得
func (ri *RedDogInteractor) Hint() string {
	return ri.cp.HintOutput(ri.Game)
}

// ActionLog 棋譜を出力する
func (ri *RedDogInteractor) ActionLog() string {
	return ri.cp.ActionLogOutput(ri.Game)
}

// RestoreRedDogInteractor deserialises JSON into a RedDogInteractor.
func RestoreRedDogInteractor(data []byte, cp presenter.RedDogPresenter) (*RedDogInteractor, error) {
	return restoreAndBuild[domain.RedDog](data, func(g *domain.RedDog) *RedDogInteractor {
		return &RedDogInteractor{GameBase: GameBase[interfaces.RedDogGame]{Game: g}, cp: cp}
	})
}
