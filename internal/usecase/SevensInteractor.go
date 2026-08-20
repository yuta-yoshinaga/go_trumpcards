package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SevensInteractorIF 7並べインタラクターインタフェース
type SevensInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
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
	GameBase[interfaces.SevensGame]
	sp presenter.SevensPresenter
}

// NewSevensInteractor コンストラクタ
func NewSevensInteractor(s interfaces.SevensGame, sp presenter.SevensPresenter) *SevensInteractor {
	mustNotNil("SevensInteractor", map[string]any{"s": s, "sp": sp})
	return &SevensInteractor{
		GameBase: GameBase[interfaces.SevensGame]{Game: s},
		sp:       sp,
	}
}

// ResetWithConfig 設定付きゲーム初期化
func (si *SevensInteractor) ResetWithConfig(cfg domain.SevensConfig) string {
	si.Game.SetConfig(cfg)
	si.Game.Reset()
	si.runCpuTurns()
	return si.sp.Output(si.Game, nil)
}

// Reset ゲーム初期化
func (si *SevensInteractor) Reset() string {
	si.Game.Reset()
	si.runCpuTurns()
	return si.sp.Output(si.Game, nil)
}

// Play 人間プレイヤーがカードを出す (または パスする)
// idx: 出すカードのインデックス。-1 の場合はパス。
func (si *SevensInteractor) Play(idx int) string {
	if out, blocked := guardNotPlayable(si.Game, si.sp); blocked {
		return out
	}
	err := si.Game.PlayerPlay(idx)
	if err == nil && !si.Game.GetGameEndFlag() {
		si.runCpuTurns()
	}
	return si.sp.Output(si.Game, err)
}

// PlayJoker 人間プレイヤーがジョーカーを指定ポジションに出す
func (si *SevensInteractor) PlayJoker(cardIdx, targetSuit, targetValue int) string {
	if out, blocked := guardNotPlayable(si.Game, si.sp); blocked {
		return out
	}
	err := si.Game.PlayerPlayJoker(cardIdx, targetSuit, targetValue)
	if err == nil && !si.Game.GetGameEndFlag() {
		si.runCpuTurns()
	}
	return si.sp.Output(si.Game, err)
}

// ActionLog 棋譜を出力する
func (si *SevensInteractor) ActionLog() string {
	return si.sp.ActionLogOutput(si.Game)
}

// runCpuTurns ゲームが終わるか人間の手番になるまでCPUターンを実行
// 人間の手番になった場合でも選択肢がなければ自動処理する
func (si *SevensInteractor) runCpuTurns() {
	for i := 0; i < MaxCpuIterations; i++ {
		if si.Game.GetGameEndFlag() {
			return
		}
		if si.Game.IsHumanTurn() {
			// 人間に選択肢がなければ自動処理 (失格)
			if !si.Game.HasAnyOption(si.Game.GetCurrentTurn()) {
				si.Game.AutoHandleNoOption()
			} else {
				break
			}
		} else {
			si.Game.CpuPlay()
		}
	}
}

// RestoreSevensInteractor deserialises JSON into a SevensInteractor.
func RestoreSevensInteractor(data []byte, sp presenter.SevensPresenter) (*SevensInteractor, error) {
	return restoreAndBuild[domain.Sevens](data, func(g *domain.Sevens) *SevensInteractor {
		return &SevensInteractor{GameBase: GameBase[interfaces.SevensGame]{Game: g}, sp: sp}
	})
}
