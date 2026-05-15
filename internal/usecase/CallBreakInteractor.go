package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// CallBreakInteractorIF Call Break インタラクターインタフェース
type CallBreakInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.CallBreakConfig) string
	// Bid ビッドを宣言
	Bid(bid int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.CallBreakConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// CallBreakInteractor Call Break インタラクタークラス
type CallBreakInteractor struct {
	GameBase[interfaces.CallBreakGame]
	sp presenter.CallBreakPresenter
}

// NewCallBreakInteractor コンストラクタ
func NewCallBreakInteractor(cb interfaces.CallBreakGame, sp presenter.CallBreakPresenter) *CallBreakInteractor {
	mustNotNil("CallBreakInteractor", map[string]any{"cb": cb, "sp": sp})
	return &CallBreakInteractor{GameBase: GameBase[interfaces.CallBreakGame]{Game: cb}, sp: sp}
}

// Reset ゲーム初期化
func (ci *CallBreakInteractor) Reset() string {
	ci.Game.Reset()
	ci.runCpuBids()
	if ci.Game.GetPhase() == domain.CallBreakPhasePlay {
		ci.runCpuTurns()
	}
	return ci.sp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *CallBreakInteractor) ResetWithConfig(cfg domain.CallBreakConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.sp, cfg, ci.Game.SetConfig, ci.Reset)
}

// Bid ビッドを宣言
func (ci *CallBreakInteractor) Bid(bid int) string {
	if out, blocked := guardGameEnd(ci.Game, ci.sp); blocked {
		return out
	}
	err := ci.Game.PlayerBid(bid)
	if err != nil {
		return ci.sp.Output(ci.Game, err)
	}
	ci.runCpuBids()
	if ci.Game.GetPhase() == domain.CallBreakPhasePlay {
		ci.runCpuTurns()
	}
	return ci.sp.Output(ci.Game, nil)
}

// Play カードをプレイ
func (ci *CallBreakInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.sp); blocked {
		return out
	}
	err := ci.Game.PlayerPlay(cardIndex)
	if err != nil {
		return ci.sp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.sp.Output(ci.Game, nil)
}

// NextTrick 次のトリックへ進む
func (ci *CallBreakInteractor) NextTrick() string {
	ci.Game.NextTrick()
	ci.runCpuTurns()
	return ci.sp.Output(ci.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (ci *CallBreakInteractor) NextRound() string {
	ci.Game.ScoreRound()
	if out, blocked := guardGameEnd(ci.Game, ci.sp); blocked {
		return out
	}
	ci.Game.NextRound()
	ci.runCpuBids()
	return ci.sp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を取得
func (ci *CallBreakInteractor) GetConfig() domain.CallBreakConfig {
	return ci.Game.GetConfig()
}

// Hint ヒント取得
func (ci *CallBreakInteractor) Hint() string {
	return ci.sp.HintOutput(ci.Game)
}

// ActionLog 棋譜を出力する
func (ci *CallBreakInteractor) ActionLog() string {
	return ci.sp.ActionLogOutput(ci.Game)
}

// runCpuBids ヒューマンのビッド番もしくはビッドフェーズ終了まで CPU ビッドを実行
func (ci *CallBreakInteractor) runCpuBids() {
	runCpuBidsLoop(ci.Game, domain.CallBreakPhaseBid)
}

// runCpuTurns ヒューマンの手番またはトリック / ラウンド終了まで CPU ターンを実行
func (ci *CallBreakInteractor) runCpuTurns() {
	runCpuTurnsLoop(ci.Game, trickPhases[domain.CallBreakPhase]{
		play:     domain.CallBreakPhasePlay,
		trickEnd: domain.CallBreakPhaseTrickEnd,
		roundEnd: domain.CallBreakPhaseRoundEnd,
		gameEnd:  domain.CallBreakPhaseGameEnd,
	})
}

// RestoreCallBreakInteractor deserialises JSON into a CallBreakInteractor.
func RestoreCallBreakInteractor(data []byte, sp presenter.CallBreakPresenter) (*CallBreakInteractor, error) {
	return restoreAndBuild[domain.CallBreak](data, func(g *domain.CallBreak) *CallBreakInteractor {
		return &CallBreakInteractor{GameBase: GameBase[interfaces.CallBreakGame]{Game: g}, sp: sp}
	})
}
