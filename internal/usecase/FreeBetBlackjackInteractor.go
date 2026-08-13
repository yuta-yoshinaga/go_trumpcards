//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// FreeBetBlackjackInteractorIF フリーベット・ブラックジャックインタラクターインタフェース
type FreeBetBlackjackInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.FreeBetBlackjackConfig) string
	// PlaceBet アンティを置く
	PlaceBet(ante int) string
	// Hit 1 枚引く
	Hit() string
	// Stand 打ち止めにする
	Stand() string
	// FreeDouble ハウス持ちで倍にする
	FreeDouble() string
	// FreeSplit ハウス持ちで分ける
	FreeSplit() string
	// NextRound 次のラウンドを始める
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.FreeBetBlackjackConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// FreeBetBlackjackInteractor フリーベット・ブラックジャックインタラクタークラス
type FreeBetBlackjackInteractor struct {
	GameBase[interfaces.FreeBetBlackjackGame]
	cp presenter.FreeBetBlackjackPresenter
}

// NewFreeBetBlackjackInteractor コンストラクタ
func NewFreeBetBlackjackInteractor(c interfaces.FreeBetBlackjackGame,
	cp presenter.FreeBetBlackjackPresenter,
) *FreeBetBlackjackInteractor {
	mustNotNil("FreeBetBlackjackInteractor", map[string]any{"c": c, "cp": cp})
	return &FreeBetBlackjackInteractor{
		GameBase: GameBase[interfaces.FreeBetBlackjackGame]{Game: c}, cp: cp,
	}
}

// Reset ゲーム初期化
func (ci *FreeBetBlackjackInteractor) Reset() string {
	ci.Game.Reset()
	return ci.cp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *FreeBetBlackjackInteractor) ResetWithConfig(cfg domain.FreeBetBlackjackConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.cp, cfg, ci.Game.SetConfig, ci.Reset)
}

// PlaceBet アンティを置く
func (ci *FreeBetBlackjackInteractor) PlaceBet(ante int) string {
	return ci.runGuarded(func() error { return ci.Game.PlaceBet(ante) })
}

// Hit 1 枚引く
func (ci *FreeBetBlackjackInteractor) Hit() string { return ci.runGuarded(ci.Game.Hit) }

// Stand 打ち止めにする
func (ci *FreeBetBlackjackInteractor) Stand() string { return ci.runGuarded(ci.Game.Stand) }

// FreeDouble ハウス持ちで倍にする
func (ci *FreeBetBlackjackInteractor) FreeDouble() string {
	return ci.runGuarded(ci.Game.FreeDouble)
}

// FreeSplit ハウス持ちで分ける
func (ci *FreeBetBlackjackInteractor) FreeSplit() string {
	return ci.runGuarded(ci.Game.FreeSplit)
}

// NextRound 次のラウンドを始める
func (ci *FreeBetBlackjackInteractor) NextRound() string {
	return ci.runGuarded(ci.Game.NextRound)
}

// runGuarded は終局後の操作を弾いてから action を実行し、結果を出力する。
//
// **無料ダブル / 無料スプリットの可否はドメインに任せます。** ハードの 9-11 か、
// 10 点札のペアでないか、といった条件をここで判定し直すと、このゲームの本体である
// 規則が 2 か所に増えて必ずずれます。
func (ci *FreeBetBlackjackInteractor) runGuarded(action func() error) string {
	if out, blocked := guardGameEnd(ci.Game, ci.cp); blocked {
		return out
	}
	if err := action(); err != nil {
		return ci.cp.Output(ci.Game, err)
	}
	return ci.cp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を取得
func (ci *FreeBetBlackjackInteractor) GetConfig() domain.FreeBetBlackjackConfig {
	return ci.Game.GetConfig()
}

// Hint ヒント取得
func (ci *FreeBetBlackjackInteractor) Hint() string { return ci.cp.HintOutput(ci.Game) }

// ActionLog 棋譜を出力する
func (ci *FreeBetBlackjackInteractor) ActionLog() string { return ci.cp.ActionLogOutput(ci.Game) }

// RestoreFreeBetBlackjackInteractor deserialises JSON into an interactor.
func RestoreFreeBetBlackjackInteractor(data []byte,
	cp presenter.FreeBetBlackjackPresenter,
) (*FreeBetBlackjackInteractor, error) {
	return restoreAndBuild[domain.FreeBetBlackjack](data,
		func(g *domain.FreeBetBlackjack) *FreeBetBlackjackInteractor {
			return NewFreeBetBlackjackInteractor(g, cp)
		})
}
