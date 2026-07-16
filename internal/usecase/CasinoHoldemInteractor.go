//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// CasinoHoldemInteractorIF カジノホールデムインタラクターインタフェース
type CasinoHoldemInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Bet アンテベットとオプションの AA ボーナスサイドベット
	Bet(ante, bonus int) string
	// Call フロップ後にコール（2×アンテ）
	Call() string
	// Fold フロップ後にフォールド
	Fold() string
	// ActionLog 棋譜を出力する
	ActionLog() string
	// Hint ヒントを出力する
	Hint() string
}

// CasinoHoldemInteractor カジノホールデムインタラクタークラス
type CasinoHoldemInteractor struct {
	GameBase[interfaces.CasinoHoldemGame]
	cp presenter.CasinoHoldemPresenter
}

// NewCasinoHoldemInteractor コンストラクタ
func NewCasinoHoldemInteractor(g interfaces.CasinoHoldemGame, cp presenter.CasinoHoldemPresenter) *CasinoHoldemInteractor {
	mustNotNil("CasinoHoldemInteractor", map[string]any{"g": g, "cp": cp})
	return &CasinoHoldemInteractor{
		GameBase: GameBase[interfaces.CasinoHoldemGame]{Game: g},
		cp:       cp,
	}
}

// Reset ゲーム初期化
func (ci *CasinoHoldemInteractor) Reset() string {
	return runAndPresent(ci.Game, ci.cp, ci.Game.Reset)
}

// Bet アンテベット
func (ci *CasinoHoldemInteractor) Bet(ante, bonus int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.Bet(ante, bonus) })
}

// Call フロップ後にコール
func (ci *CasinoHoldemInteractor) Call() string {
	return execAndPresent(ci.Game, ci.cp, ci.Game.Call)
}

// Fold フロップ後にフォールド
func (ci *CasinoHoldemInteractor) Fold() string {
	return execAndPresent(ci.Game, ci.cp, ci.Game.Fold)
}

// ActionLog 棋譜を出力する
func (ci *CasinoHoldemInteractor) ActionLog() string {
	return ci.cp.ActionLogOutput(ci.Game)
}

// Hint ヒントを出力する
func (ci *CasinoHoldemInteractor) Hint() string {
	return ci.cp.HintOutput(ci.Game)
}

// RestoreCasinoHoldemInteractor deserialises JSON into a CasinoHoldemInteractor.
func RestoreCasinoHoldemInteractor(data []byte, cp presenter.CasinoHoldemPresenter) (*CasinoHoldemInteractor, error) {
	return restoreAndBuild[domain.CasinoHoldem](data, func(g *domain.CasinoHoldem) *CasinoHoldemInteractor {
		return &CasinoHoldemInteractor{GameBase: GameBase[interfaces.CasinoHoldemGame]{Game: g}, cp: cp}
	})
}
