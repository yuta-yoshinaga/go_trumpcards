//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// RikkenInteractorIF リッケンインタラクターインタフェース
type RikkenInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.RikkenConfig) string
	// Bid 契約を宣言する
	Bid(contract int) string
	// Call 切り札を決める
	Call(trumpSuit int) string
	// PlayCard 札を出す
	PlayCard(cardIndex int) string
	// NextRound 次のラウンドを配る
	NextRound() string
	// GiveUp 投了する
	GiveUp() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.RikkenConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// RikkenInteractor リッケンインタラクタークラス
type RikkenInteractor struct {
	GameBase[interfaces.RikkenGame]
	rp presenter.RikkenPresenter
}

// NewRikkenInteractor コンストラクタ
func NewRikkenInteractor(r interfaces.RikkenGame, rp presenter.RikkenPresenter) *RikkenInteractor {
	mustNotNil("RikkenInteractor", map[string]any{"r": r, "rp": rp})
	return &RikkenInteractor{GameBase: GameBase[interfaces.RikkenGame]{Game: r}, rp: rp}
}

// Reset ゲーム初期化
func (ri *RikkenInteractor) Reset() string {
	ri.Game.Reset()
	return ri.rp.Output(ri.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ri *RikkenInteractor) ResetWithConfig(cfg domain.RikkenConfig) string {
	return resetWithValidatedConfig(ri.Game, ri.rp, cfg, ri.Game.SetConfig, ri.Reset)
}

// Bid 契約を宣言する
func (ri *RikkenInteractor) Bid(contract int) string {
	return ri.runGuarded(func() error { return ri.Game.Bid(contract) })
}

// Call 切り札を決める
func (ri *RikkenInteractor) Call(trumpSuit int) string {
	return ri.runGuarded(func() error { return ri.Game.Call(trumpSuit) })
}

// PlayCard 札を出す
func (ri *RikkenInteractor) PlayCard(cardIndex int) string {
	return ri.runGuarded(func() error { return ri.Game.PlayCard(cardIndex) })
}

// NextRound 次のラウンドを配る
func (ri *RikkenInteractor) NextRound() string {
	return ri.runGuarded(ri.Game.NextRound)
}

// runGuarded は終局後の操作を弾いてから action を実行し、結果を出力する。
//
// **フェーズ判定はドメインに任せます。** 各メソッドが自分で見ているので、
// ここに二重に書くと片方だけ直したときに黙ってずれます。
func (ri *RikkenInteractor) runGuarded(action func() error) string {
	if out, blocked := guardGameEnd(ri.Game, ri.rp); blocked {
		return out
	}
	if err := action(); err != nil {
		return ri.rp.Output(ri.Game, err)
	}
	return ri.rp.Output(ri.Game, nil)
}

// GiveUp 投了する
func (ri *RikkenInteractor) GiveUp() string {
	if out, blocked := guardGameEnd(ri.Game, ri.rp); blocked {
		return out
	}
	ri.Game.GiveUp()
	return ri.rp.Output(ri.Game, nil)
}

// GetConfig 現在の設定を取得
func (ri *RikkenInteractor) GetConfig() domain.RikkenConfig { return ri.Game.GetConfig() }

// Hint ヒント取得
func (ri *RikkenInteractor) Hint() string { return ri.rp.HintOutput(ri.Game) }

// ActionLog 棋譜を出力する
func (ri *RikkenInteractor) ActionLog() string { return ri.rp.ActionLogOutput(ri.Game) }

// RestoreRikkenInteractor deserialises JSON into a RikkenInteractor.
func RestoreRikkenInteractor(data []byte, rp presenter.RikkenPresenter) (*RikkenInteractor, error) {
	return restoreAndBuild[domain.Rikken](data, func(g *domain.Rikken) *RikkenInteractor {
		return NewRikkenInteractor(g, rp)
	})
}
