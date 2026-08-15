//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// EuchreInteractorIF ユーカーインタラクターインタフェース
type EuchreInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
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
	GameBase[interfaces.EuchreGame]
	ep presenter.EuchrePresenter
}

// NewEuchreInteractor コンストラクタ
func NewEuchreInteractor(e interfaces.EuchreGame, ep presenter.EuchrePresenter) *EuchreInteractor {
	mustNotNil("EuchreInteractor", map[string]any{"e": e, "ep": ep})
	return &EuchreInteractor{GameBase: GameBase[interfaces.EuchreGame]{Game: e}, ep: ep}
}

// Reset ゲーム初期化
func (ei *EuchreInteractor) Reset() string {
	ei.Game.Reset()
	ei.runCpuBids()
	ei.runCpuDiscard()
	ei.runCpuTurns()
	return ei.ep.Output(ei.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ei *EuchreInteractor) ResetWithConfig(cfg domain.EuchreConfig) string {
	return resetWithValidatedConfig(ei.Game, ei.ep, cfg, ei.Game.SetConfig, ei.Reset)
}

// PickUp ピックアップ判断
func (ei *EuchreInteractor) PickUp(orderUp bool, goAlone bool) string {
	if out, blocked := guardGameEnd(ei.Game, ei.ep); blocked {
		return out
	}
	err := ei.Game.PlayerPickUp(orderUp, goAlone)
	if err != nil {
		return ei.ep.Output(ei.Game, err)
	}
	ei.runCpuBids()
	ei.runCpuDiscard()
	ei.runCpuTurns()
	return ei.ep.Output(ei.Game, nil)
}

// CallTrump スートを指名する
func (ei *EuchreInteractor) CallTrump(suit int, goAlone bool) string {
	if out, blocked := guardGameEnd(ei.Game, ei.ep); blocked {
		return out
	}
	err := ei.Game.PlayerCallTrump(suit, goAlone)
	if err != nil {
		return ei.ep.Output(ei.Game, err)
	}
	ei.runCpuTurns()
	return ei.ep.Output(ei.Game, nil)
}

// Pass 現在のフェーズに応じてパスする (PickUp → PickUp(false,false), CallTrump → PassCall)
func (ei *EuchreInteractor) Pass() string {
	if out, blocked := guardGameEnd(ei.Game, ei.ep); blocked {
		return out
	}
	phase := ei.Game.GetPhase()
	switch phase {
	case domain.EuchrePhasePickUp:
		return ei.PickUp(false, false)
	case domain.EuchrePhaseCallTrump:
		return ei.PassCall()
	default:
		return ei.ep.Output(ei.Game, domain.ErrWrongPhase)
	}
}

// PassCall コールフェーズでパスする
func (ei *EuchreInteractor) PassCall() string {
	if out, blocked := guardGameEnd(ei.Game, ei.ep); blocked {
		return out
	}
	err := ei.Game.PlayerPassCall()
	if err != nil {
		return ei.ep.Output(ei.Game, err)
	}
	ei.runCpuBids()
	ei.runCpuDiscard()
	ei.runCpuTurns()
	return ei.ep.Output(ei.Game, nil)
}

// Discard カードを捨てる
func (ei *EuchreInteractor) Discard(cardIndex int) string {
	if out, blocked := guardGameEnd(ei.Game, ei.ep); blocked {
		return out
	}
	err := ei.Game.PlayerDiscard(cardIndex)
	if err != nil {
		return ei.ep.Output(ei.Game, err)
	}
	ei.runCpuTurns()
	return ei.ep.Output(ei.Game, nil)
}

// Play カードをプレイ
func (ei *EuchreInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ei.Game, ei.ep); blocked {
		return out
	}
	err := ei.Game.PlayerPlay(cardIndex)
	if err != nil {
		return ei.ep.Output(ei.Game, err)
	}
	ei.runCpuTurns()
	return ei.ep.Output(ei.Game, nil)
}

// NextTrick トリックを解決して次のトリックへ進む
func (ei *EuchreInteractor) NextTrick() string {
	ei.Game.ResolveTrick()
	if out, blocked := guardGameEnd(ei.Game, ei.ep); blocked {
		return out
	}
	ei.Game.NextTrick()
	ei.runCpuTurns()
	return ei.ep.Output(ei.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (ei *EuchreInteractor) NextRound() string {
	ei.Game.ScoreRound()
	if out, blocked := guardGameEnd(ei.Game, ei.ep); blocked {
		return out
	}
	ei.Game.NextRound()
	ei.runCpuBids()
	ei.runCpuDiscard()
	ei.runCpuTurns()
	return ei.ep.Output(ei.Game, nil)
}

// GetConfig 現在の設定を取得
func (ei *EuchreInteractor) GetConfig() domain.EuchreConfig {
	return ei.Game.GetConfig()
}

// Hint ヒント取得
func (ei *EuchreInteractor) Hint() string {
	return ei.ep.HintOutput(ei.Game)
}

// ActionLog 棋譜を出力する
func (ei *EuchreInteractor) ActionLog() string {
	return ei.ep.ActionLogOutput(ei.Game)
}

// runCpuBids PickUpとCallTrumpフェーズでCPUを自動実行する
func (ei *EuchreInteractor) runCpuBids() {
	for i := 0; i < MaxCpuIterations; i++ {
		if ei.Game.GetGameEndFlag() {
			return
		}
		phase := ei.Game.GetPhase()
		if phase == domain.EuchrePhasePickUp {
			if ei.Game.IsHumanBidTurn() {
				break
			}
			ei.Game.CpuPickUp()
		} else if phase == domain.EuchrePhaseCallTrump {
			if ei.Game.IsHumanBidTurn() {
				break
			}
			ei.Game.CpuCallTrump()
		} else {
			break
		}
	}
}

// runCpuDiscard DiscardフェーズでCPUディーラーを自動実行する
func (ei *EuchreInteractor) runCpuDiscard() {
	if ei.Game.GetGameEndFlag() {
		return
	}
	if ei.Game.GetPhase() == domain.EuchrePhaseDiscard {
		if !ei.Game.IsHumanTurn() {
			ei.Game.CpuDiscard()
		}
	}
}

// runCpuTurns プレイフェーズでCPUターンを自動実行する
func (ei *EuchreInteractor) runCpuTurns() {
	runCpuTurnsLoop(ei.Game, trickPhases[domain.EuchrePhase]{
		play:     domain.EuchrePhasePlay,
		trickEnd: domain.EuchrePhaseTrickEnd,
		roundEnd: domain.EuchrePhaseRoundEnd,
		gameEnd:  domain.EuchrePhaseGameEnd,
	})
}

// RestoreEuchreInteractor deserialises JSON into an EuchreInteractor.
func RestoreEuchreInteractor(data []byte, ep presenter.EuchrePresenter) (*EuchreInteractor, error) {
	e, err := restoreGame[domain.Euchre](data)
	if err != nil {
		return nil, err
	}
	return &EuchreInteractor{GameBase: GameBase[interfaces.EuchreGame]{Game: e}, ep: ep}, nil
}
