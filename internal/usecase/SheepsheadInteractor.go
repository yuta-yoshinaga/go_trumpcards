//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SheepsheadInteractorIF シープスヘッドのインタラクターインタフェース
type SheepsheadInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.SheepsheadConfig) string
	// Pick ピック(true)/パス(false)を選択
	Pick(pick bool) string
	// Bury ピッカーが2枚を埋める
	Bury(indices []int) string
	// Call 呼びスートを指定
	Call(suit int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.SheepsheadConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// SheepsheadInteractor シープスヘッドのインタラクタークラス
type SheepsheadInteractor struct {
	GameBase[interfaces.SheepsheadGame]
	sp presenter.SheepsheadPresenter
}

// NewSheepsheadInteractor コンストラクタ
func NewSheepsheadInteractor(g interfaces.SheepsheadGame, sp presenter.SheepsheadPresenter) *SheepsheadInteractor {
	mustNotNil("SheepsheadInteractor", map[string]any{"g": g, "sp": sp})
	return &SheepsheadInteractor{GameBase: GameBase[interfaces.SheepsheadGame]{Game: g}, sp: sp}
}

// Reset ゲーム初期化
func (si *SheepsheadInteractor) Reset() string {
	si.Game.Reset()
	si.runCpuTurns()
	return si.sp.Output(si.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (si *SheepsheadInteractor) ResetWithConfig(cfg domain.SheepsheadConfig) string {
	return resetWithValidatedConfig(si.Game, si.sp, cfg, si.Game.SetConfig, si.Reset)
}

// Pick ピック(true)/パス(false)を選択
func (si *SheepsheadInteractor) Pick(pick bool) string {
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	if err := si.Game.PlayerPick(pick); err != nil {
		return si.sp.Output(si.Game, err)
	}
	si.runCpuTurns()
	return si.sp.Output(si.Game, nil)
}

// Bury ピッカーが2枚を埋める
func (si *SheepsheadInteractor) Bury(indices []int) string {
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	if err := si.Game.PlayerBury(indices); err != nil {
		return si.sp.Output(si.Game, err)
	}
	si.runCpuTurns()
	return si.sp.Output(si.Game, nil)
}

// Call 呼びスートを指定
func (si *SheepsheadInteractor) Call(suit int) string {
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	if err := si.Game.PlayerCall(suit); err != nil {
		return si.sp.Output(si.Game, err)
	}
	si.runCpuTurns()
	return si.sp.Output(si.Game, nil)
}

// Play カードをプレイ
func (si *SheepsheadInteractor) Play(cardIndex int) string {
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	if si.Game.GetPhase() != domain.SheepsheadPhasePlay || !si.Game.IsHumanTurn() {
		return si.sp.Output(si.Game, nil)
	}
	if err := si.Game.PlayerPlay(cardIndex); err != nil {
		return si.sp.Output(si.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決する。
	if si.Game.GetPhase() == domain.SheepsheadPhaseTrickEnd {
		si.Game.ResolveTrick()
		if si.Game.GetPhase() == domain.SheepsheadPhaseTrickEnd {
			si.Game.NextTrick()
		}
	}
	si.runCpuTurns()
	return si.sp.Output(si.Game, nil)
}

// NextTrick 次のトリックへ進む
func (si *SheepsheadInteractor) NextTrick() string {
	si.Game.NextTrick()
	si.runCpuTurns()
	return si.sp.Output(si.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (si *SheepsheadInteractor) NextRound() string {
	si.Game.ScoreRound()
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	si.Game.NextRound()
	si.runCpuTurns()
	return si.sp.Output(si.Game, nil)
}

// GetConfig 現在の設定を取得
func (si *SheepsheadInteractor) GetConfig() domain.SheepsheadConfig {
	return si.Game.GetConfig()
}

// Hint ヒント取得
func (si *SheepsheadInteractor) Hint() string {
	return si.sp.HintOutput(si.Game)
}

// ActionLog 棋譜を出力する
func (si *SheepsheadInteractor) ActionLog() string {
	return si.sp.ActionLogOutput(si.Game)
}

// runCpuTurns ゲーム終了・人間の手番・トリック/ラウンド終了になるまで CPU ターンを
// 実行する。ピック/埋め/呼びの各フェーズも CPU が自動進行する。
func (si *SheepsheadInteractor) runCpuTurns() {
	for i := 0; i < MaxCpuIterations; i++ {
		if si.Game.GetGameEndFlag() {
			return
		}
		switch si.Game.GetPhase() {
		case domain.SheepsheadPhasePick, domain.SheepsheadPhaseBury, domain.SheepsheadPhaseCall:
			if si.Game.IsHumanTurn() {
				return
			}
			si.Game.CpuPlay()
		case domain.SheepsheadPhasePlay:
			if si.Game.IsHumanTurn() {
				return
			}
			si.Game.CpuPlay()
			if si.Game.GetPhase() == domain.SheepsheadPhaseTrickEnd {
				si.Game.ResolveTrick()
				if si.Game.GetPhase() == domain.SheepsheadPhaseRoundEnd {
					return
				}
				si.Game.NextTrick()
			}
		default:
			// TrickEnd / RoundEnd / GameEnd は呼び出し側の操作待ち。
			return
		}
	}
}

// RestoreSheepsheadInteractor deserialises JSON into a SheepsheadInteractor.
func RestoreSheepsheadInteractor(data []byte, sp presenter.SheepsheadPresenter) (*SheepsheadInteractor, error) {
	return restoreAndBuild[domain.Sheepshead](data, func(g *domain.Sheepshead) *SheepsheadInteractor {
		return &SheepsheadInteractor{GameBase: GameBase[interfaces.SheepsheadGame]{Game: g}, sp: sp}
	})
}
