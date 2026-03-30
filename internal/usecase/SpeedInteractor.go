package usecase

import (
	"encoding/json"

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
	s  interfaces.SpeedGame
	sp presenter.SpeedPresenter
}

// NewSpeedInteractor コンストラクタ
func NewSpeedInteractor(s interfaces.SpeedGame, sp presenter.SpeedPresenter) *SpeedInteractor {
	mustNotNil("SpeedInteractor", map[string]any{"s": s, "sp": sp})
	return &SpeedInteractor{s: s, sp: sp}
}

// Reset ゲーム初期化
func (si *SpeedInteractor) Reset() string {
	return runAndPresent(si.s, si.sp, si.s.Reset)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (si *SpeedInteractor) ResetWithConfig(cfg domain.SpeedConfig) string {
	return resetWithValidatedConfig(si.s, si.sp, cfg, si.s.SetConfig, si.Reset)
}

// Play 人間プレイヤーがカードを出す、その後CPUが自動応答する
func (si *SpeedInteractor) Play(cardIndex, pileIndex int) string {
	if out, blocked := guardGameEnd(si.s, si.sp); blocked {
		return out
	}
	err := si.s.PlayerPlay(cardIndex, pileIndex)
	if err != nil {
		return si.sp.Output(si.s, err)
	}
	// CPU自動応答ループ
	if !si.s.GetGameEndFlag() {
		si.s.CpuPlay()
	}
	// フェーズ更新 (膠着判定)
	si.s.UpdatePhase()
	return si.sp.Output(si.s, nil)
}

// Flip 膠着時に台札をめくる
func (si *SpeedInteractor) Flip() string {
	if out, blocked := guardGameEnd(si.s, si.sp); blocked {
		return out
	}
	err := si.s.Flip()
	if err != nil {
		return si.sp.Output(si.s, err)
	}
	// フリップ後にCPU自動応答
	if !si.s.GetGameEndFlag() {
		si.s.CpuPlay()
		si.s.UpdatePhase()
	}
	return si.sp.Output(si.s, nil)
}

// Hint ヒントを取得する
func (si *SpeedInteractor) Hint() string {
	return si.sp.Output(si.s, nil)
}

// GetConfig 現在の設定を取得
func (si *SpeedInteractor) GetConfig() domain.SpeedConfig {
	return si.s.GetConfig()
}

// ActionLog 棋譜を出力する
func (si *SpeedInteractor) ActionLog() string {
	return si.sp.ActionLogOutput(si.s)
}

// Snapshot serialises the game state to JSON for KV persistence.
func (si *SpeedInteractor) Snapshot() ([]byte, error) {
	return json.Marshal(si.s)
}

// RestoreSpeedInteractor deserialises JSON into a SpeedInteractor.
func RestoreSpeedInteractor(data []byte, sp presenter.SpeedPresenter) (*SpeedInteractor, error) {
	s, err := restoreGame[domain.Speed](data)
	if err != nil {
		return nil, err
	}
	return &SpeedInteractor{s: s, sp: sp}, nil
}
