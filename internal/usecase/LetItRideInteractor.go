//go:build !js || !wasm || extra4

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// LetItRideInteractorIF レット・イット・ライドインタラクターインタフェース
type LetItRideInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Bet ベット
	Bet(amount int) string
	// Pull ベットを取り下げる
	Pull() string
	// PullConfirm Pull 実行前の確認内容を出力する
	PullConfirm() string
	// LetItRide ベットをそのままにする
	LetItRide() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// LetItRideInteractor レット・イット・ライドインタラクタークラス
type LetItRideInteractor struct {
	GameBase[interfaces.LetItRideGame]
	cp presenter.LetItRidePresenter
}

// NewLetItRideInteractor コンストラクタ
func NewLetItRideInteractor(lir interfaces.LetItRideGame, cp presenter.LetItRidePresenter) *LetItRideInteractor {
	mustNotNil("LetItRideInteractor", map[string]any{"lir": lir, "cp": cp})
	return &LetItRideInteractor{
		GameBase: GameBase[interfaces.LetItRideGame]{Game: lir},
		cp:       cp,
	}
}

// Reset ゲーム初期化
func (li *LetItRideInteractor) Reset() string {
	return runAndPresent(li.Game, li.cp, li.Game.Reset)
}

// Bet ベット
func (li *LetItRideInteractor) Bet(amount int) string {
	return execAndPresent(li.Game, li.cp, func() error { return li.Game.Bet(amount) })
}

// Pull ベットを取り下げる
func (li *LetItRideInteractor) Pull() string {
	return execAndPresent(li.Game, li.cp, li.Game.Pull)
}

// PullConfirm Pull 実行前の確認内容を出力する
func (li *LetItRideInteractor) PullConfirm() string {
	return li.cp.PullConfirmOutput(li.Game)
}

// LetItRide ベットをそのままにする
func (li *LetItRideInteractor) LetItRide() string {
	return execAndPresent(li.Game, li.cp, li.Game.LetItRideAction)
}

// ActionLog 棋譜を出力する
func (li *LetItRideInteractor) ActionLog() string {
	return li.cp.ActionLogOutput(li.Game)
}

// RestoreLetItRideInteractor deserialises JSON into a LetItRideInteractor.
func RestoreLetItRideInteractor(data []byte, cp presenter.LetItRidePresenter) (*LetItRideInteractor, error) {
	return restoreAndBuild[domain.LetItRide](data, func(g *domain.LetItRide) *LetItRideInteractor {
		return &LetItRideInteractor{GameBase: GameBase[interfaces.LetItRideGame]{Game: g}, cp: cp}
	})
}
