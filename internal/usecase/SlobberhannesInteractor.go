//go:build !js || !wasm || classic

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SlobberhannesInteractorIF スロバーハンネスインタラクターインタフェース
type SlobberhannesInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.SlobberhannesConfig) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GiveUp 投了する
	GiveUp() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.SlobberhannesConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// SlobberhannesInteractor スロバーハンネスインタラクタークラス
type SlobberhannesInteractor struct {
	GameBase[interfaces.SlobberhannesGame]
	sp presenter.SlobberhannesPresenter
}

// NewSlobberhannesInteractor コンストラクタ
func NewSlobberhannesInteractor(s interfaces.SlobberhannesGame, sp presenter.SlobberhannesPresenter) *SlobberhannesInteractor {
	mustNotNil("SlobberhannesInteractor", map[string]any{"s": s, "sp": sp})
	return &SlobberhannesInteractor{GameBase: GameBase[interfaces.SlobberhannesGame]{Game: s}, sp: sp}
}

// Reset ゲーム初期化
func (si *SlobberhannesInteractor) Reset() string {
	si.Game.Reset()
	si.runCpuTurns()
	return si.sp.Output(si.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (si *SlobberhannesInteractor) ResetWithConfig(cfg domain.SlobberhannesConfig) string {
	return resetWithValidatedConfig(si.Game, si.sp, cfg, si.Game.SetConfig, si.Reset)
}

// Play カードをプレイ
func (si *SlobberhannesInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(si.Game, si.sp); blocked {
		return out
	}
	if err := si.Game.PlayerPlay(cardIndex); err != nil {
		return si.sp.Output(si.Game, err)
	}
	si.runCpuTurns()
	return si.sp.Output(si.Game, nil)
}

// NextRound 次のラウンドへ進む
func (si *SlobberhannesInteractor) NextRound() string {
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	si.Game.NextRound()
	si.runCpuTurns()
	return si.sp.Output(si.Game, nil)
}

// GiveUp 投了する
func (si *SlobberhannesInteractor) GiveUp() string {
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	si.Game.GiveUp()
	return si.sp.Output(si.Game, nil)
}

// GetConfig 現在の設定を取得
func (si *SlobberhannesInteractor) GetConfig() domain.SlobberhannesConfig {
	return si.Game.GetConfig()
}

// Hint ヒント取得
func (si *SlobberhannesInteractor) Hint() string {
	return si.sp.HintOutput(si.Game)
}

// ActionLog 棋譜を出力する
func (si *SlobberhannesInteractor) ActionLog() string {
	return si.sp.ActionLogOutput(si.Game)
}

// runCpuTurns 人間の手番になるかラウンド/ゲームが終わるまで CPU を進める。
//
// 共通の runCpuTurnsLoop は使えない。あちらはトリック解決が別ステップに
// なっているゲーム向けで、スロバーハンネスは 4 枚揃った時点でドメインが
// 自分で解決し、**ラウンド終了で必ず止まって人間の確認を待つ**ため。
func (si *SlobberhannesInteractor) runCpuTurns() {
	for turns := 0; !si.Game.GetGameEndFlag() && !si.Game.IsHumanTurn(); turns++ {
		// 進まない CpuPlay でハングしないための上限 (#4607 と同じ理由)。
		if turns >= maxCpuTurnsPerCall {
			return
		}
		if si.Game.GetPhase() != domain.SlobberhannesPhasePlay {
			// ラウンド終了。次へ進めるかは人間が決める。
			return
		}
		si.Game.CpuPlay()
	}
}

// RestoreSlobberhannesInteractor deserialises JSON into a SlobberhannesInteractor.
func RestoreSlobberhannesInteractor(data []byte, sp presenter.SlobberhannesPresenter) (*SlobberhannesInteractor, error) {
	return restoreAndBuild[domain.Slobberhannes](data, func(g *domain.Slobberhannes) *SlobberhannesInteractor {
		return &SlobberhannesInteractor{GameBase: GameBase[interfaces.SlobberhannesGame]{Game: g}, sp: sp}
	})
}
