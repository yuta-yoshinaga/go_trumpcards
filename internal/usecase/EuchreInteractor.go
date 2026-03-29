package usecase

import (
	"encoding/json"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// EuchreInteractorIF ユーカーインタラクターインタフェース
type EuchreInteractorIF interface {
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.EuchreConfig) string
	// PickUp ピックアップ判断 (orderUp=true で指名, goAlone=true でゴーアローン)
	PickUp(orderUp bool, goAlone bool) string
	// CallTrump スートを指名する
	CallTrump(suit int, goAlone bool) string
	// Pass 現在のフェーズに応じてパスする (PickUp or CallTrump)
	Pass() string
	// PassCall コールフェーズでパスする
	PassCall() string
	// Discard カードを捨てる
	Discard(cardIndex int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.EuchreConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// EuchreInteractor ユーカーインタラクタークラス
type EuchreInteractor struct {
	e  interfaces.EuchreGame
	ep presenter.EuchrePresenter
}

// NewEuchreInteractor コンストラクタ
func NewEuchreInteractor(e interfaces.EuchreGame, ep presenter.EuchrePresenter) *EuchreInteractor {
	mustNotNil("EuchreInteractor", map[string]any{"e": e, "ep": ep})
	return &EuchreInteractor{e: e, ep: ep}
}

// Reset ゲーム初期化
func (ei *EuchreInteractor) Reset() string {
	ei.e.Reset()
	ei.runCpuBids()
	ei.runCpuDiscard()
	ei.runCpuTurns()
	return ei.ep.Output(ei.e, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ei *EuchreInteractor) ResetWithConfig(cfg domain.EuchreConfig) string {
	return resetWithValidatedConfig(ei.e, ei.ep, cfg, ei.e.SetConfig, ei.Reset)
}

// PickUp ピックアップ判断
func (ei *EuchreInteractor) PickUp(orderUp bool, goAlone bool) string {
	if out, blocked := guardGameEnd(ei.e, ei.ep); blocked {
		return out
	}
	err := ei.e.PlayerPickUp(orderUp, goAlone)
	if err != nil {
		return ei.ep.Output(ei.e, err)
	}
	ei.runCpuBids()
	ei.runCpuDiscard()
	ei.runCpuTurns()
	return ei.ep.Output(ei.e, nil)
}

// CallTrump スートを指名する
func (ei *EuchreInteractor) CallTrump(suit int, goAlone bool) string {
	if out, blocked := guardGameEnd(ei.e, ei.ep); blocked {
		return out
	}
	err := ei.e.PlayerCallTrump(suit, goAlone)
	if err != nil {
		return ei.ep.Output(ei.e, err)
	}
	ei.runCpuTurns()
	return ei.ep.Output(ei.e, nil)
}

// Pass 現在のフェーズに応じてパスする (PickUp → PickUp(false,false), CallTrump → PassCall)
func (ei *EuchreInteractor) Pass() string {
	if out, blocked := guardGameEnd(ei.e, ei.ep); blocked {
		return out
	}
	phase := ei.e.GetPhase()
	switch phase {
	case domain.EuchrePhasePickUp:
		return ei.PickUp(false, false)
	case domain.EuchrePhaseCallTrump:
		return ei.PassCall()
	default:
		return ei.ep.Output(ei.e, domain.ErrWrongPhase)
	}
}

// PassCall コールフェーズでパスする
func (ei *EuchreInteractor) PassCall() string {
	if out, blocked := guardGameEnd(ei.e, ei.ep); blocked {
		return out
	}
	err := ei.e.PlayerPassCall()
	if err != nil {
		return ei.ep.Output(ei.e, err)
	}
	ei.runCpuBids()
	ei.runCpuDiscard()
	ei.runCpuTurns()
	return ei.ep.Output(ei.e, nil)
}

// Discard カードを捨てる
func (ei *EuchreInteractor) Discard(cardIndex int) string {
	if out, blocked := guardGameEnd(ei.e, ei.ep); blocked {
		return out
	}
	err := ei.e.PlayerDiscard(cardIndex)
	if err != nil {
		return ei.ep.Output(ei.e, err)
	}
	ei.runCpuTurns()
	return ei.ep.Output(ei.e, nil)
}

// Play カードをプレイ
func (ei *EuchreInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ei.e, ei.ep); blocked {
		return out
	}
	err := ei.e.PlayerPlay(cardIndex)
	if err != nil {
		return ei.ep.Output(ei.e, err)
	}
	ei.runCpuTurns()
	return ei.ep.Output(ei.e, nil)
}

// NextTrick トリックを解決して次のトリックへ進む
func (ei *EuchreInteractor) NextTrick() string {
	ei.e.ResolveTrick()
	if out, blocked := guardGameEnd(ei.e, ei.ep); blocked {
		return out
	}
	ei.e.NextTrick()
	ei.runCpuTurns()
	return ei.ep.Output(ei.e, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (ei *EuchreInteractor) NextRound() string {
	ei.e.ScoreRound()
	if out, blocked := guardGameEnd(ei.e, ei.ep); blocked {
		return out
	}
	ei.e.NextRound()
	ei.runCpuBids()
	ei.runCpuDiscard()
	ei.runCpuTurns()
	return ei.ep.Output(ei.e, nil)
}

// GetConfig 現在の設定を取得
func (ei *EuchreInteractor) GetConfig() domain.EuchreConfig {
	return ei.e.GetConfig()
}

// Hint ヒント取得
func (ei *EuchreInteractor) Hint() string {
	return ei.ep.HintOutput(ei.e)
}

// ActionLog 棋譜を出力する
func (ei *EuchreInteractor) ActionLog() string {
	return ei.ep.ActionLogOutput(ei.e)
}

// runCpuBids PickUpとCallTrumpフェーズでCPUを自動実行する
func (ei *EuchreInteractor) runCpuBids() {
	for !ei.e.GetGameEndFlag() {
		phase := ei.e.GetPhase()
		if phase == domain.EuchrePhasePickUp {
			if ei.e.IsHumanBidTurn() {
				break
			}
			ei.e.CpuPickUp()
		} else if phase == domain.EuchrePhaseCallTrump {
			if ei.e.IsHumanBidTurn() {
				break
			}
			ei.e.CpuCallTrump()
		} else {
			break
		}
	}
}

// runCpuDiscard DiscardフェーズでCPUディーラーを自動実行する
func (ei *EuchreInteractor) runCpuDiscard() {
	if ei.e.GetGameEndFlag() {
		return
	}
	if ei.e.GetPhase() == domain.EuchrePhaseDiscard {
		if !ei.e.IsHumanTurn() {
			ei.e.CpuDiscard()
		}
	}
}

// runCpuTurns プレイフェーズでCPUターンを自動実行する
func (ei *EuchreInteractor) runCpuTurns() {
	runCpuTurnsLoop(ei.e, trickPhases[domain.EuchrePhase]{
		play:     domain.EuchrePhasePlay,
		trickEnd: domain.EuchrePhaseTrickEnd,
		roundEnd: domain.EuchrePhaseRoundEnd,
		gameEnd:  domain.EuchrePhaseGameEnd,
	})
}

// Snapshot serialises the game state to JSON for KV persistence.
func (ei *EuchreInteractor) Snapshot() ([]byte, error) {
	return json.Marshal(ei.e)
}

// RestoreEuchreInteractor deserialises JSON into an EuchreInteractor.
func RestoreEuchreInteractor(data []byte, ep presenter.EuchrePresenter) (*EuchreInteractor, error) {
	e, err := restoreGame[domain.Euchre](data)
	if err != nil {
		return nil, err
	}
	return &EuchreInteractor{e: e, ep: ep}, nil
}
