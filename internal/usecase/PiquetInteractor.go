//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// PiquetInteractorIF Piquetインタラクターインタフェース
type PiquetInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.PiquetConfig) string
	// ExchangeElder Elderの交換
	ExchangeElder(discardIndices []int) string
	// ExchangeYounger Youngerの交換
	ExchangeYounger(discardIndices []int) string
	// ResolveDeclaration 現宣言ステージを比較する
	ResolveDeclaration() string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextDeal 次のディールへ進む
	NextDeal() string
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.PiquetConfig
}

// PiquetInteractor Piquetインタラクタークラス
type PiquetInteractor struct {
	GameBase[interfaces.PiquetGame]
	pp presenter.PiquetPresenter
}

// NewPiquetInteractor コンストラクタ
func NewPiquetInteractor(p interfaces.PiquetGame, pp presenter.PiquetPresenter) *PiquetInteractor {
	mustNotNil("PiquetInteractor", map[string]any{"p": p, "pp": pp})
	return &PiquetInteractor{
		GameBase: GameBase[interfaces.PiquetGame]{Game: p},
		pp:       pp,
	}
}

// Reset ゲーム初期化
func (pi *PiquetInteractor) Reset() string {
	pi.Game.Reset()
	pi.runCpuExchange()
	return pi.pp.Output(pi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (pi *PiquetInteractor) ResetWithConfig(cfg domain.PiquetConfig) string {
	return resetWithValidatedConfig(pi.Game, pi.pp, cfg, pi.Game.SetConfig, pi.Reset)
}

// ExchangeElder Elderの交換
func (pi *PiquetInteractor) ExchangeElder(discardIndices []int) string {
	if out, blocked := guardGameEnd(pi.Game, pi.pp); blocked {
		return out
	}
	if err := pi.Game.ExchangeElder(discardIndices); err != nil {
		return pi.pp.Output(pi.Game, err)
	}
	pi.runCpuExchange()
	return pi.pp.Output(pi.Game, nil)
}

// ExchangeYounger Youngerの交換
func (pi *PiquetInteractor) ExchangeYounger(discardIndices []int) string {
	if out, blocked := guardGameEnd(pi.Game, pi.pp); blocked {
		return out
	}
	if err := pi.Game.ExchangeYounger(discardIndices); err != nil {
		return pi.pp.Output(pi.Game, err)
	}
	return pi.pp.Output(pi.Game, nil)
}

// ResolveDeclaration 現宣言ステージを比較する
func (pi *PiquetInteractor) ResolveDeclaration() string {
	if out, blocked := guardGameEnd(pi.Game, pi.pp); blocked {
		return out
	}
	if _, err := pi.Game.ResolveDeclaration(); err != nil {
		return pi.pp.Output(pi.Game, err)
	}
	// 宣言完了後はCPUプレイを進める
	pi.runCpuPlay()
	return pi.pp.Output(pi.Game, nil)
}

// Play カードをプレイ
func (pi *PiquetInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(pi.Game, pi.pp); blocked {
		return out
	}
	if err := pi.Game.PlayCard(cardIndex); err != nil {
		return pi.pp.Output(pi.Game, err)
	}
	pi.runCpuPlay()
	return pi.pp.Output(pi.Game, nil)
}

// NextDeal 次のディールへ進む
func (pi *PiquetInteractor) NextDeal() string {
	if pi.Game.GetGameEndFlag() {
		return pi.pp.Output(pi.Game, nil)
	}
	pi.Game.NextDeal()
	pi.runCpuExchange()
	return pi.pp.Output(pi.Game, nil)
}

// Hint ヒント取得
func (pi *PiquetInteractor) Hint() string {
	return pi.pp.HintOutput(pi.Game)
}

// ActionLog 棋譜を出力する
func (pi *PiquetInteractor) ActionLog() string {
	return pi.pp.ActionLogOutput(pi.Game)
}

// GetConfig 現在の設定を取得
func (pi *PiquetInteractor) GetConfig() domain.PiquetConfig {
	return pi.Game.GetConfig()
}

// runCpuExchange 交換フェーズでCPUの手番を自動実行する
func (pi *PiquetInteractor) runCpuExchange() {
	for i := 0; i < MaxCpuIterations; i++ {
		if pi.Game.GetGameEndFlag() {
			return
		}
		if pi.Game.GetPhase() != domain.PiquetPhaseExchange {
			return
		}
		turn := pi.Game.GetExchangeTurn()
		if turn == domain.PiquetExchangeTurnDone {
			return
		}
		var current int
		switch turn {
		case domain.PiquetExchangeTurnElder:
			current = pi.Game.GetElderIdx()
		case domain.PiquetExchangeTurnYounger:
			current = pi.Game.GetYoungerIdx()
		}
		if pi.Game.GetPlayer(current).GetIsHuman() {
			return
		}
		pi.Game.CpuPlay()
	}
}

// runCpuPlay プレイフェーズでCPUの手番を自動実行する
func (pi *PiquetInteractor) runCpuPlay() {
	for i := 0; i < MaxCpuIterations; i++ {
		if pi.Game.GetGameEndFlag() {
			return
		}
		if pi.Game.GetPhase() != domain.PiquetPhasePlay {
			return
		}
		if pi.Game.IsHumanTurn() {
			return
		}
		pi.Game.CpuPlay()
	}
}

// RestorePiquetInteractor deserialises JSON into a PiquetInteractor.
func RestorePiquetInteractor(data []byte, pp presenter.PiquetPresenter) (*PiquetInteractor, error) {
	return restoreAndBuild[domain.Piquet](data, func(g *domain.Piquet) *PiquetInteractor {
		return &PiquetInteractor{
			GameBase: GameBase[interfaces.PiquetGame]{Game: g},
			pp:       pp,
		}
	})
}
