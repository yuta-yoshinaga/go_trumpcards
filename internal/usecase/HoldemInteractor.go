package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// HoldemInteractorIF テキサスホールデムインタラクターインタフェース
type HoldemInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	TournamentInteractorIF
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化 (profileData: JSONプロファイル、nilなら無視)
	ResetWithConfig(cfg domain.HoldemConfig, profileData []byte) string
	// Action プレイヤーアクション実行
	Action(action int, amount int, humanPlayMs int) string
	// GetConfig 現在の設定を取得
	GetConfig() domain.HoldemConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// HoldemInteractor テキサスホールデムインタラクタークラス
type HoldemInteractor struct {
	GameBase[interfaces.HoldemGame]
	hp presenter.HoldemPresenter
	tournamentActions[interfaces.HoldemGame]
}

// NewHoldemInteractor コンストラクタ
func NewHoldemInteractor(h interfaces.HoldemGame, hp presenter.HoldemPresenter) *HoldemInteractor {
	mustNotNil("HoldemInteractor", map[string]any{"h": h, "hp": hp})
	return &HoldemInteractor{
		GameBase:          GameBase[interfaces.HoldemGame]{Game: h},
		hp:                hp,
		tournamentActions: newTournamentActions[interfaces.HoldemGame](h, hp),
	}
}

// Reset ゲーム初期化
func (hi *HoldemInteractor) Reset() string {
	return execAndPresent(hi.Game, hi.hp, hi.Game.Reset)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (hi *HoldemInteractor) ResetWithConfig(cfg domain.HoldemConfig, profileData []byte) string {
	if err := cfg.Validate(); err != nil {
		return hi.hp.Output(hi.Game, err)
	}
	// テーブルサイズ変更時はプレイヤーを再構築
	if cfg.TableSize > 0 && cfg.TableSize != hi.Game.GetPlayerCnt() {
		hi.Game.Resize(domain.NewPlayersForTable(cfg.TableSize))
	}
	hi.Game.SetConfig(cfg)
	err := hi.Game.Reset()
	if len(profileData) > 0 {
		_ = hi.Game.ImportProfile(profileData)
	}
	return hi.hp.Output(hi.Game, err)
}

// Action プレイヤーアクション実行
func (hi *HoldemInteractor) Action(action int, amount int, humanPlayMs int) string {
	return execAndPresent(hi.Game, hi.hp, func() error { return hi.Game.PlayerAction(action, amount, humanPlayMs) })
}

// GetConfig 現在の設定を取得
func (hi *HoldemInteractor) GetConfig() domain.HoldemConfig {
	return hi.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (hi *HoldemInteractor) ActionLog() string {
	return hi.hp.ActionLogOutput(hi.Game)
}

// RestoreHoldemInteractor deserialises JSON into a HoldemInteractor.
func RestoreHoldemInteractor(data []byte, hp presenter.HoldemPresenter) (*HoldemInteractor, error) {
	return restoreAndBuild[domain.Holdem](data, func(g *domain.Holdem) *HoldemInteractor {
		return &HoldemInteractor{GameBase: GameBase[interfaces.HoldemGame]{Game: g}, hp: hp}
	})
}
