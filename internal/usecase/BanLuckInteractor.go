//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BanLuckInteractorIF バンラックインタラクターインタフェース
type BanLuckInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.BanLuckConfig) string
	// PlaceBet 賭け金を置く
	PlaceBet(bet int) string
	// Hit 1 枚引く
	Hit() string
	// Stand 打ち止めにする
	Stand() string
	// NextRound 次のラウンドを始める
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.BanLuckConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// BanLuckInteractor バンラックインタラクタークラス
type BanLuckInteractor struct {
	GameBase[interfaces.BanLuckGame]
	cp presenter.BanLuckPresenter
}

// NewBanLuckInteractor コンストラクタ
func NewBanLuckInteractor(c interfaces.BanLuckGame, cp presenter.BanLuckPresenter) *BanLuckInteractor {
	mustNotNil("BanLuckInteractor", map[string]any{"c": c, "cp": cp})
	return &BanLuckInteractor{GameBase: GameBase[interfaces.BanLuckGame]{Game: c}, cp: cp}
}

// Reset ゲーム初期化
func (ci *BanLuckInteractor) Reset() string {
	ci.Game.Reset()
	return ci.cp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *BanLuckInteractor) ResetWithConfig(cfg domain.BanLuckConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.cp, cfg, ci.Game.SetConfig, ci.Reset)
}

// PlaceBet 賭け金を置く
func (ci *BanLuckInteractor) PlaceBet(bet int) string {
	return ci.runGuarded(func() error { return ci.Game.PlaceBet(bet) })
}

// Hit 1 枚引く
func (ci *BanLuckInteractor) Hit() string { return ci.runGuarded(ci.Game.Hit) }

// Stand 打ち止めにする
func (ci *BanLuckInteractor) Stand() string { return ci.runGuarded(ci.Game.Stand) }

// NextRound 次のラウンドを始める
func (ci *BanLuckInteractor) NextRound() string { return ci.runGuarded(ci.Game.NextRound) }

// runGuarded は終局後の操作を弾いてから action を実行し、CPU を進めて出力する。
//
// **人間が 1 手指したら CPU を進める。** ここを忘れると、ドメインも adapter も
// 全部緑のまま盤面だけが止まる ── テストが自分でループを回していると見えない
// 種類のバグなので、駆動はこの 1 か所に置いて必ず通す。
//
// **親の義務ヒットや役の判定はドメインに任せる。** 15 未満かどうかをここで
// 見直すと、このゲームの本体である規則が 2 か所に増えて必ずずれる。
func (ci *BanLuckInteractor) runGuarded(action func() error) string {
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
func (ci *BanLuckInteractor) GetConfig() domain.BanLuckConfig { return ci.Game.GetConfig() }

// Hint ヒント取得
func (ci *BanLuckInteractor) Hint() string { return ci.cp.HintOutput(ci.Game) }

// ActionLog 棋譜を出力する
func (ci *BanLuckInteractor) ActionLog() string { return ci.cp.ActionLogOutput(ci.Game) }

// RestoreBanLuckInteractor deserialises JSON into an interactor.
func RestoreBanLuckInteractor(data []byte, cp presenter.BanLuckPresenter) (*BanLuckInteractor, error) {
	return restoreAndBuild[domain.BanLuck](data,
		func(g *domain.BanLuck) *BanLuckInteractor { return NewBanLuckInteractor(g, cp) })
}
