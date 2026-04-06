package usecase

import (
	"encoding/json"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SevenCardStudInteractorIF セブンカードスタッドインタラクターインタフェース
type SevenCardStudInteractorIF interface {
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
}

// SevenCardStudInteractor セブンカードスタッドインタラクタークラス
type SevenCardStudInteractor struct {
	s  interfaces.SevenCardStudGame
	sp presenter.SevenCardStudPresenter
	tournamentActions[interfaces.SevenCardStudGame]
}

// NewSevenCardStudInteractor コンストラクタ
func NewSevenCardStudInteractor(s interfaces.SevenCardStudGame, sp presenter.SevenCardStudPresenter) *SevenCardStudInteractor {
	mustNotNil("SevenCardStudInteractor", map[string]any{"s": s, "sp": sp})
	return &SevenCardStudInteractor{
		s:                 s,
		sp:                sp,
		tournamentActions: newTournamentActions[interfaces.SevenCardStudGame](s, sp),
	}
}

// Reset ゲーム初期化
func (si *SevenCardStudInteractor) Reset() string {
	return execAndPresent(si.s, si.sp, si.s.Reset)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (si *SevenCardStudInteractor) ResetWithConfig(cfg domain.SevenCardStudConfig, profileData []byte) string {
	if err := cfg.Validate(); err != nil {
		return si.sp.Output(si.s, err)
	}
	if cfg.TableSize > 0 && cfg.TableSize != si.s.GetPlayerCnt() {
		si.s.Resize(domain.NewSevenCardStudPlayersForTable(cfg.TableSize))
	}
	si.s.SetConfig(cfg)
	err := si.s.Reset()
	if len(profileData) > 0 {
		_ = si.s.ImportProfile(profileData)
	}
	return si.sp.Output(si.s, err)
}

// Action プレイヤーアクション実行
func (si *SevenCardStudInteractor) Action(action int, amount int, humanPlayMs int) string {
	return execAndPresent(si.s, si.sp, func() error { return si.s.PlayerAction(action, amount, humanPlayMs) })
}

// GetConfig 現在の設定を取得
func (si *SevenCardStudInteractor) GetConfig() domain.SevenCardStudConfig {
	return si.s.GetConfig()
}

// ActionLog 棋譜を出力する
func (si *SevenCardStudInteractor) ActionLog() string {
	return si.sp.ActionLogOutput(si.s)
}

// Snapshot serialises the game state to JSON for KV persistence.
func (si *SevenCardStudInteractor) Snapshot() ([]byte, error) {
	return json.Marshal(si.s)
}

// RestoreSevenCardStudInteractor deserialises JSON into a SevenCardStudInteractor.
func RestoreSevenCardStudInteractor(data []byte, sp presenter.SevenCardStudPresenter) (*SevenCardStudInteractor, error) {
	s, err := restoreGame[domain.SevenCardStud](data)
	if err != nil {
		return nil, err
	}
	return &SevenCardStudInteractor{s: s, sp: sp, tournamentActions: newTournamentActions[interfaces.SevenCardStudGame](s, sp)}, nil
}
