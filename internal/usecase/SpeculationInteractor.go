//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SpeculationInteractorIF スペキュレーションインタラクターインタフェース
type SpeculationInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Flip 手番の席の伏せ札を1枚めくる
	Flip() string
	// Accept 競りの申し出を受ける
	Accept() string
	// Decline 競りの申し出を断る
	Decline() string
	// Bid 提示額に上乗せして買う
	Bid(amount int) string
	// NextRound 次のラウンドを始める
	NextRound() string
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// SpeculationInteractor スペキュレーションインタラクタークラス
type SpeculationInteractor struct {
	GameBase[interfaces.SpeculationGame]
	cp presenter.SpeculationPresenter
}

// NewSpeculationInteractor コンストラクタ
func NewSpeculationInteractor(c interfaces.SpeculationGame, cp presenter.SpeculationPresenter) *SpeculationInteractor {
	mustNotNil("SpeculationInteractor", map[string]any{"c": c, "cp": cp})
	return &SpeculationInteractor{GameBase: GameBase[interfaces.SpeculationGame]{Game: c}, cp: cp}
}

// Reset ゲーム初期化
func (ci *SpeculationInteractor) Reset() string {
	ci.Game.Reset()
	return ci.cp.Output(ci.Game, nil)
}

// Flip は手番の席の伏せ札を1枚めくる。
func (ci *SpeculationInteractor) Flip() string { return ci.runGuarded(ci.Game.Flip) }

// Accept は競りの申し出を受ける。
//
// **損な取引かどうかはここで判断しない。** 高く買いすぎるのもプレイヤーが
// 選べる手であって、ユースケースが止めるものではない。
func (ci *SpeculationInteractor) Accept() string { return ci.runGuarded(ci.Game.Accept) }

// Decline は競りの申し出を断る。
func (ci *SpeculationInteractor) Decline() string { return ci.runGuarded(ci.Game.Decline) }

// Bid は提示額に上乗せして買う。
func (ci *SpeculationInteractor) Bid(amount int) string {
	return ci.runGuarded(func() error { return ci.Game.Bid(amount) })
}

// NextRound 次のラウンドを始める
func (ci *SpeculationInteractor) NextRound() string { return ci.runGuarded(ci.Game.NextRound) }

// runGuarded は終局後の操作を弾いてから action を実行し、結果を出力する。
func (ci *SpeculationInteractor) runGuarded(action func() error) string {
	if out, blocked := guardGameEnd(ci.Game, ci.cp); blocked {
		return out
	}
	if err := action(); err != nil {
		return ci.cp.Output(ci.Game, err)
	}
	return ci.cp.Output(ci.Game, nil)
}

// Hint ヒント取得
func (ci *SpeculationInteractor) Hint() string { return ci.cp.HintOutput(ci.Game) }

// ActionLog 棋譜を出力する
func (ci *SpeculationInteractor) ActionLog() string { return ci.cp.ActionLogOutput(ci.Game) }

// RestoreSpeculationInteractor deserialises JSON into an interactor.
func RestoreSpeculationInteractor(data []byte, cp presenter.SpeculationPresenter) (*SpeculationInteractor, error) {
	return restoreAndBuild[domain.Speculation](data,
		func(g *domain.Speculation) *SpeculationInteractor { return NewSpeculationInteractor(g, cp) })
}
