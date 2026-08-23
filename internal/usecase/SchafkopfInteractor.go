//go:build !js || !wasm || extra4

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SchafkopfInteractorIF シャーフコップのインタラクターインタフェース
type SchafkopfInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.SchafkopfConfig) string
	// Pick ピック(true)/パス(false)を選択
	Declare(pick bool, contract domain.SchafkopfContract, soloSuit int) string
	// Call 呼びスートを指定
	Call(suit int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.SchafkopfConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// SchafkopfInteractor シャーフコップのインタラクタークラス
type SchafkopfInteractor struct {
	GameBase[interfaces.SchafkopfGame]
	sp presenter.SchafkopfPresenter
}

// NewSchafkopfInteractor コンストラクタ
func NewSchafkopfInteractor(g interfaces.SchafkopfGame, sp presenter.SchafkopfPresenter) *SchafkopfInteractor {
	mustNotNil("SchafkopfInteractor", map[string]any{"g": g, "sp": sp})
	return &SchafkopfInteractor{GameBase: GameBase[interfaces.SchafkopfGame]{Game: g}, sp: sp}
}

// Reset ゲーム初期化
func (si *SchafkopfInteractor) Reset() string {
	si.Game.Reset()
	si.runCpuTurns()
	return si.sp.Output(si.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (si *SchafkopfInteractor) ResetWithConfig(cfg domain.SchafkopfConfig) string {
	return resetWithValidatedConfig(si.Game, si.sp, cfg, si.Game.SetConfig, si.Reset)
}

// Declare 契約を宣言する(true)/パスする(false)
func (si *SchafkopfInteractor) Declare(pick bool, contract domain.SchafkopfContract, soloSuit int) string {
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	if err := si.Game.PlayerDeclare(pick, contract, soloSuit); err != nil {
		return si.sp.Output(si.Game, err)
	}
	si.runCpuTurns()
	return si.sp.Output(si.Game, nil)
}

// Call 呼びスートを指定
func (si *SchafkopfInteractor) Call(suit int) string {
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
func (si *SchafkopfInteractor) Play(cardIndex int) string {
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	if si.Game.GetPhase() != domain.SchafkopfPhasePlay || !si.Game.IsHumanTurn() {
		return si.sp.Output(si.Game, nil)
	}
	if err := si.Game.PlayerPlay(cardIndex); err != nil {
		return si.sp.Output(si.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決する。
	if si.Game.GetPhase() == domain.SchafkopfPhaseTrickEnd {
		si.Game.ResolveTrick()
		if si.Game.GetPhase() == domain.SchafkopfPhaseTrickEnd {
			si.Game.NextTrick()
		}
	}
	si.runCpuTurns()
	return si.sp.Output(si.Game, nil)
}

// NextTrick 次のトリックへ進む
func (si *SchafkopfInteractor) NextTrick() string {
	si.Game.NextTrick()
	si.runCpuTurns()
	return si.sp.Output(si.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (si *SchafkopfInteractor) NextRound() string {
	si.Game.ScoreRound()
	if out, blocked := guardGameEnd(si.Game, si.sp); blocked {
		return out
	}
	si.Game.NextRound()
	si.runCpuTurns()
	return si.sp.Output(si.Game, nil)
}

// GetConfig 現在の設定を取得
func (si *SchafkopfInteractor) GetConfig() domain.SchafkopfConfig {
	return si.Game.GetConfig()
}

// Hint ヒント取得
func (si *SchafkopfInteractor) Hint() string {
	return si.sp.HintOutput(si.Game)
}

// ActionLog 棋譜を出力する
func (si *SchafkopfInteractor) ActionLog() string {
	return si.sp.ActionLogOutput(si.Game)
}

// runCpuTurns ゲーム終了・人間の手番・トリック/ラウンド終了になるまで CPU ターンを
// 実行する。ピック/埋め/呼びの各フェーズも CPU が自動進行する。
func (si *SchafkopfInteractor) runCpuTurns() {
	for i := 0; i < MaxCpuIterations; i++ {
		if si.Game.GetGameEndFlag() {
			return
		}
		switch si.Game.GetPhase() {
		case domain.SchafkopfPhasePick, domain.SchafkopfPhaseCall:
			if si.Game.IsHumanTurn() {
				return
			}
			si.Game.CpuPlay()
		case domain.SchafkopfPhasePlay:
			if si.Game.IsHumanTurn() {
				return
			}
			si.Game.CpuPlay()
			if si.Game.GetPhase() == domain.SchafkopfPhaseTrickEnd {
				si.Game.ResolveTrick()
				if si.Game.GetPhase() == domain.SchafkopfPhaseRoundEnd {
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

// RestoreSchafkopfInteractor deserialises JSON into a SchafkopfInteractor.
func RestoreSchafkopfInteractor(data []byte, sp presenter.SchafkopfPresenter) (*SchafkopfInteractor, error) {
	return restoreAndBuild[domain.Schafkopf](data, func(g *domain.Schafkopf) *SchafkopfInteractor {
		return &SchafkopfInteractor{GameBase: GameBase[interfaces.SchafkopfGame]{Game: g}, sp: sp}
	})
}
