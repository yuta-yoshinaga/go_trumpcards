//go:build !js || !wasm || extra4

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// CasinoWarInteractorIF カジノウォーインタラクターインタフェース
type CasinoWarInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Bet アンテをベットして初手まで進める
	Bet(amount int) string
	// Surrender タイ時の降参
	Surrender() string
	// War タイ時のウォー宣言
	War() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// CasinoWarInteractor カジノウォーインタラクタークラス
type CasinoWarInteractor struct {
	GameBase[interfaces.CasinoWarGame]
	cp presenter.CasinoWarPresenter
}

// NewCasinoWarInteractor コンストラクタ
func NewCasinoWarInteractor(cw interfaces.CasinoWarGame, cp presenter.CasinoWarPresenter) *CasinoWarInteractor {
	mustNotNil("CasinoWarInteractor", map[string]any{"cw": cw, "cp": cp})
	return &CasinoWarInteractor{
		GameBase: GameBase[interfaces.CasinoWarGame]{Game: cw},
		cp:       cp,
	}
}

// Reset ゲーム初期化
func (ci *CasinoWarInteractor) Reset() string {
	return runAndPresent(ci.Game, ci.cp, ci.Game.Reset)
}

// Bet アンテをベットしカードを配る。Bet 直後に ResolveInitial を呼んで初手結果まで進める。
func (ci *CasinoWarInteractor) Bet(amount int) string {
	return execAndPresent(ci.Game, ci.cp, func() error {
		if err := ci.Game.Bet(amount); err != nil {
			return err
		}
		ci.Game.ResolveInitial()
		return nil
	})
}

// Surrender タイ時の降参
func (ci *CasinoWarInteractor) Surrender() string {
	return execAndPresent(ci.Game, ci.cp, ci.Game.Surrender)
}

// War タイ時のウォー宣言
func (ci *CasinoWarInteractor) War() string {
	return execAndPresent(ci.Game, ci.cp, ci.Game.GoToWar)
}

// ActionLog 棋譜を出力する
func (ci *CasinoWarInteractor) ActionLog() string {
	return ci.cp.ActionLogOutput(ci.Game)
}

// RestoreCasinoWarInteractor deserialises JSON into a CasinoWarInteractor.
func RestoreCasinoWarInteractor(data []byte, cp presenter.CasinoWarPresenter) (*CasinoWarInteractor, error) {
	return restoreAndBuild[domain.CasinoWar](data, func(g *domain.CasinoWar) *CasinoWarInteractor {
		return &CasinoWarInteractor{GameBase: GameBase[interfaces.CasinoWarGame]{Game: g}, cp: cp}
	})
}
