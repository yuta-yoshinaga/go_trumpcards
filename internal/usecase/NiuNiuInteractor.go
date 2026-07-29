//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// NiuNiuInteractorIF 闘牛 インタラクターインタフェース
type NiuNiuInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Bet ベットして配り、精算まで進める
	Bet(amount int) string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// NiuNiuInteractor 闘牛 インタラクタークラス
type NiuNiuInteractor struct {
	GameBase[interfaces.NiuNiuGame]
	np presenter.NiuNiuPresenter
}

// NewNiuNiuInteractor コンストラクタ
func NewNiuNiuInteractor(n interfaces.NiuNiuGame, np presenter.NiuNiuPresenter) *NiuNiuInteractor {
	mustNotNil("NiuNiuInteractor", map[string]any{"n": n, "np": np})
	return &NiuNiuInteractor{
		GameBase: GameBase[interfaces.NiuNiuGame]{Game: n},
		np:       np,
	}
}

// Reset ゲーム初期化
func (ni *NiuNiuInteractor) Reset() string {
	return runAndPresent(ni.Game, ni.np, ni.Game.Reset)
}

// Bet ベットして配り、精算まで進める
func (ni *NiuNiuInteractor) Bet(amount int) string {
	return execAndPresent(ni.Game, ni.np, func() error { return ni.Game.PlaceBet(amount) })
}

// ActionLog 棋譜を出力する
func (ni *NiuNiuInteractor) ActionLog() string {
	return ni.np.ActionLogOutput(ni.Game)
}

// RestoreNiuNiuInteractor deserialises JSON into a NiuNiuInteractor.
func RestoreNiuNiuInteractor(data []byte, np presenter.NiuNiuPresenter) (*NiuNiuInteractor, error) {
	return restoreAndBuild[domain.NiuNiu](data, func(g *domain.NiuNiu) *NiuNiuInteractor {
		return &NiuNiuInteractor{
			GameBase: GameBase[interfaces.NiuNiuGame]{Game: g},
			np:       np,
		}
	})
}
