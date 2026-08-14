//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// CincinnatiInteractorIF シンシナティインタラクターインタフェース
type CincinnatiInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.CincinnatiConfig) string
	// Action 人間の手を処理する
	Action(action, amount int) string
	// NextHand 次のハンドを始める
	NextHand() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.CincinnatiConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// CincinnatiInteractor シンシナティインタラクタークラス
type CincinnatiInteractor struct {
	GameBase[interfaces.CincinnatiGame]
	cp presenter.CincinnatiPresenter
}

// NewCincinnatiInteractor コンストラクタ
func NewCincinnatiInteractor(c interfaces.CincinnatiGame, cp presenter.CincinnatiPresenter) *CincinnatiInteractor {
	mustNotNil("CincinnatiInteractor", map[string]any{"c": c, "cp": cp})
	return &CincinnatiInteractor{GameBase: GameBase[interfaces.CincinnatiGame]{Game: c}, cp: cp}
}

// Reset ゲーム初期化
func (ci *CincinnatiInteractor) Reset() string {
	ci.Game.Reset()
	return ci.cp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *CincinnatiInteractor) ResetWithConfig(cfg domain.CincinnatiConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.cp, cfg, ci.Game.SetConfig, ci.Reset)
}

// Action 人間の手を処理する
//
// **手の合法性はドメインに任せる。** チェックできるか、レイズの上限に達して
// いないかは場況で決まる ── ここで真似ると規則が 2 か所に増えてずれる。
func (ci *CincinnatiInteractor) Action(action, amount int) string {
	return ci.runGuarded(func() error { return ci.Game.PlayerAction(action, amount) })
}

// NextHand 次のハンドを始める
func (ci *CincinnatiInteractor) NextHand() string { return ci.runGuarded(ci.Game.NextHand) }

// runGuarded は終局後の操作を弾いてから action を実行し、CPU を進めて出力する。
//
// **人間が 1 手指したら CPU を進める。** 忘れるとドメインも adapter も全部緑の
// まま盤面が止まる。ドメイン側でも進めているが、`NextHand` の直後のように
// 人間の手を経ない経路があるので、ここでも必ず通す。
func (ci *CincinnatiInteractor) runGuarded(action func() error) string {
	if out, blocked := guardGameEnd(ci.Game, ci.cp); blocked {
		return out
	}
	if err := action(); err != nil {
		return ci.cp.Output(ci.Game, err)
	}
	ci.Game.CpuPlay()
	return ci.cp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を取得
func (ci *CincinnatiInteractor) GetConfig() domain.CincinnatiConfig { return ci.Game.GetConfig() }

// Hint ヒント取得
func (ci *CincinnatiInteractor) Hint() string { return ci.cp.HintOutput(ci.Game) }

// ActionLog 棋譜を出力する
func (ci *CincinnatiInteractor) ActionLog() string { return ci.cp.ActionLogOutput(ci.Game) }

// RestoreCincinnatiInteractor deserialises JSON into an interactor.
func RestoreCincinnatiInteractor(data []byte, cp presenter.CincinnatiPresenter) (*CincinnatiInteractor, error) {
	return restoreAndBuild[domain.Cincinnati](data,
		func(g *domain.Cincinnati) *CincinnatiInteractor { return NewCincinnatiInteractor(g, cp) })
}
