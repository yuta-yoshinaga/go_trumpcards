//go:build !js || !wasm || extra5

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// OichoKabuInteractorIF おいちょかぶインタラクターインタフェース
type OichoKabuInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Bet 掛け金を置き 2 枚ずつ配る
	Bet(amount int) string
	// Draw 子が 3 枚目を引き勝負する
	Draw() string
	// Stand 子が引かずに勝負する
	Stand() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// OichoKabuInteractor おいちょかぶインタラクタークラス
type OichoKabuInteractor struct {
	GameBase[interfaces.OichoKabuGame]
	cp presenter.OichoKabuPresenter
}

// NewOichoKabuInteractor コンストラクタ
func NewOichoKabuInteractor(og interfaces.OichoKabuGame, cp presenter.OichoKabuPresenter) *OichoKabuInteractor {
	mustNotNil("OichoKabuInteractor", map[string]any{"og": og, "cp": cp})
	return &OichoKabuInteractor{
		GameBase: GameBase[interfaces.OichoKabuGame]{Game: og},
		cp:       cp,
	}
}

// Reset ゲーム初期化
func (oi *OichoKabuInteractor) Reset() string {
	return runAndPresent(oi.Game, oi.cp, oi.Game.Reset)
}

// Bet 掛け金を置き 2 枚ずつ配る
func (oi *OichoKabuInteractor) Bet(amount int) string {
	return execAndPresent(oi.Game, oi.cp, func() error {
		return oi.Game.Bet(amount)
	})
}

// Draw 子が 3 枚目を引き勝負する
func (oi *OichoKabuInteractor) Draw() string {
	return execAndPresent(oi.Game, oi.cp, oi.Game.Draw)
}

// Stand 子が引かずに勝負する
func (oi *OichoKabuInteractor) Stand() string {
	return execAndPresent(oi.Game, oi.cp, oi.Game.Stand)
}

// ActionLog 棋譜を出力する
func (oi *OichoKabuInteractor) ActionLog() string {
	return oi.cp.ActionLogOutput(oi.Game)
}

// RestoreOichoKabuInteractor deserialises JSON into an OichoKabuInteractor.
func RestoreOichoKabuInteractor(data []byte, cp presenter.OichoKabuPresenter) (*OichoKabuInteractor, error) {
	return restoreAndBuild[domain.OichoKabu](data, func(g *domain.OichoKabu) *OichoKabuInteractor {
		return &OichoKabuInteractor{GameBase: GameBase[interfaces.OichoKabuGame]{Game: g}, cp: cp}
	})
}
