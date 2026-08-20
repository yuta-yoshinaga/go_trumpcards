//go:build !js || !wasm || extra4

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// DragonTigerInteractorIF ドラゴンタイガーインタラクターインタフェース
type DragonTigerInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Bet ベットしてゲームを進める
	Bet(amount, betType int) string
	// ClearHistory 罫線履歴をクリアする
	ClearHistory() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// DragonTigerInteractor ドラゴンタイガーインタラクタークラス
type DragonTigerInteractor struct {
	GameBase[interfaces.DragonTigerGame]
	dp presenter.DragonTigerPresenter
}

// NewDragonTigerInteractor コンストラクタ
func NewDragonTigerInteractor(dt interfaces.DragonTigerGame, dp presenter.DragonTigerPresenter) *DragonTigerInteractor {
	mustNotNil("DragonTigerInteractor", map[string]any{"dt": dt, "dp": dp})
	return &DragonTigerInteractor{
		GameBase: GameBase[interfaces.DragonTigerGame]{Game: dt},
		dp:       dp,
	}
}

// Reset ゲーム初期化
func (di *DragonTigerInteractor) Reset() string {
	return runAndPresent(di.Game, di.dp, di.Game.Reset)
}

// Bet ベットして勝敗判定まで自動進行
func (di *DragonTigerInteractor) Bet(amount, betType int) string {
	return execAndPresent(di.Game, di.dp, func() error {
		return di.Game.Bet(amount, betType)
	})
}

// ClearHistory 罫線履歴をクリアする
func (di *DragonTigerInteractor) ClearHistory() string {
	return runAndPresent(di.Game, di.dp, di.Game.ClearHistory)
}

// ActionLog 棋譜を出力する
func (di *DragonTigerInteractor) ActionLog() string {
	return di.dp.ActionLogOutput(di.Game)
}

// RestoreDragonTigerInteractor deserialises JSON into a DragonTigerInteractor.
func RestoreDragonTigerInteractor(data []byte, dp presenter.DragonTigerPresenter) (*DragonTigerInteractor, error) {
	return restoreAndBuild[domain.DragonTiger](data, func(g *domain.DragonTiger) *DragonTigerInteractor {
		return &DragonTigerInteractor{GameBase: GameBase[interfaces.DragonTigerGame]{Game: g}, dp: dp}
	})
}
