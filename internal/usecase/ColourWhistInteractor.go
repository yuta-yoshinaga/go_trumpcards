//go:build !js || !wasm || classic

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// ColourWhistInteractorIF カラーホイストインタラクターインタフェース
type ColourWhistInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.ColourWhistConfig) string
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
	GetConfig() domain.ColourWhistConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// ColourWhistInteractor カラーホイストインタラクタークラス
type ColourWhistInteractor struct {
	GameBase[interfaces.ColourWhistGame]
	cp presenter.ColourWhistPresenter
}

// NewColourWhistInteractor コンストラクタ
func NewColourWhistInteractor(c interfaces.ColourWhistGame, cp presenter.ColourWhistPresenter) *ColourWhistInteractor {
	mustNotNil("ColourWhistInteractor", map[string]any{"c": c, "cp": cp})
	return &ColourWhistInteractor{GameBase: GameBase[interfaces.ColourWhistGame]{Game: c}, cp: cp}
}

// Reset ゲーム初期化
func (ci *ColourWhistInteractor) Reset() string {
	ci.Game.Reset()
	return ci.cp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *ColourWhistInteractor) ResetWithConfig(cfg domain.ColourWhistConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.cp, cfg, ci.Game.SetConfig, ci.Reset)
}

// Bid 契約を宣言する
func (ci *ColourWhistInteractor) Bid(contract int) string {
	return ci.runGuarded(func() error { return ci.Game.Bid(contract) })
}

// Call 切り札を決める
func (ci *ColourWhistInteractor) Call(trumpSuit int) string {
	return ci.runGuarded(func() error { return ci.Game.Call(trumpSuit) })
}

// PlayCard 札を出す
func (ci *ColourWhistInteractor) PlayCard(cardIndex int) string {
	return ci.runGuarded(func() error { return ci.Game.PlayCard(cardIndex) })
}

// NextRound 次のラウンドを配る
func (ci *ColourWhistInteractor) NextRound() string {
	return ci.runGuarded(ci.Game.NextRound)
}

// runGuarded は終局後の操作を弾いてから action を実行し、結果を出力する。
//
// **フェーズ判定はドメインに任せます。** 二重に書くとずれます。
func (ci *ColourWhistInteractor) runGuarded(action func() error) string {
	if out, blocked := guardGameEnd(ci.Game, ci.cp); blocked {
		return out
	}
	if err := action(); err != nil {
		return ci.cp.Output(ci.Game, err)
	}
	return ci.cp.Output(ci.Game, nil)
}

// GiveUp 投了する
func (ci *ColourWhistInteractor) GiveUp() string {
	if out, blocked := guardGameEnd(ci.Game, ci.cp); blocked {
		return out
	}
	ci.Game.GiveUp()
	return ci.cp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を取得
func (ci *ColourWhistInteractor) GetConfig() domain.ColourWhistConfig { return ci.Game.GetConfig() }

// Hint ヒント取得
func (ci *ColourWhistInteractor) Hint() string { return ci.cp.HintOutput(ci.Game) }

// ActionLog 棋譜を出力する
func (ci *ColourWhistInteractor) ActionLog() string { return ci.cp.ActionLogOutput(ci.Game) }

// RestoreColourWhistInteractor deserialises JSON into a ColourWhistInteractor.
func RestoreColourWhistInteractor(data []byte, cp presenter.ColourWhistPresenter) (*ColourWhistInteractor, error) {
	return restoreAndBuild[domain.ColourWhist](data, func(g *domain.ColourWhist) *ColourWhistInteractor {
		return NewColourWhistInteractor(g, cp)
	})
}
