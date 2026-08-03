//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// PontoonInteractorIF ポンツーン インタラクターインタフェース
type PontoonInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Bet ベットして配る
	Bet(amount int) string
	// Deal 人間が親の局を配る
	Deal() string
	// Stick スティック
	Stick() string
	// Twist ツイスト
	Twist() string
	// Buy バイ
	Buy(extra int) string
	// Split スプリット
	Split() string
	// BankerTwist 親が引く
	BankerTwist() string
	// BankerStay 親が止める
	BankerStay() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// PontoonInteractor ポンツーン インタラクタークラス
type PontoonInteractor struct {
	GameBase[interfaces.PontoonGame]
	pp presenter.PontoonPresenter
}

// NewPontoonInteractor コンストラクタ
func NewPontoonInteractor(p interfaces.PontoonGame, pp presenter.PontoonPresenter) *PontoonInteractor {
	mustNotNil("PontoonInteractor", map[string]any{"p": p, "pp": pp})
	return &PontoonInteractor{
		GameBase: GameBase[interfaces.PontoonGame]{Game: p},
		pp:       pp,
	}
}

// Reset ゲーム初期化
func (pi *PontoonInteractor) Reset() string {
	return runAndPresent(pi.Game, pi.pp, pi.Game.Reset)
}

// Bet ベットして配る
func (pi *PontoonInteractor) Bet(amount int) string {
	return execAndPresent(pi.Game, pi.pp, func() error { return pi.Game.PlaceBet(amount) })
}

// Deal 人間が親の局を配る
func (pi *PontoonInteractor) Deal() string {
	return execAndPresent(pi.Game, pi.pp, pi.Game.StartAsBanker)
}

// Stick スティック
func (pi *PontoonInteractor) Stick() string {
	return execAndPresent(pi.Game, pi.pp, pi.Game.Stick)
}

// Twist ツイスト
func (pi *PontoonInteractor) Twist() string {
	return execAndPresent(pi.Game, pi.pp, pi.Game.Twist)
}

// Buy バイ
func (pi *PontoonInteractor) Buy(extra int) string {
	return execAndPresent(pi.Game, pi.pp, func() error { return pi.Game.Buy(extra) })
}

// Split スプリット
func (pi *PontoonInteractor) Split() string {
	return execAndPresent(pi.Game, pi.pp, pi.Game.Split)
}

// BankerTwist 親が引く
func (pi *PontoonInteractor) BankerTwist() string {
	return execAndPresent(pi.Game, pi.pp, pi.Game.BankerTwist)
}

// BankerStay 親が止める
func (pi *PontoonInteractor) BankerStay() string {
	return execAndPresent(pi.Game, pi.pp, pi.Game.BankerStay)
}

// ActionLog 棋譜を出力する
func (pi *PontoonInteractor) ActionLog() string {
	return pi.pp.ActionLogOutput(pi.Game)
}

// RestorePontoonInteractor deserialises JSON into a PontoonInteractor.
func RestorePontoonInteractor(data []byte, pp presenter.PontoonPresenter) (*PontoonInteractor, error) {
	return restoreAndBuild[domain.Pontoon](data, func(g *domain.Pontoon) *PontoonInteractor {
		return &PontoonInteractor{
			GameBase: GameBase[interfaces.PontoonGame]{Game: g},
			pp:       pp,
		}
	})
}
