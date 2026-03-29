package usecase

import (
	"encoding/json"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// OhHellInteractorIF オー・ヘルインタラクターインタフェース
type OhHellInteractorIF interface {
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
	o  interfaces.OhHellGame
	op presenter.OhHellPresenter
}

// NewOhHellInteractor コンストラクタ
func NewOhHellInteractor(o interfaces.OhHellGame, op presenter.OhHellPresenter) *OhHellInteractor {
	mustNotNil("OhHellInteractor", map[string]any{"o": o, "op": op})
	return &OhHellInteractor{o: o, op: op}
}

// Reset ゲーム初期化
func (oi *OhHellInteractor) Reset() string {
	oi.o.Reset()
	oi.runCpuBids()
	if oi.o.GetPhase() == domain.OhHellPhasePlay {
		oi.runCpuTurns()
	}
	return oi.op.Output(oi.o, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (oi *OhHellInteractor) ResetWithConfig(cfg domain.OhHellConfig) string {
	return resetWithValidatedConfig(oi.o, oi.op, cfg, oi.o.SetConfig, oi.Reset)
}

// Bid ビッドを宣言
func (oi *OhHellInteractor) Bid(bid int) string {
	if out, blocked := guardGameEnd(oi.o, oi.op); blocked {
		return out
	}
	err := oi.o.PlayerBid(bid)
	if err != nil {
		return oi.op.Output(oi.o, err)
	}
	oi.runCpuBids()
	if oi.o.GetPhase() == domain.OhHellPhasePlay {
		oi.runCpuTurns()
	}
	return oi.op.Output(oi.o, nil)
}

// Play カードをプレイ
func (oi *OhHellInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(oi.o, oi.op); blocked {
		return out
	}
	err := oi.o.PlayerPlay(cardIndex)
	if err != nil {
		return oi.op.Output(oi.o, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決
	if oi.o.GetPhase() == domain.OhHellPhaseTrickEnd {
		oi.o.ResolveTrick()
	}
	oi.runCpuTurns()
	return oi.op.Output(oi.o, nil)
}

// NextTrick 次のトリックへ進む
func (oi *OhHellInteractor) NextTrick() string {
	oi.o.NextTrick()
	oi.runCpuTurns()
	return oi.op.Output(oi.o, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (oi *OhHellInteractor) NextRound() string {
	oi.o.ScoreRound()
	if out, blocked := guardGameEnd(oi.o, oi.op); blocked {
		return out
	}
	oi.o.NextRound()
	oi.runCpuBids()
	if oi.o.GetPhase() == domain.OhHellPhasePlay {
		oi.runCpuTurns()
	}
	return oi.op.Output(oi.o, nil)
}

// GetConfig 現在の設定を取得
func (oi *OhHellInteractor) GetConfig() domain.OhHellConfig {
	return oi.o.GetConfig()
}

// Hint ヒント取得
func (oi *OhHellInteractor) Hint() string {
	return oi.op.HintOutput(oi.o)
}

// ActionLog 棋譜を出力する
func (oi *OhHellInteractor) ActionLog() string {
	return oi.op.ActionLogOutput(oi.o)
}

// runCpuBids ゲームが終わるかヒューマンのビッド番またはビッドフェーズが終了するまでCPUビッドを実行
func (oi *OhHellInteractor) runCpuBids() {
	for !oi.o.GetGameEndFlag() {
		if oi.o.GetPhase() != domain.OhHellPhaseBid {
			break
		}
		if oi.o.IsHumanBidTurn() {
			break
		}
		oi.o.CpuBid()
	}
}

// runCpuTurns ゲームが終わるか人間の手番またはトリック/ラウンド終了になるまでCPUターンを実行
func (oi *OhHellInteractor) runCpuTurns() {
	runCpuTurnsLoop(oi.o, trickPhases[domain.OhHellPhase]{
		play:     domain.OhHellPhasePlay,
		trickEnd: domain.OhHellPhaseTrickEnd,
		roundEnd: domain.OhHellPhaseRoundEnd,
		gameEnd:  domain.OhHellPhaseGameEnd,
	})
}

// Snapshot serialises the game state to JSON for KV persistence.
func (oi *OhHellInteractor) Snapshot() ([]byte, error) {
	return json.Marshal(oi.o)
}

// RestoreOhHellInteractor deserialises JSON into an OhHellInteractor.
func RestoreOhHellInteractor(data []byte, op presenter.OhHellPresenter) (*OhHellInteractor, error) {
	o, err := restoreGame[domain.OhHell](data)
	if err != nil {
		return nil, err
	}
	return &OhHellInteractor{o: o, op: op}, nil
}
