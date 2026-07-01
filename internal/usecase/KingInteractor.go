//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// KingInteractorIF はキング (King) インタラクターインタフェース。
type KingInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化 (新規ゲーム)
	Reset() string
	// NextDeal 次のディール開始
	NextDeal() string
	// SelectContract 親がコントラクトを選択する (trumpSuit は King Trump のみ)
	SelectContract(contract, trumpSuit int) string
	// Play 手札を出す
	Play(handIdx int) string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(config domain.KingConfig) string
	// GetConfig 現在の設定を返す
	GetConfig() domain.KingConfig
	// Hint ヒントを出力する
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// KingInteractor はキングインタラクター。
type KingInteractor struct {
	GameBase[interfaces.KingGame]
	kp presenter.KingPresenter
}

// NewKingInteractor コンストラクタ。
func NewKingInteractor(kg interfaces.KingGame, kp presenter.KingPresenter) *KingInteractor {
	mustNotNil("KingInteractor", map[string]any{"kg": kg, "kp": kp})
	return &KingInteractor{
		GameBase: GameBase[interfaces.KingGame]{Game: kg},
		kp:       kp,
	}
}

// Reset ゲーム初期化 (新規ゲーム)。
func (ki *KingInteractor) Reset() string {
	ki.Game.Reset()
	ki.runCpuTurns()
	return ki.kp.Output(ki.Game, nil)
}

// NextDeal 次のディール開始。
func (ki *KingInteractor) NextDeal() string {
	if ki.Game.GetGameEndFlag() {
		return ki.kp.Output(ki.Game, nil)
	}
	ki.Game.NextDeal()
	ki.runCpuTurns()
	return ki.kp.Output(ki.Game, nil)
}

// SelectContract 親がコントラクトを選択する。
func (ki *KingInteractor) SelectContract(contract, trumpSuit int) string {
	if out, blocked := guardGameEnd(ki.Game, ki.kp); blocked {
		return out
	}
	err := ki.Game.SelectContract(contract, trumpSuit)
	if err == nil && !ki.Game.GetGameEndFlag() {
		ki.runCpuTurns()
	}
	return ki.kp.Output(ki.Game, err)
}

// Play 手札を出す。
func (ki *KingInteractor) Play(handIdx int) string {
	if out, blocked := guardNotPlayable(ki.Game, ki.kp); blocked {
		return out
	}
	err := ki.Game.PlayerPlay(handIdx)
	if err == nil && !ki.Game.GetGameEndFlag() {
		ki.runCpuTurns()
	}
	return ki.kp.Output(ki.Game, err)
}

// GetConfig 現在の設定を返す。
func (ki *KingInteractor) GetConfig() domain.KingConfig { return ki.Game.GetConfig() }

// ResetWithConfig 設定を変更してゲームを初期化。
func (ki *KingInteractor) ResetWithConfig(config domain.KingConfig) string {
	return resetWithValidatedConfig(ki.Game, ki.kp, config, ki.Game.SetConfig, ki.Reset)
}

// ActionLog 棋譜を出力する。
func (ki *KingInteractor) ActionLog() string {
	return ki.kp.ActionLogOutput(ki.Game)
}

// Hint ヒントを出力する。
func (ki *KingInteractor) Hint() string {
	return ki.kp.HintOutput(ki.Game)
}

// kingMaxCpuIterations は runCpuTurns の防御的な反復上限。
const kingMaxCpuIterations = 1000

// runCpuTurns はゲーム終了・人間の手番・ディール終了のいずれかに到達するまで
// CPU ステップを回す。ディール終了 (KingPhaseDealEnd) では人間が次ディールへ
// 進める操作を待つため、ここでは自動進行しない。
func (ki *KingInteractor) runCpuTurns() {
	for i := 0; i < kingMaxCpuIterations; i++ {
		if ki.Game.GetGameEndFlag() || ki.Game.IsHumanTurn() {
			return
		}
		if ki.Game.GetPhase() == domain.KingPhaseDealEnd {
			return
		}
		ki.Game.CpuPlay()
	}
}

// RestoreKingInteractor deserialises JSON into a KingInteractor.
func RestoreKingInteractor(data []byte, kp presenter.KingPresenter) (*KingInteractor, error) {
	return restoreAndBuild[domain.King](data, func(g *domain.King) *KingInteractor {
		return &KingInteractor{GameBase: GameBase[interfaces.KingGame]{Game: g}, kp: kp}
	})
}
