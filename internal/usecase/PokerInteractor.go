package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// PokerInteractorIF ポーカーインタラクターインタフェース
type PokerInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化 (profileData: JSONプロファイル、nilなら無視)
	ResetWithConfig(cfg domain.PokerConfig, profileData []byte) string
	// GetConfig 現在の設定を取得
	GetConfig() domain.PokerConfig
	// Action プレイヤーアクション実行
	Action(action int, amount int, humanPlayMs int) string
	// Exchange カード交換
	Exchange(indices []int) string
	// Stand カード交換なし
	Stand() string
	// Odds ドローオッズ計算
	Odds(indices []int) string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// PokerInteractor ポーカーインタラクタークラス
type PokerInteractor struct {
	GameBase[interfaces.PokerGame]
	pp presenter.PokerPresenter
}

// NewPokerInteractor コンストラクタ
func NewPokerInteractor(p interfaces.PokerGame, pp presenter.PokerPresenter) *PokerInteractor {
	mustNotNil("PokerInteractor", map[string]any{"p": p, "pp": pp})
	return &PokerInteractor{
		GameBase: GameBase[interfaces.PokerGame]{Game: p},
		pp:       pp,
	}
}

// Reset ゲーム初期化
func (pi *PokerInteractor) Reset() string {
	return execAndPresent(pi.Game, pi.pp, pi.Game.Reset)
}

// GetConfig 現在の設定を取得
func (pi *PokerInteractor) GetConfig() domain.PokerConfig {
	return pi.Game.GetConfig()
}

// ResetWithConfig 設定を変更してゲーム初期化
func (pi *PokerInteractor) ResetWithConfig(cfg domain.PokerConfig, profileData []byte) string {
	if err := cfg.Validate(); err != nil {
		return pi.pp.Output(pi.Game, err)
	}
	pi.Game.SetConfig(cfg)
	err := pi.Game.Reset()
	if len(profileData) > 0 {
		_ = pi.Game.ImportProfile(profileData)
	}
	return pi.pp.Output(pi.Game, err)
}

// Action プレイヤーアクション実行
func (pi *PokerInteractor) Action(action int, amount int, humanPlayMs int) string {
	return execAndPresent(pi.Game, pi.pp, func() error { return pi.Game.PlayerAction(action, amount, humanPlayMs) })
}

// Exchange カード交換
func (pi *PokerInteractor) Exchange(indices []int) string {
	return execAndPresent(pi.Game, pi.pp, func() error { return pi.Game.PlayerExchange(indices) })
}

// Stand カード交換なし
func (pi *PokerInteractor) Stand() string {
	return execAndPresent(pi.Game, pi.pp, pi.Game.PlayerStand)
}

// ActionLog 棋譜を出力する
func (pi *PokerInteractor) ActionLog() string {
	return pi.pp.ActionLogOutput(pi.Game)
}

// Odds ドローオッズ計算
func (pi *PokerInteractor) Odds(indices []int) string {
	odds, err := pi.Game.CalcDrawOdds(indices)
	if err != nil {
		return pi.pp.Output(pi.Game, err)
	}
	return pi.pp.OutputWithOdds(pi.Game, nil, odds)
}

// RestorePokerInteractor deserialises JSON into a PokerInteractor.
func RestorePokerInteractor(data []byte, pp presenter.PokerPresenter) (*PokerInteractor, error) {
	return restoreAndBuild[domain.Poker](data, func(g *domain.Poker) *PokerInteractor {
		return &PokerInteractor{GameBase: GameBase[interfaces.PokerGame]{Game: g}, pp: pp}
	})
}
