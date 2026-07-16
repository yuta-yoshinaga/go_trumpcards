package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// PageOneInteractorIF ページワンインタラクターインタフェース
type PageOneInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.PageOneConfig) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// Draw カードを引く
	Draw() string
	// Declare 「ページワン！」と宣言する
	Declare() string
	// SkipDeclare 宣言をスキップしてペナルティを受ける
	SkipDeclare() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.PageOneConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
	// Hint ヒントを出力する
	Hint() string
}

// PageOneInteractor ページワンインタラクタークラス
type PageOneInteractor struct {
	GameBase[interfaces.PageOneGame]
	gp presenter.PageOnePresenter
}

// NewPageOneInteractor コンストラクタ
func NewPageOneInteractor(g interfaces.PageOneGame, gp presenter.PageOnePresenter) *PageOneInteractor {
	mustNotNil("PageOneInteractor", map[string]any{"g": g, "gp": gp})
	return &PageOneInteractor{GameBase: GameBase[interfaces.PageOneGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (ci *PageOneInteractor) Reset() string {
	ci.Game.Reset()
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *PageOneInteractor) ResetWithConfig(cfg domain.PageOneConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.gp, cfg, ci.Game.SetConfig, ci.Reset)
}

// Play カードをプレイ
func (ci *PageOneInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.PlayerPlay(cardIndex)
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// Draw カードを引く
func (ci *PageOneInteractor) Draw() string {
	if out, blocked := guardNotPlayable(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.PlayerDraw()
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// Declare 「ページワン！」と宣言する
func (ci *PageOneInteractor) Declare() string {
	if out, blocked := guardGameEnd(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.PlayerDeclare()
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// SkipDeclare 宣言をスキップしてペナルティを受ける
func (ci *PageOneInteractor) SkipDeclare() string {
	if out, blocked := guardGameEnd(ci.Game, ci.gp); blocked {
		return out
	}
	err := ci.Game.PlayerSkipDeclare()
	if err != nil {
		return ci.gp.Output(ci.Game, err)
	}
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (ci *PageOneInteractor) NextRound() string {
	ci.Game.ScoreRound()
	if out, blocked := guardGameEnd(ci.Game, ci.gp); blocked {
		return out
	}
	ci.Game.NextRound()
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を取得
func (ci *PageOneInteractor) GetConfig() domain.PageOneConfig {
	return ci.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (ci *PageOneInteractor) ActionLog() string {
	return ci.gp.ActionLogOutput(ci.Game)
}

// Hint ヒントを出力する
func (ci *PageOneInteractor) Hint() string {
	return ci.gp.HintOutput(ci.Game)
}

// runCpuTurns ゲームが終わるか人間の手番またはラウンド/ゲーム終了になるまでCPUターンを実行
func (ci *PageOneInteractor) runCpuTurns() {
	for !ci.Game.GetGameEndFlag() {
		phase := ci.Game.GetPhase()
		if phase == PageOnePhaseRoundEnd || phase == PageOnePhaseGameEnd {
			break
		}
		if phase == domain.PageOnePhaseMustDeclare {
			if ci.Game.IsHumanTurn() {
				break
			}
			ci.Game.CpuDeclare()
			continue
		}
		if phase != domain.PageOnePhasePlay {
			break
		}
		if ci.Game.IsHumanTurn() {
			break
		}
		ci.Game.CpuPlay()
	}
}

const (
	// PageOnePhaseRoundEnd ラウンド終了フェーズ (domain からの再エクスポート)
	PageOnePhaseRoundEnd = domain.PageOnePhaseRoundEnd
	// PageOnePhaseGameEnd ゲーム終了フェーズ (domain からの再エクスポート)
	PageOnePhaseGameEnd = domain.PageOnePhaseGameEnd
)

// RestorePageOneInteractor deserialises JSON into a PageOneInteractor.
func RestorePageOneInteractor(data []byte, gp presenter.PageOnePresenter) (*PageOneInteractor, error) {
	return restoreAndBuild[domain.PageOne](data, func(g *domain.PageOne) *PageOneInteractor {
		return &PageOneInteractor{GameBase: GameBase[interfaces.PageOneGame]{Game: g}, gp: gp}
	})
}
