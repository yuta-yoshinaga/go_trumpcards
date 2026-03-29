package usecase

import (
	"encoding/json"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SevensInteractorIF 7並べインタラクターインタフェース
type SevensInteractorIF interface {
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定付きゲーム初期化
	ResetWithConfig(cfg domain.SevensConfig) string
	// Play 人間プレイヤーがカードを出す (または パスする)
	Play(idx int) string
	// PlayJoker 人間プレイヤーがジョーカーを指定ポジションに出す
	PlayJoker(cardIdx, targetSuit, targetValue int) string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// SevensInteractor 7並べインタラクタークラス
type SevensInteractor struct {
	s  interfaces.SevensGame
	sp presenter.SevensPresenter
}

// NewSevensInteractor コンストラクタ
func NewSevensInteractor(s interfaces.SevensGame, sp presenter.SevensPresenter) *SevensInteractor {
	mustNotNil("SevensInteractor", map[string]any{"s": s, "sp": sp})
	return &SevensInteractor{
		s:  s,
		sp: sp,
	}
}

// ResetWithConfig 設定付きゲーム初期化
func (si *SevensInteractor) ResetWithConfig(cfg domain.SevensConfig) string {
	si.s.SetConfig(cfg)
	si.s.Reset()
	si.runCpuTurns()
	return si.sp.Output(si.s, nil)
}

// Reset ゲーム初期化
func (si *SevensInteractor) Reset() string {
	si.s.Reset()
	si.runCpuTurns()
	return si.sp.Output(si.s, nil)
}

// Play 人間プレイヤーがカードを出す (または パスする)
// idx: 出すカードのインデックス。-1 の場合はパス。
func (si *SevensInteractor) Play(idx int) string {
	if out, blocked := guardNotPlayable(si.s, si.sp); blocked {
		return out
	}
	err := si.s.PlayerPlay(idx)
	if err == nil && !si.s.GetGameEndFlag() {
		si.runCpuTurns()
	}
	return si.sp.Output(si.s, err)
}

// PlayJoker 人間プレイヤーがジョーカーを指定ポジションに出す
func (si *SevensInteractor) PlayJoker(cardIdx, targetSuit, targetValue int) string {
	if out, blocked := guardNotPlayable(si.s, si.sp); blocked {
		return out
	}
	err := si.s.PlayerPlayJoker(cardIdx, targetSuit, targetValue)
	if err == nil && !si.s.GetGameEndFlag() {
		si.runCpuTurns()
	}
	return si.sp.Output(si.s, err)
}

// ActionLog 棋譜を出力する
func (si *SevensInteractor) ActionLog() string {
	return si.sp.ActionLogOutput(si.s)
}

// runCpuTurns ゲームが終わるか人間の手番になるまでCPUターンを実行
// 人間の手番になった場合でも選択肢がなければ自動処理する
func (si *SevensInteractor) runCpuTurns() {
	for !si.s.GetGameEndFlag() {
		if si.s.IsHumanTurn() {
			// 人間に選択肢がなければ自動処理 (失格)
			if !si.s.HasAnyOption(si.s.GetCurrentTurn()) {
				si.s.AutoHandleNoOption()
			} else {
				break
			}
		} else {
			si.s.CpuPlay()
		}
	}
}

// Snapshot serialises the game state to JSON for KV persistence.
func (si *SevensInteractor) Snapshot() ([]byte, error) {
	return json.Marshal(si.s)
}

// RestoreSevensInteractor deserialises JSON into a SevensInteractor.
func RestoreSevensInteractor(data []byte, sp presenter.SevensPresenter) (*SevensInteractor, error) {
	s, err := restoreGame[domain.Sevens](data)
	if err != nil {
		return nil, err
	}
	return &SevensInteractor{s: s, sp: sp}, nil
}
