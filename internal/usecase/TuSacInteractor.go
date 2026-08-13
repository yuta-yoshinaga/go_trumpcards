//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// TuSacInteractorIF 四色牌インタラクターインタフェース
type TuSacInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.TuSacConfig) string
	// Draw 山または捨て札から 1 枚引く
	Draw(fromDiscard bool) string
	// Meld 手札の指定の札を場に出す
	Meld(indexes []int) string
	// Discard 手札から 1 枚捨てる
	Discard(index int) string
	// NextRound 次のラウンドを始める
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.TuSacConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// TuSacInteractor 四色牌インタラクタークラス
type TuSacInteractor struct {
	GameBase[interfaces.TuSacGame]
	cp presenter.TuSacPresenter
}

// NewTuSacInteractor コンストラクタ
func NewTuSacInteractor(c interfaces.TuSacGame, cp presenter.TuSacPresenter) *TuSacInteractor {
	mustNotNil("TuSacInteractor", map[string]any{"c": c, "cp": cp})
	return &TuSacInteractor{GameBase: GameBase[interfaces.TuSacGame]{Game: c}, cp: cp}
}

// Reset ゲーム初期化
func (ci *TuSacInteractor) Reset() string {
	ci.Game.Reset()
	return ci.cp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *TuSacInteractor) ResetWithConfig(cfg domain.TuSacConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.cp, cfg, ci.Game.SetConfig, ci.Reset)
}

// Draw 山または捨て札から 1 枚引く
func (ci *TuSacInteractor) Draw(fromDiscard bool) string {
	return ci.runGuarded(func() error { return ci.Game.Draw(fromDiscard) })
}

// Meld は指定の札を場に出す。
//
// **出す組み合わせは選ばせる。** 出せるものを勝手に全部出すと、卒 5 枚を
// 狙って 3 枚組を我慢する、という判断が消える ── 場に出た札は戻せないので、
// 何を出さずに残すかがこのゲームの読みどころ。
func (ci *TuSacInteractor) Meld(indexes []int) string {
	return ci.runGuarded(func() error { return ci.Game.Meld(indexes) })
}

// Discard 手札から 1 枚捨てる
func (ci *TuSacInteractor) Discard(index int) string {
	return ci.runGuarded(func() error { return ci.Game.Discard(index) })
}

// NextRound 次のラウンドを始める
func (ci *TuSacInteractor) NextRound() string { return ci.runGuarded(ci.Game.NextRound) }

// runGuarded は終局後の操作を弾いてから action を実行して出力する。
//
// **CPU を進める呼び出しは要らない。** 捨てた時点でドメインが CPU の手番を
// 回し切り、人間の手番か決着で戻ってくる。
func (ci *TuSacInteractor) runGuarded(action func() error) string {
	if out, blocked := guardGameEnd(ci.Game, ci.cp); blocked {
		return out
	}
	if err := action(); err != nil {
		return ci.cp.Output(ci.Game, err)
	}
	return ci.cp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を取得
func (ci *TuSacInteractor) GetConfig() domain.TuSacConfig { return ci.Game.GetConfig() }

// Hint ヒント取得
func (ci *TuSacInteractor) Hint() string { return ci.cp.HintOutput(ci.Game) }

// ActionLog 棋譜を出力する
func (ci *TuSacInteractor) ActionLog() string { return ci.cp.ActionLogOutput(ci.Game) }

// RestoreTuSacInteractor deserialises JSON into an interactor.
func RestoreTuSacInteractor(data []byte, cp presenter.TuSacPresenter) (*TuSacInteractor, error) {
	return restoreAndBuild[domain.TuSac](data,
		func(g *domain.TuSac) *TuSacInteractor { return NewTuSacInteractor(g, cp) })
}
