//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// QuinzeInteractorIF カーンズ インタラクターインタフェース
type QuinzeInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Bet ベットして配る
	Bet(amount int) string
	// Deal 人間が親の局を配る
	Deal() string
	// Hit 1 枚引く
	Hit() string
	// Stand 引き止める
	Stand() string
	// BankerHit 親が引く
	BankerHit() string
	// BankerStand 親が止める
	BankerStand() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// QuinzeInteractor カーンズ インタラクタークラス
type QuinzeInteractor struct {
	GameBase[interfaces.QuinzeGame]
	sp presenter.QuinzePresenter
}

// NewQuinzeInteractor コンストラクタ
func NewQuinzeInteractor(s interfaces.QuinzeGame, sp presenter.QuinzePresenter) *QuinzeInteractor {
	mustNotNil("QuinzeInteractor", map[string]any{"s": s, "sp": sp})
	return &QuinzeInteractor{
		GameBase: GameBase[interfaces.QuinzeGame]{Game: s},
		sp:       sp,
	}
}

// Reset ゲーム初期化
func (si *QuinzeInteractor) Reset() string {
	return runAndPresent(si.Game, si.sp, si.Game.Reset)
}

// Bet ベットして配る
func (si *QuinzeInteractor) Bet(amount int) string {
	return execAndPresent(si.Game, si.sp, func() error { return si.Game.PlaceBet(amount) })
}

// Deal 人間が親の局を配る
func (si *QuinzeInteractor) Deal() string {
	return execAndPresent(si.Game, si.sp, si.Game.StartAsBanker)
}

// Hit 1 枚引く
func (si *QuinzeInteractor) Hit() string {
	return execAndPresent(si.Game, si.sp, si.Game.Hit)
}

// Stand 引き止める
func (si *QuinzeInteractor) Stand() string {
	return execAndPresent(si.Game, si.sp, si.Game.Stand)
}

// BankerHit 親が引く
func (si *QuinzeInteractor) BankerHit() string {
	return execAndPresent(si.Game, si.sp, si.Game.BankerHit)
}

// BankerStand 親が止める
func (si *QuinzeInteractor) BankerStand() string {
	return execAndPresent(si.Game, si.sp, si.Game.BankerStand)
}

// ActionLog 棋譜を出力する
func (si *QuinzeInteractor) ActionLog() string {
	return si.sp.ActionLogOutput(si.Game)
}

// RestoreQuinzeInteractor deserialises JSON into a QuinzeInteractor.
func RestoreQuinzeInteractor(data []byte, sp presenter.QuinzePresenter) (*QuinzeInteractor, error) {
	return restoreAndBuild[domain.Quinze](data, func(g *domain.Quinze) *QuinzeInteractor {
		return &QuinzeInteractor{
			GameBase: GameBase[interfaces.QuinzeGame]{Game: g},
			sp:       sp,
		}
	})
}
