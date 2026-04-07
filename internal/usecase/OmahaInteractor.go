package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// OmahaInteractorIF オマハホールデムインタラクターインタフェース
type OmahaInteractorIF interface {
	TournamentInteractorIF
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化 (profileData: JSONプロファイル、nilなら無視)
	ResetWithConfig(cfg domain.OmahaConfig, profileData []byte) string
	// Action プレイヤーアクション実行
	Action(action int, amount int, humanPlayMs int) string
	// GetConfig 現在の設定を取得
	GetConfig() domain.OmahaConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// OmahaInteractor オマハホールデムインタラクタークラス
type OmahaInteractor struct {
	GameBase[interfaces.OmahaGame]
	op presenter.OmahaPresenter
	tournamentActions[interfaces.OmahaGame]
}

// NewOmahaInteractor コンストラクタ
func NewOmahaInteractor(o interfaces.OmahaGame, op presenter.OmahaPresenter) *OmahaInteractor {
	mustNotNil("OmahaInteractor", map[string]any{"o": o, "op": op})
	return &OmahaInteractor{
		GameBase:          GameBase[interfaces.OmahaGame]{Game: o},
		op:                op,
		tournamentActions: newTournamentActions[interfaces.OmahaGame](o, op),
	}
}

// Reset ゲーム初期化
func (oi *OmahaInteractor) Reset() string {
	return execAndPresent(oi.Game, oi.op, oi.Game.Reset)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (oi *OmahaInteractor) ResetWithConfig(cfg domain.OmahaConfig, profileData []byte) string {
	if err := cfg.Validate(); err != nil {
		return oi.op.Output(oi.Game, err)
	}
	if cfg.TableSize > 0 && cfg.TableSize != oi.Game.GetPlayerCnt() {
		oi.Game.Resize(domain.NewOmahaPlayersForTable(cfg.TableSize))
	}
	oi.Game.SetConfig(cfg)
	err := oi.Game.Reset()
	if len(profileData) > 0 {
		_ = oi.Game.ImportProfile(profileData)
	}
	return oi.op.Output(oi.Game, err)
}

// Action プレイヤーアクション実行
func (oi *OmahaInteractor) Action(action int, amount int, humanPlayMs int) string {
	return execAndPresent(oi.Game, oi.op, func() error { return oi.Game.PlayerAction(action, amount, humanPlayMs) })
}

// GetConfig 現在の設定を取得
func (oi *OmahaInteractor) GetConfig() domain.OmahaConfig {
	return oi.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (oi *OmahaInteractor) ActionLog() string {
	return oi.op.ActionLogOutput(oi.Game)
}

// RestoreOmahaInteractor deserialises JSON into an OmahaInteractor.
func RestoreOmahaInteractor(data []byte, op presenter.OmahaPresenter) (*OmahaInteractor, error) {
	o, err := restoreGame[domain.Omaha](data)
	if err != nil {
		return nil, err
	}
	return &OmahaInteractor{GameBase: GameBase[interfaces.OmahaGame]{Game: o}, op: op, tournamentActions: newTournamentActions[interfaces.OmahaGame](o, op)}, nil
}
