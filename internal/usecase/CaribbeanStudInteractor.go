//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// CaribbeanStudInteractorIF カリビアンスタッドポーカーインタラクターインタフェース
type CaribbeanStudInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Bet アンテベット
	Bet(ante, jackpot int) string
	// Play プレイ（コール）
	Play() string
	// Fold フォールド
	Fold() string
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// CaribbeanStudInteractor カリビアンスタッドポーカーインタラクタークラス
type CaribbeanStudInteractor struct {
	GameBase[interfaces.CaribbeanStudGame]
	cp presenter.CaribbeanStudPresenter
}

// NewCaribbeanStudInteractor コンストラクタ
func NewCaribbeanStudInteractor(cs interfaces.CaribbeanStudGame, cp presenter.CaribbeanStudPresenter) *CaribbeanStudInteractor {
	mustNotNil("CaribbeanStudInteractor", map[string]any{"cs": cs, "cp": cp})
	return &CaribbeanStudInteractor{
		GameBase: GameBase[interfaces.CaribbeanStudGame]{Game: cs},
		cp:       cp,
	}
}

// Reset ゲーム初期化
func (ci *CaribbeanStudInteractor) Reset() string {
	return runAndPresent(ci.Game, ci.cp, ci.Game.Reset)
}

// Bet アンテベット
func (ci *CaribbeanStudInteractor) Bet(ante, jackpot int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.Bet(ante, jackpot) })
}

// Play プレイ
func (ci *CaribbeanStudInteractor) Play() string {
	return execAndPresent(ci.Game, ci.cp, ci.Game.Play)
}

// Fold フォールド
func (ci *CaribbeanStudInteractor) Fold() string {
	return execAndPresent(ci.Game, ci.cp, ci.Game.Fold)
}

// Hint ヒント取得
func (ci *CaribbeanStudInteractor) Hint() string {
	return ci.cp.HintOutput(ci.Game)
}

// ActionLog 棋譜を出力する
func (ci *CaribbeanStudInteractor) ActionLog() string {
	return ci.cp.ActionLogOutput(ci.Game)
}

// RestoreCaribbeanStudInteractor deserialises JSON into a CaribbeanStudInteractor.
func RestoreCaribbeanStudInteractor(data []byte, cp presenter.CaribbeanStudPresenter) (*CaribbeanStudInteractor, error) {
	return restoreAndBuild[domain.CaribbeanStud](data, func(g *domain.CaribbeanStud) *CaribbeanStudInteractor {
		return &CaribbeanStudInteractor{GameBase: GameBase[interfaces.CaribbeanStudGame]{Game: g}, cp: cp}
	})
}
