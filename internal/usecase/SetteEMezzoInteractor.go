//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SetteEMezzoInteractorIF セッテ・エ・メッツォ インタラクターインタフェース
type SetteEMezzoInteractorIF interface {
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
	// Matta マッタの値を選ぶ（半点単位）
	Matta(halves int) string
	// BankerHit 親が引く
	BankerHit() string
	// BankerStand 親が止める
	BankerStand() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// SetteEMezzoInteractor セッテ・エ・メッツォ インタラクタークラス
type SetteEMezzoInteractor struct {
	GameBase[interfaces.SetteEMezzoGame]
	sp presenter.SetteEMezzoPresenter
}

// NewSetteEMezzoInteractor コンストラクタ
func NewSetteEMezzoInteractor(s interfaces.SetteEMezzoGame, sp presenter.SetteEMezzoPresenter) *SetteEMezzoInteractor {
	mustNotNil("SetteEMezzoInteractor", map[string]any{"s": s, "sp": sp})
	return &SetteEMezzoInteractor{
		GameBase: GameBase[interfaces.SetteEMezzoGame]{Game: s},
		sp:       sp,
	}
}

// Reset ゲーム初期化
func (si *SetteEMezzoInteractor) Reset() string {
	return runAndPresent(si.Game, si.sp, si.Game.Reset)
}

// Bet ベットして配る
func (si *SetteEMezzoInteractor) Bet(amount int) string {
	return execAndPresent(si.Game, si.sp, func() error { return si.Game.PlaceBet(amount) })
}

// Deal 人間が親の局を配る
func (si *SetteEMezzoInteractor) Deal() string {
	return execAndPresent(si.Game, si.sp, si.Game.StartAsBanker)
}

// Hit 1 枚引く
func (si *SetteEMezzoInteractor) Hit() string {
	return execAndPresent(si.Game, si.sp, si.Game.Hit)
}

// Stand 引き止める
func (si *SetteEMezzoInteractor) Stand() string {
	return execAndPresent(si.Game, si.sp, si.Game.Stand)
}

// Matta マッタの値を選ぶ
func (si *SetteEMezzoInteractor) Matta(halves int) string {
	return execAndPresent(si.Game, si.sp, func() error { return si.Game.SetMattaValue(halves) })
}

// BankerHit 親が引く
func (si *SetteEMezzoInteractor) BankerHit() string {
	return execAndPresent(si.Game, si.sp, si.Game.BankerHit)
}

// BankerStand 親が止める
func (si *SetteEMezzoInteractor) BankerStand() string {
	return execAndPresent(si.Game, si.sp, si.Game.BankerStand)
}

// ActionLog 棋譜を出力する
func (si *SetteEMezzoInteractor) ActionLog() string {
	return si.sp.ActionLogOutput(si.Game)
}

// RestoreSetteEMezzoInteractor deserialises JSON into a SetteEMezzoInteractor.
func RestoreSetteEMezzoInteractor(data []byte, sp presenter.SetteEMezzoPresenter) (*SetteEMezzoInteractor, error) {
	return restoreAndBuild[domain.SetteEMezzo](data, func(g *domain.SetteEMezzo) *SetteEMezzoInteractor {
		return &SetteEMezzoInteractor{
			GameBase: GameBase[interfaces.SetteEMezzoGame]{Game: g},
			sp:       sp,
		}
	})
}
