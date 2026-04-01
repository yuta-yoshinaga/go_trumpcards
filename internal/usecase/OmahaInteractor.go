package usecase

import (
	"encoding/json"

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
	o  interfaces.OmahaGame
	op presenter.OmahaPresenter
	tournamentActions[interfaces.OmahaGame]
}

// NewOmahaInteractor コンストラクタ
func NewOmahaInteractor(o interfaces.OmahaGame, op presenter.OmahaPresenter) *OmahaInteractor {
	mustNotNil("OmahaInteractor", map[string]any{"o": o, "op": op})
	return &OmahaInteractor{
		o:                 o,
		op:                op,
		tournamentActions: newTournamentActions[interfaces.OmahaGame](o, op),
	}
}

// Reset ゲーム初期化
func (oi *OmahaInteractor) Reset() string {
	return execAndPresent(oi.o, oi.op, oi.o.Reset)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (oi *OmahaInteractor) ResetWithConfig(cfg domain.OmahaConfig, profileData []byte) string {
	if err := cfg.Validate(); err != nil {
		return oi.op.Output(oi.o, err)
	}
	if cfg.TableSize > 0 && cfg.TableSize != oi.o.GetPlayerCnt() {
		oi.o.Resize(domain.NewOmahaPlayersForTable(cfg.TableSize))
	}
	oi.o.SetConfig(cfg)
	err := oi.o.Reset()
	if len(profileData) > 0 {
		_ = oi.o.ImportProfile(profileData)
	}
	return oi.op.Output(oi.o, err)
}

// Action プレイヤーアクション実行
func (oi *OmahaInteractor) Action(action int, amount int, humanPlayMs int) string {
	return execAndPresent(oi.o, oi.op, func() error { return oi.o.PlayerAction(action, amount, humanPlayMs) })
}

// GetConfig 現在の設定を取得
func (oi *OmahaInteractor) GetConfig() domain.OmahaConfig {
	return oi.o.GetConfig()
}

// ActionLog 棋譜を出力する
func (oi *OmahaInteractor) ActionLog() string {
	return oi.op.ActionLogOutput(oi.o)
}

// Snapshot serialises the game state to JSON for KV persistence.
func (oi *OmahaInteractor) Snapshot() ([]byte, error) {
	return json.Marshal(oi.o)
}

// RestoreOmahaInteractor deserialises JSON into an OmahaInteractor.
func RestoreOmahaInteractor(data []byte, op presenter.OmahaPresenter) (*OmahaInteractor, error) {
	o, err := restoreGame[domain.Omaha](data)
	if err != nil {
		return nil, err
	}
	return &OmahaInteractor{o: o, op: op, tournamentActions: newTournamentActions[interfaces.OmahaGame](o, op)}, nil
}
