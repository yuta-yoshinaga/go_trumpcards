package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// HeartsInteractorIF ハーツインタラクターインタフェース
type HeartsInteractorIF interface {
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.HeartsConfig) string
	// Pass カード交換
	Pass(cardIndices []int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.HeartsConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// HeartsInteractor ハーツインタラクタークラス
type HeartsInteractor struct {
	h  interfaces.HeartsGame
	hp presenter.HeartsPresenter
}

// NewHeartsInteractor コンストラクタ
func NewHeartsInteractor(h interfaces.HeartsGame, hp presenter.HeartsPresenter) *HeartsInteractor {
	mustNotNil("HeartsInteractor", map[string]any{"h": h, "hp": hp})
	return &HeartsInteractor{h: h, hp: hp}
}

// Reset ゲーム初期化
func (hi *HeartsInteractor) Reset() string {
	hi.h.Reset()
	if hi.h.GetPassDirection() == domain.HeartsPassNone {
		hi.runCpuTurns()
	}
	return hi.hp.Output(hi.h, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (hi *HeartsInteractor) ResetWithConfig(cfg domain.HeartsConfig) string {
	return resetWithValidatedConfig(hi.h, hi.hp, cfg, hi.h.SetConfig, hi.Reset)
}

// Pass カード交換
func (hi *HeartsInteractor) Pass(cardIndices []int) string {
	if out, blocked := guardGameEnd(hi.h, hi.hp); blocked {
		return out
	}
	err := hi.h.PlayerPass(cardIndices)
	if err != nil {
		return hi.hp.Output(hi.h, err)
	}
	hi.h.CpuPass()
	hi.h.ExecutePass()
	hi.runCpuTurns()
	return hi.hp.Output(hi.h, nil)
}

// Play カードをプレイ
func (hi *HeartsInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(hi.h, hi.hp); blocked {
		return out
	}
	err := hi.h.PlayerPlay(cardIndex)
	if err != nil {
		return hi.hp.Output(hi.h, err)
	}
	hi.runCpuTurns()
	return hi.hp.Output(hi.h, nil)
}

// NextTrick 次のトリックへ進む
func (hi *HeartsInteractor) NextTrick() string {
	hi.h.NextTrick()
	hi.runCpuTurns()
	return hi.hp.Output(hi.h, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (hi *HeartsInteractor) NextRound() string {
	hi.h.ScoreRound()
	if out, blocked := guardGameEnd(hi.h, hi.hp); blocked {
		return out
	}
	hi.h.NextRound()
	return hi.hp.Output(hi.h, nil)
}

// GetConfig 現在の設定を取得
func (hi *HeartsInteractor) GetConfig() domain.HeartsConfig {
	return hi.h.GetConfig()
}

// Hint ヒント取得
func (hi *HeartsInteractor) Hint() string {
	return hi.hp.HintOutput(hi.h)
}

// ActionLog 棋譜を出力する
func (hi *HeartsInteractor) ActionLog() string {
	return hi.hp.ActionLogOutput(hi.h)
}

// runCpuTurns ゲームが終わるか人間の手番またはトリック/ラウンド終了になるまでCPUターンを実行
func (hi *HeartsInteractor) runCpuTurns() {
	for !hi.h.GetGameEndFlag() {
		phase := hi.h.GetPhase()
		if phase == domain.HeartsPhaseTrickEnd || phase == domain.HeartsPhaseRoundEnd || phase == domain.HeartsPhaseGameEnd {
			break
		}
		if phase != domain.HeartsPhasePlay {
			break
		}
		if hi.h.IsHumanTurn() {
			break
		}
		hi.h.CpuPlay()
		if hi.h.GetPhase() == domain.HeartsPhaseTrickEnd {
			hi.h.ResolveTrick()
			if hi.h.GetPhase() == domain.HeartsPhaseRoundEnd {
				break
			}
			hi.h.NextTrick()
		}
	}
}
