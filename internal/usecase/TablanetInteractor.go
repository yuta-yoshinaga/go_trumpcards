//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// TablanetInteractorIF はタブラネット (Tablanet) のインタラクターインタフェース。
type TablanetInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化 (新規ゲーム)
	Reset() string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(cfg domain.TablanetConfig) string
	// Play 手札を出す (tableIdxs で捕獲対象を指定; 空ならトレイル/ジャック一掃)
	Play(handIdx int, tableIdxs []int) string
	// NextRound 次のゲームを開始する
	NextRound() string
	// GetConfig 現在の設定を返す
	GetConfig() domain.TablanetConfig
	// Hint ヒントを出力する
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// TablanetInteractor はタブラネットインタラクター。
type TablanetInteractor struct {
	GameBase[interfaces.TablanetGame]
	cp presenter.TablanetPresenter
}

// NewTablanetInteractor コンストラクタ。
func NewTablanetInteractor(bg interfaces.TablanetGame, cp presenter.TablanetPresenter) *TablanetInteractor {
	mustNotNil("TablanetInteractor", map[string]any{"bg": bg, "cp": cp})
	return &TablanetInteractor{GameBase: GameBase[interfaces.TablanetGame]{Game: bg}, cp: cp}
}

// Reset ゲーム初期化 (新規ゲーム)。
func (bi *TablanetInteractor) Reset() string {
	bi.Game.Reset()
	bi.advance()
	return bi.cp.Output(bi.Game, nil)
}

// ResetWithConfig 設定を変更してゲームを初期化。
func (bi *TablanetInteractor) ResetWithConfig(cfg domain.TablanetConfig) string {
	return resetWithValidatedConfig(bi.Game, bi.cp, cfg, bi.Game.SetConfig, bi.Reset)
}

// Play 手札を出す。
func (bi *TablanetInteractor) Play(handIdx int, tableIdxs []int) string {
	if out, blocked := guardNotPlayable(bi.Game, bi.cp); blocked {
		return out
	}
	if err := bi.Game.PlayerPlay(handIdx, tableIdxs); err != nil {
		return bi.cp.Output(bi.Game, err)
	}
	bi.advance()
	return bi.cp.Output(bi.Game, nil)
}

// NextRound 次のゲームを開始する。タブラネットは山札を配り切る 1 セッションで完結するため、
// gameEnd 後の「次へ」は新規ゲーム開始として扱う。
func (bi *TablanetInteractor) NextRound() string {
	bi.Game.NextRound()
	bi.advance()
	return bi.cp.Output(bi.Game, nil)
}

// GetConfig 現在の設定を返す。
func (bi *TablanetInteractor) GetConfig() domain.TablanetConfig { return bi.Game.GetConfig() }

// Hint ヒントを出力する。
func (bi *TablanetInteractor) Hint() string { return bi.cp.HintOutput(bi.Game) }

// ActionLog 棋譜を出力する。
func (bi *TablanetInteractor) ActionLog() string { return bi.cp.ActionLogOutput(bi.Game) }

// tablanetMaxCpuIterations は advance の防御的な反復上限。
const tablanetMaxCpuIterations = 1000

// advance はゲーム終了・人間の手番のいずれかに到達するまで CPU ステップを回す。
func (bi *TablanetInteractor) advance() {
	for i := 0; i < tablanetMaxCpuIterations; i++ {
		if bi.Game.GetGameEndFlag() {
			return
		}
		if bi.Game.GetPhase() != domain.TablanetPhasePlay {
			return
		}
		if bi.Game.IsHumanTurn() {
			return
		}
		bi.Game.CpuPlay()
	}
}

// RestoreTablanetInteractor deserialises JSON into a TablanetInteractor.
func RestoreTablanetInteractor(data []byte, cp presenter.TablanetPresenter) (*TablanetInteractor, error) {
	return restoreAndBuild[domain.Tablanet](data, func(g *domain.Tablanet) *TablanetInteractor {
		return &TablanetInteractor{GameBase: GameBase[interfaces.TablanetGame]{Game: g}, cp: cp}
	})
}
