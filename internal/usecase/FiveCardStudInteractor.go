//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// FiveCardStudInteractorIF ファイブカードスタッドインタラクターインタフェース
type FiveCardStudInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	TournamentInteractorIF
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.FiveCardStudConfig, profileData []byte) string
	// Action プレイヤーアクション実行
	Action(action int, amount int, humanPlayMs int) string
	// GetConfig 現在の設定を取得
	GetConfig() domain.FiveCardStudConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// FiveCardStudInteractor ファイブカードスタッドインタラクタークラス
type FiveCardStudInteractor struct {
	GameBase[interfaces.FiveCardStudGame]
	sp presenter.FiveCardStudPresenter
	tournamentActions[interfaces.FiveCardStudGame]
}

// NewFiveCardStudInteractor コンストラクタ
func NewFiveCardStudInteractor(s interfaces.FiveCardStudGame, sp presenter.FiveCardStudPresenter) *FiveCardStudInteractor {
	mustNotNil("FiveCardStudInteractor", map[string]any{"s": s, "sp": sp})
	return &FiveCardStudInteractor{
		GameBase:          GameBase[interfaces.FiveCardStudGame]{Game: s},
		sp:                sp,
		tournamentActions: newTournamentActions[interfaces.FiveCardStudGame](s, sp),
	}
}

// Reset ゲーム初期化
func (si *FiveCardStudInteractor) Reset() string {
	return execAndPresent(si.Game, si.sp, si.Game.Reset)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (si *FiveCardStudInteractor) ResetWithConfig(cfg domain.FiveCardStudConfig, profileData []byte) string {
	if err := cfg.Validate(); err != nil {
		return si.sp.Output(si.Game, err)
	}
	if cfg.TableSize > 0 && cfg.TableSize != si.Game.GetPlayerCnt() {
		si.Game.Resize(domain.NewFiveCardStudPlayersForTable(cfg.TableSize))
	}
	si.Game.SetConfig(cfg)
	err := si.Game.Reset()
	if len(profileData) > 0 {
		_ = si.Game.ImportProfile(profileData)
	}
	return si.sp.Output(si.Game, err)
}

// Action プレイヤーアクション実行
func (si *FiveCardStudInteractor) Action(action int, amount int, humanPlayMs int) string {
	return execAndPresent(si.Game, si.sp, func() error { return si.Game.PlayerAction(action, amount, humanPlayMs) })
}

// GetConfig 現在の設定を取得
func (si *FiveCardStudInteractor) GetConfig() domain.FiveCardStudConfig {
	return si.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (si *FiveCardStudInteractor) ActionLog() string {
	return si.sp.ActionLogOutput(si.Game)
}

// RestoreFiveCardStudInteractor deserialises JSON into a FiveCardStudInteractor.
func RestoreFiveCardStudInteractor(data []byte, sp presenter.FiveCardStudPresenter) (*FiveCardStudInteractor, error) {
	return restoreAndBuild[domain.FiveCardStud](data, func(g *domain.FiveCardStud) *FiveCardStudInteractor {
		return &FiveCardStudInteractor{GameBase: GameBase[interfaces.FiveCardStudGame]{Game: g}, sp: sp}
	})
}
