//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// DoubleAttackBlackjackInteractorIF 追加ベット・ブラックジャックインタラクターインタフェース
type DoubleAttackBlackjackInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.DoubleAttackBlackjackConfig) string
	// PlaceBet アンティと任意の Bust It を置く
	PlaceBet(ante, bustIt int) string
	// Attack 追加ベットを置く
	Attack(amount int) string
	// Hit 1 枚引く
	Hit() string
	// Stand 打ち止めにする
	Stand() string
	// Double 倍にして 1 枚引く
	Double() string
	// Split 手札を分ける
	Split() string
	// NextRound 次のラウンドを始める
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.DoubleAttackBlackjackConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// DoubleAttackBlackjackInteractor 追加ベット・ブラックジャックインタラクタークラス
type DoubleAttackBlackjackInteractor struct {
	GameBase[interfaces.DoubleAttackBlackjackGame]
	cp presenter.DoubleAttackBlackjackPresenter
}

// NewDoubleAttackBlackjackInteractor コンストラクタ
func NewDoubleAttackBlackjackInteractor(c interfaces.DoubleAttackBlackjackGame,
	cp presenter.DoubleAttackBlackjackPresenter,
) *DoubleAttackBlackjackInteractor {
	mustNotNil("DoubleAttackBlackjackInteractor", map[string]any{"c": c, "cp": cp})
	return &DoubleAttackBlackjackInteractor{
		GameBase: GameBase[interfaces.DoubleAttackBlackjackGame]{Game: c}, cp: cp,
	}
}

// Reset ゲーム初期化
func (ci *DoubleAttackBlackjackInteractor) Reset() string {
	ci.Game.Reset()
	return ci.cp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *DoubleAttackBlackjackInteractor) ResetWithConfig(cfg domain.DoubleAttackBlackjackConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.cp, cfg, ci.Game.SetConfig, ci.Reset)
}

// PlaceBet アンティと任意の Bust It を置く
func (ci *DoubleAttackBlackjackInteractor) PlaceBet(ante, bustIt int) string {
	return ci.runGuarded(func() error { return ci.Game.PlaceBet(ante, bustIt) })
}

// Attack 追加ベットを置く
func (ci *DoubleAttackBlackjackInteractor) Attack(amount int) string {
	return ci.runGuarded(func() error { return ci.Game.Attack(amount) })
}

// Hit 1 枚引く
func (ci *DoubleAttackBlackjackInteractor) Hit() string { return ci.runGuarded(ci.Game.Hit) }

// Stand 打ち止めにする
func (ci *DoubleAttackBlackjackInteractor) Stand() string { return ci.runGuarded(ci.Game.Stand) }

// Double 倍にして 1 枚引く
func (ci *DoubleAttackBlackjackInteractor) Double() string { return ci.runGuarded(ci.Game.Double) }

// Split 手札を分ける
func (ci *DoubleAttackBlackjackInteractor) Split() string { return ci.runGuarded(ci.Game.Split) }

// NextRound 次のラウンドを始める
func (ci *DoubleAttackBlackjackInteractor) NextRound() string {
	return ci.runGuarded(ci.Game.NextRound)
}

// runGuarded は終局後の操作を弾いてから action を実行し、結果を出力する。
//
// **追加ベットの上限もダブル/スプリットの可否もドメインに任せます。** ここで
// 判定し直すと、このゲームの本体である「アップカードを見てから賭け増しできる」
// 規則が 2 か所に増えて必ずずれます。
func (ci *DoubleAttackBlackjackInteractor) runGuarded(action func() error) string {
	if out, blocked := guardGameEnd(ci.Game, ci.cp); blocked {
		return out
	}
	if err := action(); err != nil {
		return ci.cp.Output(ci.Game, err)
	}
	return ci.cp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を取得
func (ci *DoubleAttackBlackjackInteractor) GetConfig() domain.DoubleAttackBlackjackConfig {
	return ci.Game.GetConfig()
}

// Hint ヒント取得
func (ci *DoubleAttackBlackjackInteractor) Hint() string { return ci.cp.HintOutput(ci.Game) }

// ActionLog 棋譜を出力する
func (ci *DoubleAttackBlackjackInteractor) ActionLog() string {
	return ci.cp.ActionLogOutput(ci.Game)
}

// RestoreDoubleAttackBlackjackInteractor deserialises JSON into an interactor.
func RestoreDoubleAttackBlackjackInteractor(data []byte,
	cp presenter.DoubleAttackBlackjackPresenter,
) (*DoubleAttackBlackjackInteractor, error) {
	return restoreAndBuild[domain.DoubleAttackBlackjack](data,
		func(g *domain.DoubleAttackBlackjack) *DoubleAttackBlackjackInteractor {
			return NewDoubleAttackBlackjackInteractor(g, cp)
		})
}
