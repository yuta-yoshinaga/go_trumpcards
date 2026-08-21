//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// FollowTheQueenInteractorIF フォロー・ザ・クイーンインタラクターインタフェース
type FollowTheQueenInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	TournamentInteractorIF
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.FollowTheQueenConfig, profileData []byte) string
	// Action プレイヤーアクション実行
	Action(action int, amount int, humanPlayMs int) string
	// GetConfig 現在の設定を取得
	GetConfig() domain.FollowTheQueenConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
	// Hint ヒントを出力する
	Hint() string
}

// FollowTheQueenInteractor フォロー・ザ・クイーンインタラクタークラス
type FollowTheQueenInteractor struct {
	GameBase[interfaces.FollowTheQueenGame]
	sp presenter.FollowTheQueenPresenter
	tournamentActions[interfaces.FollowTheQueenGame]
}

// NewFollowTheQueenInteractor コンストラクタ
func NewFollowTheQueenInteractor(s interfaces.FollowTheQueenGame, sp presenter.FollowTheQueenPresenter) *FollowTheQueenInteractor {
	mustNotNil("FollowTheQueenInteractor", map[string]any{"s": s, "sp": sp})
	return &FollowTheQueenInteractor{
		GameBase:          GameBase[interfaces.FollowTheQueenGame]{Game: s},
		sp:                sp,
		tournamentActions: newTournamentActions[interfaces.FollowTheQueenGame](s, sp),
	}
}

// Reset ゲーム初期化
func (si *FollowTheQueenInteractor) Reset() string {
	return execAndPresent(si.Game, si.sp, si.Game.Reset)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (si *FollowTheQueenInteractor) ResetWithConfig(cfg domain.FollowTheQueenConfig, profileData []byte) string {
	if err := cfg.Validate(); err != nil {
		return si.sp.Output(si.Game, err)
	}
	if cfg.TableSize > 0 && cfg.TableSize != si.Game.GetPlayerCnt() {
		si.Game.Resize(domain.NewFollowTheQueenPlayersForTable(cfg.TableSize))
	}
	si.Game.SetConfig(cfg)
	err := si.Game.Reset()
	if len(profileData) > 0 {
		_ = si.Game.ImportProfile(profileData)
	}
	return si.sp.Output(si.Game, err)
}

// Action プレイヤーアクション実行
func (si *FollowTheQueenInteractor) Action(action int, amount int, humanPlayMs int) string {
	return execAndPresent(si.Game, si.sp, func() error { return si.Game.PlayerAction(action, amount, humanPlayMs) })
}

// GetConfig 現在の設定を取得
func (si *FollowTheQueenInteractor) GetConfig() domain.FollowTheQueenConfig {
	return si.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (si *FollowTheQueenInteractor) ActionLog() string {
	return si.sp.ActionLogOutput(si.Game)
}

// Hint ヒントを出力する
func (si *FollowTheQueenInteractor) Hint() string {
	return si.sp.HintOutput(si.Game)
}

// RestoreFollowTheQueenInteractor deserialises JSON into a FollowTheQueenInteractor.
func RestoreFollowTheQueenInteractor(data []byte, sp presenter.FollowTheQueenPresenter) (*FollowTheQueenInteractor, error) {
	return restoreAndBuild[domain.FollowTheQueen](data, func(g *domain.FollowTheQueen) *FollowTheQueenInteractor {
		return &FollowTheQueenInteractor{GameBase: GameBase[interfaces.FollowTheQueenGame]{Game: g}, sp: sp}
	})
}
