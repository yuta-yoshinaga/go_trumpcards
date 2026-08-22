//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// CaribbeanDrawInteractorIF カリビアン・ドロー・ポーカーインタラクターインタフェース
type CaribbeanDrawInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Bet アンテベット
	Bet(ante, jackpot int) string
	// Play プレイ（コール）
	Draw(indices []int) string
	Play() string
	// Fold フォールド
	Fold() string
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// CaribbeanDrawInteractor カリビアン・ドロー・ポーカーインタラクタークラス
type CaribbeanDrawInteractor struct {
	GameBase[interfaces.CaribbeanDrawGame]
	cp presenter.CaribbeanDrawPresenter
}

// NewCaribbeanDrawInteractor コンストラクタ
func NewCaribbeanDrawInteractor(cs interfaces.CaribbeanDrawGame, cp presenter.CaribbeanDrawPresenter) *CaribbeanDrawInteractor {
	mustNotNil("CaribbeanDrawInteractor", map[string]any{"cs": cs, "cp": cp})
	return &CaribbeanDrawInteractor{
		GameBase: GameBase[interfaces.CaribbeanDrawGame]{Game: cs},
		cp:       cp,
	}
}

// Reset ゲーム初期化
func (ci *CaribbeanDrawInteractor) Reset() string {
	return runAndPresent(ci.Game, ci.cp, ci.Game.Reset)
}

// Bet アンテベット
func (ci *CaribbeanDrawInteractor) Bet(ante, jackpot int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.Bet(ante, jackpot) })
}

// Draw は手札のうち最大2枚を交換する。空なら交換しない。
func (ci *CaribbeanDrawInteractor) Draw(indices []int) string {
	return execAndPresent(ci.Game, ci.cp, func() error { return ci.Game.Draw(indices) })
}

// Play プレイ
func (ci *CaribbeanDrawInteractor) Play() string {
	return execAndPresent(ci.Game, ci.cp, ci.Game.Play)
}

// Fold フォールド
func (ci *CaribbeanDrawInteractor) Fold() string {
	return execAndPresent(ci.Game, ci.cp, ci.Game.Fold)
}

// Hint ヒント取得
func (ci *CaribbeanDrawInteractor) Hint() string {
	return ci.cp.HintOutput(ci.Game)
}

// ActionLog 棋譜を出力する
func (ci *CaribbeanDrawInteractor) ActionLog() string {
	return ci.cp.ActionLogOutput(ci.Game)
}

// RestoreCaribbeanDrawInteractor deserialises JSON into a CaribbeanDrawInteractor.
func RestoreCaribbeanDrawInteractor(data []byte, cp presenter.CaribbeanDrawPresenter) (*CaribbeanDrawInteractor, error) {
	return restoreAndBuild[domain.CaribbeanDraw](data, func(g *domain.CaribbeanDraw) *CaribbeanDrawInteractor {
		return &CaribbeanDrawInteractor{GameBase: GameBase[interfaces.CaribbeanDrawGame]{Game: g}, cp: cp}
	})
}
