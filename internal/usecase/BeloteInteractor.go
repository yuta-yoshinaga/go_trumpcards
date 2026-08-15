//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BeloteInteractorIF ベロートインタラクターインタフェース
type BeloteInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.BeloteConfig) string
	// PickUp ピックアップ判断 (orderUp=true で指名)
	PickUp(orderUp bool) string
	// CallTrump スートを指名する
	CallTrump(suit int) string
	// Pass 現在のフェーズに応じてパスする (PickUp or CallTrump)
	Pass() string
	// PassCall コールフェーズでパスする
	PassCall() string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.BeloteConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// BeloteInteractor ベロートインタラクタークラス
type BeloteInteractor struct {
	GameBase[interfaces.BeloteGame]
	bp presenter.BelotePresenter
}

// NewBeloteInteractor コンストラクタ
func NewBeloteInteractor(b interfaces.BeloteGame, bp presenter.BelotePresenter) *BeloteInteractor {
	mustNotNil("BeloteInteractor", map[string]any{"b": b, "bp": bp})
	return &BeloteInteractor{GameBase: GameBase[interfaces.BeloteGame]{Game: b}, bp: bp}
}

// Reset ゲーム初期化
func (bi *BeloteInteractor) Reset() string {
	bi.Game.Reset()
	bi.runCpuBids()
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (bi *BeloteInteractor) ResetWithConfig(cfg domain.BeloteConfig) string {
	return resetWithValidatedConfig(bi.Game, bi.bp, cfg, bi.Game.SetConfig, bi.Reset)
}

// PickUp ピックアップ判断
func (bi *BeloteInteractor) PickUp(orderUp bool) string {
	if out, blocked := guardGameEnd(bi.Game, bi.bp); blocked {
		return out
	}
	err := bi.Game.PlayerPickUp(orderUp)
	if err != nil {
		return bi.bp.Output(bi.Game, err)
	}
	bi.runCpuBids()
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// CallTrump スートを指名する
func (bi *BeloteInteractor) CallTrump(suit int) string {
	if out, blocked := guardGameEnd(bi.Game, bi.bp); blocked {
		return out
	}
	err := bi.Game.PlayerCallTrump(suit)
	if err != nil {
		return bi.bp.Output(bi.Game, err)
	}
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// Pass 現在のフェーズに応じてパスする (PickUp → PickUp(false), CallTrump → PassCall)
func (bi *BeloteInteractor) Pass() string {
	if out, blocked := guardGameEnd(bi.Game, bi.bp); blocked {
		return out
	}
	switch bi.Game.GetPhase() {
	case domain.BelotePhaseBidPickUp:
		return bi.PickUp(false)
	case domain.BelotePhaseBidCallTrump:
		return bi.PassCall()
	default:
		return bi.bp.Output(bi.Game, domain.ErrWrongPhase)
	}
}

// PassCall コールフェーズでパスする
func (bi *BeloteInteractor) PassCall() string {
	if out, blocked := guardGameEnd(bi.Game, bi.bp); blocked {
		return out
	}
	err := bi.Game.PlayerPassCall()
	if err != nil {
		return bi.bp.Output(bi.Game, err)
	}
	bi.runCpuBids()
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// Play カードをプレイ
func (bi *BeloteInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(bi.Game, bi.bp); blocked {
		return out
	}
	err := bi.Game.PlayerPlay(cardIndex)
	if err != nil {
		return bi.bp.Output(bi.Game, err)
	}
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// NextTrick トリックを解決して次のトリックへ進む
func (bi *BeloteInteractor) NextTrick() string {
	bi.Game.ResolveTrick()
	if out, blocked := guardGameEnd(bi.Game, bi.bp); blocked {
		return out
	}
	bi.Game.NextTrick()
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (bi *BeloteInteractor) NextRound() string {
	bi.Game.ScoreRound()
	if out, blocked := guardGameEnd(bi.Game, bi.bp); blocked {
		return out
	}
	bi.Game.NextRound()
	bi.runCpuBids()
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// GetConfig 現在の設定を取得
func (bi *BeloteInteractor) GetConfig() domain.BeloteConfig {
	return bi.Game.GetConfig()
}

// Hint ヒント取得
func (bi *BeloteInteractor) Hint() string {
	return bi.bp.HintOutput(bi.Game)
}

// ActionLog 棋譜を出力する
func (bi *BeloteInteractor) ActionLog() string {
	return bi.bp.ActionLogOutput(bi.Game)
}

// runCpuBids ビッドフェーズでCPUを自動実行する (PickUp と CallTrump)
func (bi *BeloteInteractor) runCpuBids() {
	for i := 0; i < MaxCpuIterations; i++ {
		if bi.Game.GetGameEndFlag() {
			return
		}
		phase := bi.Game.GetPhase()
		switch phase {
		case domain.BelotePhaseBidPickUp:
			if bi.Game.IsHumanBidTurn() {
				return
			}
			bi.Game.CpuPickUp()
		case domain.BelotePhaseBidCallTrump:
			if bi.Game.IsHumanBidTurn() {
				return
			}
			bi.Game.CpuCallTrump()
		default:
			return
		}
	}
}

// runCpuTurns プレイフェーズでCPUターンを自動実行する
func (bi *BeloteInteractor) runCpuTurns() {
	runCpuTurnsLoop(bi.Game, trickPhases[domain.BelotePhase]{
		play:     domain.BelotePhasePlay,
		trickEnd: domain.BelotePhaseTrickEnd,
		roundEnd: domain.BelotePhaseRoundEnd,
		gameEnd:  domain.BelotePhaseGameEnd,
	})
}

// RestoreBeloteInteractor deserialises JSON into a BeloteInteractor.
func RestoreBeloteInteractor(data []byte, bp presenter.BelotePresenter) (*BeloteInteractor, error) {
	b, err := restoreGame[domain.Belote](data)
	if err != nil {
		return nil, err
	}
	return &BeloteInteractor{GameBase: GameBase[interfaces.BeloteGame]{Game: b}, bp: bp}, nil
}
