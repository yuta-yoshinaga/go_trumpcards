//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// PineappleInteractorIF パイナップルポーカーインタラクターインタフェース
type PineappleInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	TournamentInteractorIF
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化 (profileData: JSONプロファイル、nilなら無視)
	ResetWithConfig(cfg domain.PineappleConfig, profileData []byte) string
	// Action プレイヤーアクション実行
	Action(action int, amount int, humanPlayMs int) string
	// Discard ディスカード実行 (1枚)
	Discard(cardIdx int) string
	// DiscardMany ディスカード実行 (複数枚を一括: Irish Poker の2枚捨て等)
	DiscardMany(cardIdxs []int) string
	// GetConfig 現在の設定を取得
	GetConfig() domain.PineappleConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// PineappleInteractor パイナップルポーカーインタラクタークラス
type PineappleInteractor struct {
	GameBase[interfaces.PineappleGame]
	pp presenter.PineapplePresenter
	tournamentActions[interfaces.PineappleGame]
}

// NewPineappleInteractor コンストラクタ
func NewPineappleInteractor(p interfaces.PineappleGame, pp presenter.PineapplePresenter) *PineappleInteractor {
	mustNotNil("PineappleInteractor", map[string]any{"p": p, "pp": pp})
	return &PineappleInteractor{
		GameBase:          GameBase[interfaces.PineappleGame]{Game: p},
		pp:                pp,
		tournamentActions: newTournamentActions[interfaces.PineappleGame](p, pp),
	}
}

// Reset ゲーム初期化
func (pi *PineappleInteractor) Reset() string {
	return execAndPresent(pi.Game, pi.pp, pi.Game.Reset)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (pi *PineappleInteractor) ResetWithConfig(cfg domain.PineappleConfig, profileData []byte) string {
	if err := cfg.Validate(); err != nil {
		return pi.pp.Output(pi.Game, err)
	}
	// テーブルサイズ変更時はプレイヤーを再構築
	if cfg.TableSize > 0 && cfg.TableSize != pi.Game.GetPlayerCnt() {
		pi.Game.Resize(domain.NewPineapplePlayersForTable(cfg.TableSize))
	}
	pi.Game.SetConfig(cfg)
	err := pi.Game.Reset()
	if len(profileData) > 0 {
		_ = pi.Game.ImportProfile(profileData)
	}
	return pi.pp.Output(pi.Game, err)
}

// Action プレイヤーアクション実行
func (pi *PineappleInteractor) Action(action int, amount int, humanPlayMs int) string {
	return execAndPresent(pi.Game, pi.pp, func() error { return pi.Game.PlayerAction(action, amount, humanPlayMs) })
}

// Discard ディスカード実行
func (pi *PineappleInteractor) Discard(cardIdx int) string {
	return execAndPresent(pi.Game, pi.pp, func() error { return pi.Game.DiscardCard(cardIdx) })
}

// DiscardMany ディスカード実行 (複数枚を一括)
func (pi *PineappleInteractor) DiscardMany(cardIdxs []int) string {
	return execAndPresent(pi.Game, pi.pp, func() error { return pi.Game.DiscardCards(cardIdxs) })
}

// GetConfig 現在の設定を取得
func (pi *PineappleInteractor) GetConfig() domain.PineappleConfig {
	return pi.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (pi *PineappleInteractor) ActionLog() string {
	return pi.pp.ActionLogOutput(pi.Game)
}

// RestorePineappleInteractor deserialises JSON into a PineappleInteractor.
func RestorePineappleInteractor(data []byte, pp presenter.PineapplePresenter) (*PineappleInteractor, error) {
	return restoreAndBuild[domain.Pineapple](data, func(g *domain.Pineapple) *PineappleInteractor {
		return &PineappleInteractor{GameBase: GameBase[interfaces.PineappleGame]{Game: g}, pp: pp}
	})
}
