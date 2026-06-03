//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BlackJackSwitchInteractorIF ブラックジャック・スイッチインタラクターインタフェース
type BlackJackSwitchInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Bet 1ハンドあたりの amount をベットしてカードを配る
	Bet(amount int) string
	// Switch 2ハンドの2枚目を交換してアクションフェーズへ進む
	Switch() string
	// Keep スイッチを行わずアクションフェーズへ進む
	Keep() string
	// Hit ヒット
	Hit() string
	// Stand スタンド
	Stand() string
	// DoubleDown ダブルダウン
	DoubleDown() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// BlackJackSwitchInteractor ブラックジャック・スイッチインタラクター
type BlackJackSwitchInteractor struct {
	GameBase[interfaces.BlackJackSwitchGame]
	dp presenter.BlackJackSwitchPresenter
}

// NewBlackJackSwitchInteractor コンストラクタ
func NewBlackJackSwitchInteractor(game interfaces.BlackJackSwitchGame, dp presenter.BlackJackSwitchPresenter) *BlackJackSwitchInteractor {
	mustNotNil("BlackJackSwitchInteractor", map[string]any{"game": game, "dp": dp})
	return &BlackJackSwitchInteractor{
		GameBase: GameBase[interfaces.BlackJackSwitchGame]{Game: game},
		dp:       dp,
	}
}

// Reset ゲーム初期化
func (bi *BlackJackSwitchInteractor) Reset() string {
	return runAndPresent(bi.Game, bi.dp, bi.Game.Reset)
}

// Bet 1ハンドあたりの amount をベット
func (bi *BlackJackSwitchInteractor) Bet(amount int) string {
	return execAndPresent(bi.Game, bi.dp, func() error {
		return bi.Game.PlayerBet(amount)
	})
}

// Switch 2枚目交換
func (bi *BlackJackSwitchInteractor) Switch() string {
	return execAndPresent(bi.Game, bi.dp, bi.Game.PlayerSwitch)
}

// Keep スイッチ放棄
func (bi *BlackJackSwitchInteractor) Keep() string {
	return execAndPresent(bi.Game, bi.dp, bi.Game.PlayerKeep)
}

// Hit ヒット
func (bi *BlackJackSwitchInteractor) Hit() string {
	return execAndPresent(bi.Game, bi.dp, bi.Game.PlayerHit)
}

// Stand スタンド
func (bi *BlackJackSwitchInteractor) Stand() string {
	return execAndPresent(bi.Game, bi.dp, bi.Game.PlayerStand)
}

// DoubleDown ダブルダウン
func (bi *BlackJackSwitchInteractor) DoubleDown() string {
	return execAndPresent(bi.Game, bi.dp, bi.Game.PlayerDoubleDown)
}

// ActionLog 棋譜を出力する
func (bi *BlackJackSwitchInteractor) ActionLog() string {
	return bi.dp.ActionLogOutput(bi.Game)
}

// RestoreBlackJackSwitchInteractor deserialises JSON into a BlackJackSwitchInteractor.
func RestoreBlackJackSwitchInteractor(data []byte, dp presenter.BlackJackSwitchPresenter) (*BlackJackSwitchInteractor, error) {
	return restoreAndBuild[domain.BlackJackSwitch](data, func(g *domain.BlackJackSwitch) *BlackJackSwitchInteractor {
		return &BlackJackSwitchInteractor{GameBase: GameBase[interfaces.BlackJackSwitchGame]{Game: g}, dp: dp}
	})
}
