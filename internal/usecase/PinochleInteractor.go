package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// PinochleInteractorIF ピノクルインタラクターインタフェース
type PinochleInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.PinochleConfig) string
	// Bid ビッドする
	Bid(amount int) string
	// Pass パスする
	Pass() string
	// CallTrump トランプスートを宣言する
	CallTrump(suit int) string
	// ConfirmMelds メルドを確認
	ConfirmMelds() string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.PinochleConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// PinochleInteractor ピノクルインタラクタークラス
type PinochleInteractor struct {
	GameBase[interfaces.PinochleGame]
	pp presenter.PinochlePresenter
}

// NewPinochleInteractor コンストラクタ
func NewPinochleInteractor(p interfaces.PinochleGame, pp presenter.PinochlePresenter) *PinochleInteractor {
	mustNotNil("PinochleInteractor", map[string]any{"p": p, "pp": pp})
	return &PinochleInteractor{GameBase: GameBase[interfaces.PinochleGame]{Game: p}, pp: pp}
}

// Reset ゲーム初期化
func (pi *PinochleInteractor) Reset() string {
	pi.Game.Reset()
	pi.runCpuBids()
	return pi.pp.Output(pi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (pi *PinochleInteractor) ResetWithConfig(cfg domain.PinochleConfig) string {
	return resetWithValidatedConfig(pi.Game, pi.pp, cfg, pi.Game.SetConfig, pi.Reset)
}

// Bid ビッドする
func (pi *PinochleInteractor) Bid(amount int) string {
	if out, blocked := guardGameEnd(pi.Game, pi.pp); blocked {
		return out
	}
	err := pi.Game.PlayerBid(amount)
	if err != nil {
		return pi.pp.Output(pi.Game, err)
	}
	pi.runCpuBids()
	return pi.pp.Output(pi.Game, nil)
}

// Pass パスする
func (pi *PinochleInteractor) Pass() string {
	if out, blocked := guardGameEnd(pi.Game, pi.pp); blocked {
		return out
	}
	err := pi.Game.PlayerPass()
	if err != nil {
		return pi.pp.Output(pi.Game, err)
	}
	pi.runCpuBids()
	return pi.pp.Output(pi.Game, nil)
}

// CallTrump トランプスートを宣言する
func (pi *PinochleInteractor) CallTrump(suit int) string {
	if out, blocked := guardGameEnd(pi.Game, pi.pp); blocked {
		return out
	}
	err := pi.Game.PlayerCallTrump(suit)
	if err != nil {
		return pi.pp.Output(pi.Game, err)
	}
	return pi.pp.Output(pi.Game, nil)
}

// ConfirmMelds メルドを確認してプレイフェーズに進む
func (pi *PinochleInteractor) ConfirmMelds() string {
	if out, blocked := guardGameEnd(pi.Game, pi.pp); blocked {
		return out
	}
	pi.Game.ConfirmMelds()
	pi.runCpuTurns()
	return pi.pp.Output(pi.Game, nil)
}

// Play カードをプレイ
func (pi *PinochleInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(pi.Game, pi.pp); blocked {
		return out
	}
	err := pi.Game.PlayerPlay(cardIndex)
	if err != nil {
		return pi.pp.Output(pi.Game, err)
	}
	pi.runCpuTurns()
	return pi.pp.Output(pi.Game, nil)
}

// NextTrick 次のトリックへ進む
func (pi *PinochleInteractor) NextTrick() string {
	pi.Game.ResolveTrick()
	if out, blocked := guardGameEnd(pi.Game, pi.pp); blocked {
		return out
	}
	pi.Game.NextTrick()
	pi.runCpuTurns()
	return pi.pp.Output(pi.Game, nil)
}

// NextRound 次のラウンドへ進む
func (pi *PinochleInteractor) NextRound() string {
	if out, blocked := guardGameEnd(pi.Game, pi.pp); blocked {
		return out
	}
	pi.Game.NextRound()
	pi.runCpuBids()
	return pi.pp.Output(pi.Game, nil)
}

// GetConfig 現在の設定を取得
func (pi *PinochleInteractor) GetConfig() domain.PinochleConfig {
	return pi.Game.GetConfig()
}

// Hint ヒント取得
func (pi *PinochleInteractor) Hint() string {
	return pi.pp.HintOutput(pi.Game)
}

// ActionLog 棋譜を出力する
func (pi *PinochleInteractor) ActionLog() string {
	return pi.pp.ActionLogOutput(pi.Game)
}

// runCpuBids ビッドおよびトランプ宣言フェーズでCPUを自動実行する
func (pi *PinochleInteractor) runCpuBids() {
	for i := 0; i < MaxCpuIterations; i++ {
		if pi.Game.GetGameEndFlag() {
			return
		}
		phase := pi.Game.GetPhase()
		if phase == domain.PinochlePhaseBid {
			if pi.Game.IsHumanBidTurn() {
				break
			}
			pi.Game.CpuBid()
		} else if phase == domain.PinochlePhaseTrump {
			if pi.Game.IsHumanTurn() {
				break
			}
			pi.Game.CpuCallTrump()
		} else {
			break
		}
	}
}

// runCpuTurns プレイフェーズでCPUターンを自動実行する
func (pi *PinochleInteractor) runCpuTurns() {
	runCpuTurnsLoop(pi.Game, trickPhases[domain.PinochlePhase]{
		play:     domain.PinochlePhasePlay,
		trickEnd: domain.PinochlePhaseTrickEnd,
		roundEnd: domain.PinochlePhaseRoundEnd,
		gameEnd:  domain.PinochlePhaseGameEnd,
	})
}

// RestorePinochleInteractor deserialises JSON into a PinochleInteractor.
func RestorePinochleInteractor(data []byte, pp presenter.PinochlePresenter) (*PinochleInteractor, error) {
	return restoreAndBuild[domain.Pinochle](data, func(g *domain.Pinochle) *PinochleInteractor {
		return &PinochleInteractor{GameBase: GameBase[interfaces.PinochleGame]{Game: g}, pp: pp}
	})
}
