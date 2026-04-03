package usecase

import (
	"encoding/json"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// PinochleInteractorIF ピノクルインタラクターインタフェース
type PinochleInteractorIF interface {
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
	p  interfaces.PinochleGame
	pp presenter.PinochlePresenter
}

// NewPinochleInteractor コンストラクタ
func NewPinochleInteractor(p interfaces.PinochleGame, pp presenter.PinochlePresenter) *PinochleInteractor {
	mustNotNil("PinochleInteractor", map[string]any{"p": p, "pp": pp})
	return &PinochleInteractor{p: p, pp: pp}
}

// Reset ゲーム初期化
func (pi *PinochleInteractor) Reset() string {
	pi.p.Reset()
	pi.runCpuBids()
	return pi.pp.Output(pi.p, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (pi *PinochleInteractor) ResetWithConfig(cfg domain.PinochleConfig) string {
	return resetWithValidatedConfig(pi.p, pi.pp, cfg, pi.p.SetConfig, pi.Reset)
}

// Bid ビッドする
func (pi *PinochleInteractor) Bid(amount int) string {
	if out, blocked := guardGameEnd(pi.p, pi.pp); blocked {
		return out
	}
	err := pi.p.PlayerBid(amount)
	if err != nil {
		return pi.pp.Output(pi.p, err)
	}
	pi.runCpuBids()
	return pi.pp.Output(pi.p, nil)
}

// Pass パスする
func (pi *PinochleInteractor) Pass() string {
	if out, blocked := guardGameEnd(pi.p, pi.pp); blocked {
		return out
	}
	err := pi.p.PlayerPass()
	if err != nil {
		return pi.pp.Output(pi.p, err)
	}
	pi.runCpuBids()
	return pi.pp.Output(pi.p, nil)
}

// CallTrump トランプスートを宣言する
func (pi *PinochleInteractor) CallTrump(suit int) string {
	if out, blocked := guardGameEnd(pi.p, pi.pp); blocked {
		return out
	}
	err := pi.p.PlayerCallTrump(suit)
	if err != nil {
		return pi.pp.Output(pi.p, err)
	}
	return pi.pp.Output(pi.p, nil)
}

// ConfirmMelds メルドを確認してプレイフェーズに進む
func (pi *PinochleInteractor) ConfirmMelds() string {
	if out, blocked := guardGameEnd(pi.p, pi.pp); blocked {
		return out
	}
	pi.p.ConfirmMelds()
	pi.runCpuTurns()
	return pi.pp.Output(pi.p, nil)
}

// Play カードをプレイ
func (pi *PinochleInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(pi.p, pi.pp); blocked {
		return out
	}
	err := pi.p.PlayerPlay(cardIndex)
	if err != nil {
		return pi.pp.Output(pi.p, err)
	}
	pi.runCpuTurns()
	return pi.pp.Output(pi.p, nil)
}

// NextTrick 次のトリックへ進む
func (pi *PinochleInteractor) NextTrick() string {
	pi.p.ResolveTrick()
	if out, blocked := guardGameEnd(pi.p, pi.pp); blocked {
		return out
	}
	pi.p.NextTrick()
	pi.runCpuTurns()
	return pi.pp.Output(pi.p, nil)
}

// NextRound 次のラウンドへ進む
func (pi *PinochleInteractor) NextRound() string {
	if out, blocked := guardGameEnd(pi.p, pi.pp); blocked {
		return out
	}
	pi.p.NextRound()
	pi.runCpuBids()
	return pi.pp.Output(pi.p, nil)
}

// GetConfig 現在の設定を取得
func (pi *PinochleInteractor) GetConfig() domain.PinochleConfig {
	return pi.p.GetConfig()
}

// Hint ヒント取得
func (pi *PinochleInteractor) Hint() string {
	return pi.pp.HintOutput(pi.p)
}

// ActionLog 棋譜を出力する
func (pi *PinochleInteractor) ActionLog() string {
	return pi.pp.ActionLogOutput(pi.p)
}

// runCpuBids ビッドおよびトランプ宣言フェーズでCPUを自動実行する
func (pi *PinochleInteractor) runCpuBids() {
	for !pi.p.GetGameEndFlag() {
		phase := pi.p.GetPhase()
		if phase == domain.PinochlePhaseBid {
			if pi.p.IsHumanBidTurn() {
				break
			}
			pi.p.CpuBid()
		} else if phase == domain.PinochlePhaseTrump {
			if pi.p.IsHumanTurn() {
				break
			}
			pi.p.CpuCallTrump()
		} else {
			break
		}
	}
}

// runCpuTurns プレイフェーズでCPUターンを自動実行する
func (pi *PinochleInteractor) runCpuTurns() {
	runCpuTurnsLoop(pi.p, trickPhases[domain.PinochlePhase]{
		play:     domain.PinochlePhasePlay,
		trickEnd: domain.PinochlePhaseTrickEnd,
		roundEnd: domain.PinochlePhaseRoundEnd,
		gameEnd:  domain.PinochlePhaseGameEnd,
	})
}

// Snapshot serialises the game state to JSON for KV persistence.
func (pi *PinochleInteractor) Snapshot() ([]byte, error) {
	return json.Marshal(pi.p)
}

// RestorePinochleInteractor deserialises JSON into a PinochleInteractor.
func RestorePinochleInteractor(data []byte, pp presenter.PinochlePresenter) (*PinochleInteractor, error) {
	p, err := restoreGame[domain.Pinochle](data)
	if err != nil {
		return nil, err
	}
	return &PinochleInteractor{p: p, pp: pp}, nil
}
