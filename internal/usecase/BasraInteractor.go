//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BasraInteractorIF はバスラ (Basra) のインタラクターインタフェース。
type BasraInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化 (新規ゲーム)
	Reset() string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(cfg domain.BasraConfig) string
	// Play 手札を出す (tableIdxs で捕獲対象を指定; 空ならトレイル/ジャック一掃)
	Play(handIdx int, tableIdxs []int) string
	// NextRound 次のゲームを開始する
	NextRound() string
	// GetConfig 現在の設定を返す
	GetConfig() domain.BasraConfig
	// Hint ヒントを出力する
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// BasraInteractor はバスラインタラクター。
type BasraInteractor struct {
	GameBase[interfaces.BasraGame]
	cp presenter.BasraPresenter
}

// NewBasraInteractor コンストラクタ。
func NewBasraInteractor(bg interfaces.BasraGame, cp presenter.BasraPresenter) *BasraInteractor {
	mustNotNil("BasraInteractor", map[string]any{"bg": bg, "cp": cp})
	return &BasraInteractor{GameBase: GameBase[interfaces.BasraGame]{Game: bg}, cp: cp}
}

// Reset ゲーム初期化 (新規ゲーム)。
func (bi *BasraInteractor) Reset() string {
	bi.Game.Reset()
	bi.advance()
	return bi.cp.Output(bi.Game, nil)
}

// ResetWithConfig 設定を変更してゲームを初期化。
func (bi *BasraInteractor) ResetWithConfig(cfg domain.BasraConfig) string {
	return resetWithValidatedConfig(bi.Game, bi.cp, cfg, bi.Game.SetConfig, bi.Reset)
}

// Play 手札を出す。
func (bi *BasraInteractor) Play(handIdx int, tableIdxs []int) string {
	if out, blocked := guardNotPlayable(bi.Game, bi.cp); blocked {
		return out
	}
	if err := bi.Game.PlayerPlay(handIdx, tableIdxs); err != nil {
		return bi.cp.Output(bi.Game, err)
	}
	bi.advance()
	return bi.cp.Output(bi.Game, nil)
}

// NextRound 次のゲームを開始する。バスラは山札を配り切る 1 セッションで完結するため、
// gameEnd 後の「次へ」は新規ゲーム開始として扱う。
func (bi *BasraInteractor) NextRound() string {
	bi.Game.NextRound()
	bi.advance()
	return bi.cp.Output(bi.Game, nil)
}

// GetConfig 現在の設定を返す。
func (bi *BasraInteractor) GetConfig() domain.BasraConfig { return bi.Game.GetConfig() }

// Hint ヒントを出力する。
func (bi *BasraInteractor) Hint() string { return bi.cp.HintOutput(bi.Game) }

// ActionLog 棋譜を出力する。
func (bi *BasraInteractor) ActionLog() string { return bi.cp.ActionLogOutput(bi.Game) }

// basraMaxCpuIterations は advance の防御的な反復上限。
const basraMaxCpuIterations = 1000

// advance はゲーム終了・人間の手番のいずれかに到達するまで CPU ステップを回す。
func (bi *BasraInteractor) advance() {
	for i := 0; i < basraMaxCpuIterations; i++ {
		if bi.Game.GetGameEndFlag() {
			return
		}
		if bi.Game.GetPhase() != domain.BasraPhasePlay {
			return
		}
		if bi.Game.IsHumanTurn() {
			return
		}
		bi.Game.CpuPlay()
	}
}

// RestoreBasraInteractor deserialises JSON into a BasraInteractor.
func RestoreBasraInteractor(data []byte, cp presenter.BasraPresenter) (*BasraInteractor, error) {
	return restoreAndBuild[domain.Basra](data, func(g *domain.Basra) *BasraInteractor {
		return &BasraInteractor{GameBase: GameBase[interfaces.BasraGame]{Game: g}, cp: cp}
	})
}
