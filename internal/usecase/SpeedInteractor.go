package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SpeedInteractorIF スピードインタラクターインタフェース
type SpeedInteractorIF interface {
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.SpeedConfig) string
	// Play 人間プレイヤーがカードを出す
	Play(cardIndex, pileIndex int) string
	// Flip 膠着時に台札をめくる
	Flip() string
	// Hint ヒントを取得する
	Hint() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.SpeedConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// SpeedInteractor スピードインタラクタークラス
type SpeedInteractor struct {
	GameBase[interfaces.SpeedGame]
	sp presenter.SpeedPresenter
}

// NewSpeedInteractor コンストラクタ
func NewSpeedInteractor(s interfaces.SpeedGame, sp presenter.SpeedPresenter) *SpeedInteractor {
	mustNotNil("SpeedInteractor", map[string]any{"s": s, "sp": sp})
	return &SpeedInteractor{GameBase: GameBase[interfaces.SpeedGame]{Game: s}, sp: sp}
}

// Reset ゲーム初期化
func (si *SpeedInteractor) Reset() string {
	return runAndPresent(si.Game, si.sp, si.Game.Reset)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (si *SpeedInteractor) ResetWithConfig(cfg domain.SpeedConfig) string {
	return resetWithValidatedConfig(si.Game, si.sp, cfg, si.Game.SetConfig, si.Reset)
}

// Play 人間プレイヤーがカードを出す、その後CPUが自動応答する
func (si *SpeedInteractor) Play(cardIndex, pileIndex int) string {
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	err := si.Game.PlayerPlay(cardIndex, pileIndex)
	if err != nil {
		return si.sp.Output(si.Game, err)
	}
	// CPU自動応答ループ
	if !si.Game.GetGameEndFlag() {
		si.Game.CpuPlay()
	}
	// フェーズ更新 (膠着判定)
	si.Game.UpdatePhase()
	return si.sp.Output(si.Game, nil)
}

// Flip 膠着時に台札をめくる
func (si *SpeedInteractor) Flip() string {
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	err := si.Game.Flip()
	if err != nil {
		return si.sp.Output(si.Game, err)
	}
	// フリップ後にCPU自動応答
	if !si.Game.GetGameEndFlag() {
		si.Game.CpuPlay()
		si.Game.UpdatePhase()
	}
	return si.sp.Output(si.Game, nil)
}

// Hint ヒントを取得する
func (si *SpeedInteractor) Hint() string {
	return si.sp.Output(si.Game, nil)
}

// GetConfig 現在の設定を取得
func (si *SpeedInteractor) GetConfig() domain.SpeedConfig {
	return si.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (si *SpeedInteractor) ActionLog() string {
	return si.sp.ActionLogOutput(si.Game)
}

// RestoreSpeedInteractor deserialises JSON into a SpeedInteractor.
func RestoreSpeedInteractor(data []byte, sp presenter.SpeedPresenter) (*SpeedInteractor, error) {
	return restoreAndBuild[domain.Speed](data, func(g *domain.Speed) *SpeedInteractor {
		return &SpeedInteractor{GameBase: GameBase[interfaces.SpeedGame]{Game: g}, sp: sp}
	})
}
