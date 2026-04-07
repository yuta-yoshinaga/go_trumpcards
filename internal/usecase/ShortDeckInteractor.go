package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// ShortDeckInteractorIF ショートデックホールデムインタラクターインタフェース
type ShortDeckInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
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
	GameBase[interfaces.ShortDeckGame]
	op presenter.ShortDeckPresenter
	tournamentActions[interfaces.ShortDeckGame]
}

// NewShortDeckInteractor コンストラクタ
func NewShortDeckInteractor(o interfaces.ShortDeckGame, op presenter.ShortDeckPresenter) *ShortDeckInteractor {
	mustNotNil("ShortDeckInteractor", map[string]any{"o": o, "op": op})
	return &ShortDeckInteractor{
		GameBase:          GameBase[interfaces.ShortDeckGame]{Game: o},
		op:                op,
		tournamentActions: newTournamentActions[interfaces.ShortDeckGame](o, op),
	}
}

// Reset ゲーム初期化
func (oi *ShortDeckInteractor) Reset() string {
	return execAndPresent(oi.Game, oi.op, oi.Game.Reset)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (oi *ShortDeckInteractor) ResetWithConfig(cfg domain.ShortDeckConfig, profileData []byte) string {
	if err := cfg.Validate(); err != nil {
		return oi.op.Output(oi.Game, err)
	}
	if cfg.TableSize > 0 && cfg.TableSize != oi.Game.GetPlayerCnt() {
		oi.Game.Resize(domain.NewShortDeckPlayersForTable(cfg.TableSize))
	}
	oi.Game.SetConfig(cfg)
	err := oi.Game.Reset()
	if len(profileData) > 0 {
		_ = oi.Game.ImportProfile(profileData)
	}
	return oi.op.Output(oi.Game, err)
}

// Action プレイヤーアクション実行
func (oi *ShortDeckInteractor) Action(action int, amount int, humanPlayMs int) string {
	return execAndPresent(oi.Game, oi.op, func() error { return oi.Game.PlayerAction(action, amount, humanPlayMs) })
}

// GetConfig 現在の設定を取得
func (oi *ShortDeckInteractor) GetConfig() domain.ShortDeckConfig {
	return oi.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (oi *ShortDeckInteractor) ActionLog() string {
	return oi.op.ActionLogOutput(oi.Game)
}

// RestoreShortDeckInteractor deserialises JSON into a ShortDeckInteractor.
func RestoreShortDeckInteractor(data []byte, op presenter.ShortDeckPresenter) (*ShortDeckInteractor, error) {
	return restoreAndBuild[domain.ShortDeck](data, func(g *domain.ShortDeck) *ShortDeckInteractor {
		return &ShortDeckInteractor{GameBase: GameBase[interfaces.ShortDeckGame]{Game: g}, op: op}
	})
}
