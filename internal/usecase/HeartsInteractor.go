package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// HeartsInteractorIF ハーツインタラクターインタフェース
type HeartsInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
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
	GameBase[interfaces.HeartsGame]
	hp presenter.HeartsPresenter
}

// NewHeartsInteractor コンストラクタ
func NewHeartsInteractor(h interfaces.HeartsGame, hp presenter.HeartsPresenter) *HeartsInteractor {
	mustNotNil("HeartsInteractor", map[string]any{"h": h, "hp": hp})
	return &HeartsInteractor{GameBase: GameBase[interfaces.HeartsGame]{Game: h}, hp: hp}
}

// Reset ゲーム初期化
func (hi *HeartsInteractor) Reset() string {
	hi.Game.Reset()
	if hi.Game.GetPassDirection() == domain.HeartsPassNone {
		hi.runCpuTurns()
	}
	return hi.hp.Output(hi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (hi *HeartsInteractor) ResetWithConfig(cfg domain.HeartsConfig) string {
	return resetWithValidatedConfig(hi.Game, hi.hp, cfg, hi.Game.SetConfig, hi.Reset)
}

// Pass カード交換
func (hi *HeartsInteractor) Pass(cardIndices []int) string {
	if out, blocked := guardGameEnd(hi.Game, hi.hp); blocked {
		return out
	}
	err := hi.Game.PlayerPass(cardIndices)
	if err != nil {
		return hi.hp.Output(hi.Game, err)
	}
	hi.Game.CpuPass()
	hi.Game.ExecutePass()
	hi.runCpuTurns()
	return hi.hp.Output(hi.Game, nil)
}

// Play カードをプレイ
func (hi *HeartsInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(hi.Game, hi.hp); blocked {
		return out
	}
	err := hi.Game.PlayerPlay(cardIndex)
	if err != nil {
		return hi.hp.Output(hi.Game, err)
	}
	hi.runCpuTurns()
	return hi.hp.Output(hi.Game, nil)
}

// NextTrick 次のトリックへ進む
func (hi *HeartsInteractor) NextTrick() string {
	hi.Game.NextTrick()
	hi.runCpuTurns()
	return hi.hp.Output(hi.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (hi *HeartsInteractor) NextRound() string {
	hi.Game.ScoreRound()
	if out, blocked := guardGameEnd(hi.Game, hi.hp); blocked {
		return out
	}
	hi.Game.NextRound()
	return hi.hp.Output(hi.Game, nil)
}

// GetConfig 現在の設定を取得
func (hi *HeartsInteractor) GetConfig() domain.HeartsConfig {
	return hi.Game.GetConfig()
}

// Hint ヒント取得
func (hi *HeartsInteractor) Hint() string {
	return hi.hp.HintOutput(hi.Game)
}

// ActionLog 棋譜を出力する
func (hi *HeartsInteractor) ActionLog() string {
	return hi.hp.ActionLogOutput(hi.Game)
}

// runCpuTurns ゲームが終わるか人間の手番またはトリック/ラウンド終了になるまでCPUターンを実行
func (hi *HeartsInteractor) runCpuTurns() {
	runCpuTurnsLoop(hi.Game, trickPhases[domain.HeartsPhase]{
		play:     domain.HeartsPhasePlay,
		trickEnd: domain.HeartsPhaseTrickEnd,
		roundEnd: domain.HeartsPhaseRoundEnd,
		gameEnd:  domain.HeartsPhaseGameEnd,
	})
}

// RestoreHeartsInteractor deserialises JSON into a HeartsInteractor.
func RestoreHeartsInteractor(data []byte, hp presenter.HeartsPresenter) (*HeartsInteractor, error) {
	return restoreAndBuild[domain.Hearts](data, func(g *domain.Hearts) *HeartsInteractor {
		return &HeartsInteractor{GameBase: GameBase[interfaces.HeartsGame]{Game: g}, hp: hp}
	})
}
