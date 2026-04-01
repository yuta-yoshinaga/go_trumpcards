package usecase

import (
	"encoding/json"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// ShortDeckInteractorIF ショートデックホールデムインタラクターインタフェース
type ShortDeckInteractorIF interface {
	TournamentInteractorIF
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化 (profileData: JSONプロファイル、nilなら無視)
	ResetWithConfig(cfg domain.ShortDeckConfig, profileData []byte) string
	// Action プレイヤーアクション実行
	Action(action int, amount int, humanPlayMs int) string
	// GetConfig 現在の設定を取得
	GetConfig() domain.ShortDeckConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// ShortDeckInteractor ショートデックホールデムインタラクタークラス
type ShortDeckInteractor struct {
	o  interfaces.ShortDeckGame
	op presenter.ShortDeckPresenter
	tournamentActions[interfaces.ShortDeckGame]
}

// NewShortDeckInteractor コンストラクタ
func NewShortDeckInteractor(o interfaces.ShortDeckGame, op presenter.ShortDeckPresenter) *ShortDeckInteractor {
	mustNotNil("ShortDeckInteractor", map[string]any{"o": o, "op": op})
	return &ShortDeckInteractor{
		o:                 o,
		op:                op,
		tournamentActions: newTournamentActions[interfaces.ShortDeckGame](o, op),
	}
}

// Reset ゲーム初期化
func (oi *ShortDeckInteractor) Reset() string {
	return execAndPresent(oi.o, oi.op, oi.o.Reset)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (oi *ShortDeckInteractor) ResetWithConfig(cfg domain.ShortDeckConfig, profileData []byte) string {
	if err := cfg.Validate(); err != nil {
		return oi.op.Output(oi.o, err)
	}
	if cfg.TableSize > 0 && cfg.TableSize != oi.o.GetPlayerCnt() {
		oi.o.Resize(domain.NewShortDeckPlayersForTable(cfg.TableSize))
	}
	oi.o.SetConfig(cfg)
	err := oi.o.Reset()
	if len(profileData) > 0 {
		_ = oi.o.ImportProfile(profileData)
	}
	return oi.op.Output(oi.o, err)
}

// Action プレイヤーアクション実行
func (oi *ShortDeckInteractor) Action(action int, amount int, humanPlayMs int) string {
	return execAndPresent(oi.o, oi.op, func() error { return oi.o.PlayerAction(action, amount, humanPlayMs) })
}

// GetConfig 現在の設定を取得
func (oi *ShortDeckInteractor) GetConfig() domain.ShortDeckConfig {
	return oi.o.GetConfig()
}

// ActionLog 棋譜を出力する
func (oi *ShortDeckInteractor) ActionLog() string {
	return oi.op.ActionLogOutput(oi.o)
}

// Snapshot serialises the game state to JSON for KV persistence.
func (oi *ShortDeckInteractor) Snapshot() ([]byte, error) {
	return json.Marshal(oi.o)
}

// RestoreShortDeckInteractor deserialises JSON into a ShortDeckInteractor.
func RestoreShortDeckInteractor(data []byte, op presenter.ShortDeckPresenter) (*ShortDeckInteractor, error) {
	sd, err := restoreGame[domain.ShortDeck](data)
	if err != nil {
		return nil, err
	}
	return &ShortDeckInteractor{o: sd, op: op}, nil
}
