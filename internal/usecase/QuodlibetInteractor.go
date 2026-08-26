//go:build !js || !wasm || solo

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// QuodlibetInteractorIF はクオドリベットのインタラクターインタフェース。
type QuodlibetInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化 (新規ゲーム)
	Reset() string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(config domain.QuodlibetConfig) string
	// SelectContract ディーラーがコントラクトを選択する
	SelectContract(contract int) string
	// Play 手札を出す (シェディング系では handIdx == -1 でパス)
	Play(handIdx int) string
	// NextDeal 次のディールへ進む
	NextDeal() string
	// GetConfig 現在の設定を返す
	GetConfig() domain.QuodlibetConfig
	// Hint ヒントを出力する
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// QuodlibetInteractor はクオドリベットのインタラクター。
type QuodlibetInteractor struct {
	GameBase[interfaces.QuodlibetGame]
	qp presenter.QuodlibetPresenter
}

// NewQuodlibetInteractor コンストラクタ。
func NewQuodlibetInteractor(
	g interfaces.QuodlibetGame, qp presenter.QuodlibetPresenter,
) *QuodlibetInteractor {
	mustNotNil("QuodlibetInteractor", map[string]any{"g": g, "qp": qp})
	return &QuodlibetInteractor{GameBase: GameBase[interfaces.QuodlibetGame]{Game: g}, qp: qp}
}

// Reset ゲーム初期化 (新規ゲーム)。
func (qi *QuodlibetInteractor) Reset() string {
	qi.Game.Reset()
	qi.runCpuTurns()
	return qi.qp.Output(qi.Game, nil)
}

// ResetWithConfig 設定を変更してゲームを初期化。
func (qi *QuodlibetInteractor) ResetWithConfig(config domain.QuodlibetConfig) string {
	return resetWithValidatedConfig(qi.Game, qi.qp, config, qi.Game.SetConfig, qi.Reset)
}

// SelectContract ディーラーがコントラクトを選択する。
func (qi *QuodlibetInteractor) SelectContract(contract int) string {
	if out, blocked := guardGameEnd(qi.Game, qi.qp); blocked {
		return out
	}
	err := qi.Game.SelectContract(contract)
	if err == nil {
		qi.runCpuTurns()
	}
	return qi.qp.Output(qi.Game, err)
}

// Play 手札を出す。
func (qi *QuodlibetInteractor) Play(handIdx int) string {
	if out, blocked := guardNotPlayable(qi.Game, qi.qp); blocked {
		return out
	}
	err := qi.Game.PlayerPlay(handIdx)
	if err == nil {
		qi.runCpuTurns()
	}
	return qi.qp.Output(qi.Game, err)
}

// NextDeal 次のディールへ進む。
func (qi *QuodlibetInteractor) NextDeal() string {
	if qi.Game.GetGameEndFlag() {
		return qi.qp.Output(qi.Game, nil)
	}
	qi.Game.NextDeal()
	qi.runCpuTurns()
	return qi.qp.Output(qi.Game, nil)
}

// GetConfig 現在の設定を返す。
func (qi *QuodlibetInteractor) GetConfig() domain.QuodlibetConfig { return qi.Game.GetConfig() }

// Hint ヒントを出力する。
func (qi *QuodlibetInteractor) Hint() string { return qi.qp.HintOutput(qi.Game) }

// ActionLog 棋譜を出力する。
func (qi *QuodlibetInteractor) ActionLog() string { return qi.qp.ActionLogOutput(qi.Game) }

// quodlibetMaxCpuIterations は runCpuTurns の防御的な反復上限。
const quodlibetMaxCpuIterations = 1000

// runCpuTurns は人間が決める番か、ディール終了か、終局まで CPU を回す。
//
// **コントラクト選択も CPU の番になりうる。** ディーラーが CPU の輪では、
// 選ばせずに抜けると盤面が選択フェーズのまま止まる。
func (qi *QuodlibetInteractor) runCpuTurns() {
	for i := 0; i < quodlibetMaxCpuIterations; i++ {
		if qi.Game.GetGameEndFlag() || qi.Game.IsHumanTurn() {
			return
		}
		switch qi.Game.GetPhase() {
		case domain.QuodlibetPhaseSelectContract:
			qi.Game.CpuSelectContract()
		case domain.QuodlibetPhasePlay:
			qi.Game.CpuPlay()
		default:
			// ディール終了では人間が次へ進める操作を待つ。
			return
		}
	}
}

// RestoreQuodlibetInteractor deserialises JSON into an interactor.
func RestoreQuodlibetInteractor(
	data []byte, qp presenter.QuodlibetPresenter,
) (*QuodlibetInteractor, error) {
	return restoreAndBuild[domain.Quodlibet](data, func(g *domain.Quodlibet) *QuodlibetInteractor {
		return &QuodlibetInteractor{GameBase: GameBase[interfaces.QuodlibetGame]{Game: g}, qp: qp}
	})
}
