//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SevenCardStudInteractorIF セブンカードスタッドインタラクターインタフェース
type SevenCardStudInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	TournamentInteractorIF
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.SevenCardStudConfig, profileData []byte) string
	// Action プレイヤーアクション実行
	Action(action int, amount int, humanPlayMs int) string
	// GetConfig 現在の設定を取得
	GetConfig() domain.SevenCardStudConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
	// Hint ヒントを出力する
	Hint() string
}

// SevenCardStudInteractor セブンカードスタッドインタラクタークラス
type SevenCardStudInteractor struct {
	GameBase[interfaces.SevenCardStudGame]
	sp presenter.SevenCardStudPresenter
	tournamentActions[interfaces.SevenCardStudGame]
}

// NewSevenCardStudInteractor コンストラクタ
func NewSevenCardStudInteractor(s interfaces.SevenCardStudGame, sp presenter.SevenCardStudPresenter) *SevenCardStudInteractor {
	mustNotNil("SevenCardStudInteractor", map[string]any{"s": s, "sp": sp})
	return &SevenCardStudInteractor{
		GameBase:          GameBase[interfaces.SevenCardStudGame]{Game: s},
		sp:                sp,
		tournamentActions: newTournamentActions[interfaces.SevenCardStudGame](s, sp),
	}
}

// Reset ゲーム初期化
func (si *SevenCardStudInteractor) Reset() string {
	return execAndPresent(si.Game, si.sp, si.Game.Reset)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (si *SevenCardStudInteractor) ResetWithConfig(cfg domain.SevenCardStudConfig, profileData []byte) string {
	if err := cfg.Validate(); err != nil {
		return si.sp.Output(si.Game, err)
	}
	if cfg.TableSize > 0 && cfg.TableSize != si.Game.GetPlayerCnt() {
		si.Game.Resize(domain.NewSevenCardStudPlayersForTable(cfg.TableSize))
	}
	si.Game.SetConfig(cfg)
	err := si.Game.Reset()
	if len(profileData) > 0 {
		_ = si.Game.ImportProfile(profileData)
	}
	return si.sp.Output(si.Game, err)
}

// Action プレイヤーアクション実行
func (si *SevenCardStudInteractor) Action(action int, amount int, humanPlayMs int) string {
	return execAndPresent(si.Game, si.sp, func() error { return si.Game.PlayerAction(action, amount, humanPlayMs) })
}

// GetConfig 現在の設定を取得
func (si *SevenCardStudInteractor) GetConfig() domain.SevenCardStudConfig {
	return si.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (si *SevenCardStudInteractor) ActionLog() string {
	return si.sp.ActionLogOutput(si.Game)
}

// Hint ヒントを出力する
func (si *SevenCardStudInteractor) Hint() string {
	return si.sp.HintOutput(si.Game)
}

// RestoreSevenCardStudInteractor deserialises JSON into a SevenCardStudInteractor.
func RestoreSevenCardStudInteractor(data []byte, sp presenter.SevenCardStudPresenter) (*SevenCardStudInteractor, error) {
	return restoreAndBuild[domain.SevenCardStud](data, func(g *domain.SevenCardStud) *SevenCardStudInteractor {
		return &SevenCardStudInteractor{GameBase: GameBase[interfaces.SevenCardStudGame]{Game: g}, sp: sp}
	})
}
