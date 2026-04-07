package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// OhHellInteractorIF オー・ヘルインタラクターインタフェース
type OhHellInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.OhHellConfig) string
	// Bid ビッドを宣言
	Bid(bid int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.OhHellConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// OhHellInteractor オー・ヘルインタラクタークラス
type OhHellInteractor struct {
	GameBase[interfaces.OhHellGame]
	op presenter.OhHellPresenter
}

// NewOhHellInteractor コンストラクタ
func NewOhHellInteractor(o interfaces.OhHellGame, op presenter.OhHellPresenter) *OhHellInteractor {
	mustNotNil("OhHellInteractor", map[string]any{"o": o, "op": op})
	return &OhHellInteractor{GameBase: GameBase[interfaces.OhHellGame]{Game: o}, op: op}
}

// Reset ゲーム初期化
func (oi *OhHellInteractor) Reset() string {
	oi.Game.Reset()
	oi.runCpuBids()
	if oi.Game.GetPhase() == domain.OhHellPhasePlay {
		oi.runCpuTurns()
	}
	return oi.op.Output(oi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (oi *OhHellInteractor) ResetWithConfig(cfg domain.OhHellConfig) string {
	return resetWithValidatedConfig(oi.Game, oi.op, cfg, oi.Game.SetConfig, oi.Reset)
}

// Bid ビッドを宣言
func (oi *OhHellInteractor) Bid(bid int) string {
	if out, blocked := guardGameEnd(oi.Game, oi.op); blocked {
		return out
	}
	err := oi.Game.PlayerBid(bid)
	if err != nil {
		return oi.op.Output(oi.Game, err)
	}
	oi.runCpuBids()
	if oi.Game.GetPhase() == domain.OhHellPhasePlay {
		oi.runCpuTurns()
	}
	return oi.op.Output(oi.Game, nil)
}

// Play カードをプレイ
func (oi *OhHellInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(oi.Game, oi.op); blocked {
		return out
	}
	err := oi.Game.PlayerPlay(cardIndex)
	if err != nil {
		return oi.op.Output(oi.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決
	if oi.Game.GetPhase() == domain.OhHellPhaseTrickEnd {
		oi.Game.ResolveTrick()
	}
	oi.runCpuTurns()
	return oi.op.Output(oi.Game, nil)
}

// NextTrick 次のトリックへ進む
func (oi *OhHellInteractor) NextTrick() string {
	oi.Game.NextTrick()
	oi.runCpuTurns()
	return oi.op.Output(oi.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (oi *OhHellInteractor) NextRound() string {
	oi.Game.ScoreRound()
	if out, blocked := guardGameEnd(oi.Game, oi.op); blocked {
		return out
	}
	oi.Game.NextRound()
	oi.runCpuBids()
	if oi.Game.GetPhase() == domain.OhHellPhasePlay {
		oi.runCpuTurns()
	}
	return oi.op.Output(oi.Game, nil)
}

// GetConfig 現在の設定を取得
func (oi *OhHellInteractor) GetConfig() domain.OhHellConfig {
	return oi.Game.GetConfig()
}

// Hint ヒント取得
func (oi *OhHellInteractor) Hint() string {
	return oi.op.HintOutput(oi.Game)
}

// ActionLog 棋譜を出力する
func (oi *OhHellInteractor) ActionLog() string {
	return oi.op.ActionLogOutput(oi.Game)
}

// runCpuBids ゲームが終わるかヒューマンのビッド番またはビッドフェーズが終了するまでCPUビッドを実行
func (oi *OhHellInteractor) runCpuBids() {
	runCpuBidsLoop(oi.Game, domain.OhHellPhaseBid)
}

// runCpuTurns ゲームが終わるか人間の手番またはトリック/ラウンド終了になるまでCPUターンを実行
func (oi *OhHellInteractor) runCpuTurns() {
	runCpuTurnsLoop(oi.Game, trickPhases[domain.OhHellPhase]{
		play:     domain.OhHellPhasePlay,
		trickEnd: domain.OhHellPhaseTrickEnd,
		roundEnd: domain.OhHellPhaseRoundEnd,
		gameEnd:  domain.OhHellPhaseGameEnd,
	})
}

// RestoreOhHellInteractor deserialises JSON into an OhHellInteractor.
func RestoreOhHellInteractor(data []byte, op presenter.OhHellPresenter) (*OhHellInteractor, error) {
	o, err := restoreGame[domain.OhHell](data)
	if err != nil {
		return nil, err
	}
	return &OhHellInteractor{GameBase: GameBase[interfaces.OhHellGame]{Game: o}, op: op}, nil
}
