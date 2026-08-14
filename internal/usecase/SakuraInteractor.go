//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SakuraInteractorIF はさくら (肥後花) のインタラクターインタフェース。
type SakuraInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化 (新規ゲーム)
	Reset() string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(cfg domain.SakuraConfig) string
	// Play 手札を出す (fieldIdx で 2 枚一致時の獲得対象を指定; 不要なら -1)
	Play(handIdx, fieldIdx int) string
	// NextRound 次のラウンドを開始する
	NextRound() string
	// GetConfig 現在の設定を返す
	GetConfig() domain.SakuraConfig
	// Hint ヒントを出力する
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// SakuraInteractor はさくらインタラクター。
type SakuraInteractor struct {
	GameBase[interfaces.SakuraGame]
	cp presenter.SakuraPresenter
}

// NewSakuraInteractor コンストラクタ。
func NewSakuraInteractor(sg interfaces.SakuraGame, cp presenter.SakuraPresenter) *SakuraInteractor {
	mustNotNil("SakuraInteractor", map[string]any{"sg": sg, "cp": cp})
	return &SakuraInteractor{GameBase: GameBase[interfaces.SakuraGame]{Game: sg}, cp: cp}
}

// Reset ゲーム初期化 (新規ゲーム)。
func (si *SakuraInteractor) Reset() string {
	si.Game.Reset()
	si.advance()
	return si.cp.Output(si.Game, nil)
}

// ResetWithConfig 設定を変更してゲームを初期化。
func (si *SakuraInteractor) ResetWithConfig(cfg domain.SakuraConfig) string {
	return resetWithValidatedConfig(si.Game, si.cp, cfg, si.Game.SetConfig, si.Reset)
}

// Play 手札を出す。
func (si *SakuraInteractor) Play(handIdx, fieldIdx int) string {
	if out, blocked := guardNotPlayable(si.Game, si.cp); blocked {
		return out
	}
	if err := si.Game.PlayerPlay(handIdx, fieldIdx); err != nil {
		return si.cp.Output(si.Game, err)
	}
	si.advance()
	return si.cp.Output(si.Game, nil)
}

// NextRound 次のラウンドを開始する。
func (si *SakuraInteractor) NextRound() string {
	si.Game.NextRound()
	si.advance()
	return si.cp.Output(si.Game, nil)
}

// GetConfig 現在の設定を返す。
func (si *SakuraInteractor) GetConfig() domain.SakuraConfig { return si.Game.GetConfig() }

// Hint ヒントを出力する。
func (si *SakuraInteractor) Hint() string { return si.cp.HintOutput(si.Game) }

// ActionLog 棋譜を出力する。
func (si *SakuraInteractor) ActionLog() string { return si.cp.ActionLogOutput(si.Game) }

// sakuraMaxCpuIterations は advance の防御的な反復上限。
//
// 1 ラウンドは席数 × 手札枚数の手番で終わるので、実際に必要なのは数十回。
// 上限は暴走を止めるためだけに置く。
const sakuraMaxCpuIterations = 1000

// advance は終局・ラウンド終了・人間の手番のいずれかに到達するまで CPU を進める。
func (si *SakuraInteractor) advance() {
	for range sakuraMaxCpuIterations {
		if si.Game.GetGameEndFlag() || si.Game.GetPhase() != domain.SakuraPhasePlay ||
			si.Game.IsHumanTurn() {
			return
		}
		si.Game.CpuPlay()
	}
}

// RestoreSakuraInteractor deserialises JSON into a SakuraInteractor.
func RestoreSakuraInteractor(data []byte, cp presenter.SakuraPresenter) (*SakuraInteractor, error) {
	return restoreAndBuild[domain.Sakura](data, func(g *domain.Sakura) *SakuraInteractor {
		return &SakuraInteractor{GameBase: GameBase[interfaces.SakuraGame]{Game: g}, cp: cp}
	})
}
