package usecase

import (
	"encoding/json"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// PokerInteractorIF ポーカーインタラクターインタフェース
type PokerInteractorIF interface {
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
	p  interfaces.PokerGame
	pp presenter.PokerPresenter
}

// NewPokerInteractor コンストラクタ
func NewPokerInteractor(p interfaces.PokerGame, pp presenter.PokerPresenter) *PokerInteractor {
	mustNotNil("PokerInteractor", map[string]any{"p": p, "pp": pp})
	return &PokerInteractor{
		p:  p,
		pp: pp,
	}
}

// Reset ゲーム初期化
func (pi *PokerInteractor) Reset() string {
	return execAndPresent(pi.p, pi.pp, pi.p.Reset)
}

// GetConfig 現在の設定を取得
func (pi *PokerInteractor) GetConfig() domain.PokerConfig {
	return pi.p.GetConfig()
}

// ResetWithConfig 設定を変更してゲーム初期化
func (pi *PokerInteractor) ResetWithConfig(cfg domain.PokerConfig, profileData []byte) string {
	if err := cfg.Validate(); err != nil {
		return pi.pp.Output(pi.p, err)
	}
	pi.p.SetConfig(cfg)
	err := pi.p.Reset()
	if len(profileData) > 0 {
		_ = pi.p.ImportProfile(profileData)
	}
	return pi.pp.Output(pi.p, err)
}

// Action プレイヤーアクション実行
func (pi *PokerInteractor) Action(action int, amount int, humanPlayMs int) string {
	return execAndPresent(pi.p, pi.pp, func() error { return pi.p.PlayerAction(action, amount, humanPlayMs) })
}

// Exchange カード交換
func (pi *PokerInteractor) Exchange(indices []int) string {
	return execAndPresent(pi.p, pi.pp, func() error { return pi.p.PlayerExchange(indices) })
}

// Stand カード交換なし
func (pi *PokerInteractor) Stand() string {
	return execAndPresent(pi.p, pi.pp, pi.p.PlayerStand)
}

// ActionLog 棋譜を出力する
func (pi *PokerInteractor) ActionLog() string {
	return pi.pp.ActionLogOutput(pi.p)
}

// Odds ドローオッズ計算
func (pi *PokerInteractor) Odds(indices []int) string {
	odds, err := pi.p.CalcDrawOdds(indices)
	if err != nil {
		return pi.pp.Output(pi.p, err)
	}
	return pi.pp.OutputWithOdds(pi.p, nil, odds)
}

// Snapshot serialises the game state to JSON for KV persistence.
func (pi *PokerInteractor) Snapshot() ([]byte, error) {
	return json.Marshal(pi.p)
}

// RestorePokerInteractor deserialises JSON into a PokerInteractor.
func RestorePokerInteractor(data []byte, pp presenter.PokerPresenter) (*PokerInteractor, error) {
	p, err := restoreGame[domain.Poker](data)
	if err != nil {
		return nil, err
	}
	return &PokerInteractor{p: p, pp: pp}, nil
}
