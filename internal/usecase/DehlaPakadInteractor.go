//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// DehlaPakadInteractorIF はデーラ・パカドのインタラクターインタフェース。
type DehlaPakadInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化 (新規ゲーム)
	Reset() string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(config domain.DehlaPakadConfig) string
	// SelectTrump 切り札を宣言する
	SelectTrump(suit int) string
	// Play 手札を出す
	Play(cardIndex int) string
	// NextHand 次のハンドへ進む
	NextHand() string
	// GetConfig 現在の設定を返す
	GetConfig() domain.DehlaPakadConfig
	// Hint ヒントを出力する
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// DehlaPakadInteractor はデーラ・パカドのインタラクター。
type DehlaPakadInteractor struct {
	GameBase[interfaces.DehlaPakadGame]
	dp presenter.DehlaPakadPresenter
}

// NewDehlaPakadInteractor コンストラクタ。
func NewDehlaPakadInteractor(
	g interfaces.DehlaPakadGame, dp presenter.DehlaPakadPresenter,
) *DehlaPakadInteractor {
	mustNotNil("DehlaPakadInteractor", map[string]any{"g": g, "dp": dp})
	return &DehlaPakadInteractor{GameBase: GameBase[interfaces.DehlaPakadGame]{Game: g}, dp: dp}
}

// Reset ゲーム初期化 (新規ゲーム)。
func (di *DehlaPakadInteractor) Reset() string {
	di.Game.Reset()
	di.runCpuTurns()
	return di.dp.Output(di.Game, nil)
}

// ResetWithConfig 設定を変更してゲームを初期化。
func (di *DehlaPakadInteractor) ResetWithConfig(config domain.DehlaPakadConfig) string {
	return resetWithValidatedConfig(di.Game, di.dp, config, di.Game.SetConfig, di.Reset)
}

// SelectTrump 切り札を宣言する。
func (di *DehlaPakadInteractor) SelectTrump(suit int) string {
	if out, blocked := guardGameEnd(di.Game, di.dp); blocked {
		return out
	}
	err := di.Game.SelectTrump(suit)
	if err == nil {
		di.runCpuTurns()
	}
	return di.dp.Output(di.Game, err)
}

// Play 手札を出す。
func (di *DehlaPakadInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(di.Game, di.dp); blocked {
		return out
	}
	err := di.Game.PlayerPlay(cardIndex)
	if err == nil {
		di.runCpuTurns()
	}
	return di.dp.Output(di.Game, err)
}

// NextHand 次のハンドへ進む。
func (di *DehlaPakadInteractor) NextHand() string {
	if di.Game.GetGameEndFlag() {
		return di.dp.Output(di.Game, nil)
	}
	di.Game.NextHand()
	di.runCpuTurns()
	return di.dp.Output(di.Game, nil)
}

// GetConfig 現在の設定を返す。
func (di *DehlaPakadInteractor) GetConfig() domain.DehlaPakadConfig { return di.Game.GetConfig() }

// Hint ヒントを出力する。
func (di *DehlaPakadInteractor) Hint() string { return di.dp.HintOutput(di.Game) }

// ActionLog 棋譜を出力する。
func (di *DehlaPakadInteractor) ActionLog() string { return di.dp.ActionLogOutput(di.Game) }

// dehlaPakadMaxCpuIterations は runCpuTurns の防御的な反復上限。
const dehlaPakadMaxCpuIterations = 1000

// runCpuTurns は人間が決める番か、ハンド終了か、終局まで CPU を回す。
//
// **切り札の宣言も CPU の番になりうる。** 宣言させずに抜けると、盤面が
// 宣言フェーズのまま止まる。
func (di *DehlaPakadInteractor) runCpuTurns() {
	for i := 0; i < dehlaPakadMaxCpuIterations; i++ {
		if di.Game.GetGameEndFlag() || di.Game.IsHumanTurn() {
			return
		}
		switch di.Game.GetPhase() {
		case domain.DehlaPakadPhaseSelectTrump:
			di.Game.CpuSelectTrump()
		case domain.DehlaPakadPhasePlay:
			di.Game.CpuPlay()
		default:
			// ハンド終了では人間が次へ進める操作を待つ。
			return
		}
	}
}

// RestoreDehlaPakadInteractor deserialises JSON into an interactor.
func RestoreDehlaPakadInteractor(
	data []byte, dp presenter.DehlaPakadPresenter,
) (*DehlaPakadInteractor, error) {
	return restoreAndBuild[domain.DehlaPakad](data, func(g *domain.DehlaPakad) *DehlaPakadInteractor {
		return &DehlaPakadInteractor{GameBase: GameBase[interfaces.DehlaPakadGame]{Game: g}, dp: dp}
	})
}
