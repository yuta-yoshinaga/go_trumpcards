//go:build !js || !wasm || extra4

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// ScartoInteractorIF スカルト (Scarto) のインタラクターインタフェース
type ScartoInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.ScartoConfig) string
	// Discard スカルトで 3 枚を捨てる
	Discard(cardIndices []int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ディールをスコアリングして次のディールへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.ScartoConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// ScartoInteractor スカルトのインタラクタークラス
type ScartoInteractor struct {
	GameBase[interfaces.ScartoGame]
	tp presenter.ScartoPresenter
}

// NewScartoInteractor コンストラクタ
func NewScartoInteractor(g interfaces.ScartoGame, tp presenter.ScartoPresenter) *ScartoInteractor {
	mustNotNil("ScartoInteractor", map[string]any{"g": g, "tp": tp})
	return &ScartoInteractor{GameBase: GameBase[interfaces.ScartoGame]{Game: g}, tp: tp}
}

// scartoTrickPhases Scarto のトリックフェーズ定数
func scartoTrickPhases() trickPhases[domain.ScartoPhase] {
	return trickPhases[domain.ScartoPhase]{
		play:     domain.ScartoPhasePlay,
		trickEnd: domain.ScartoPhaseTrickEnd,
		roundEnd: domain.ScartoPhaseRoundEnd,
		gameEnd:  domain.ScartoPhaseGameEnd,
	}
}

// Reset ゲーム初期化
func (ci *ScartoInteractor) Reset() string {
	ci.Game.Reset()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *ScartoInteractor) ResetWithConfig(cfg domain.ScartoConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.tp, cfg, ci.Game.SetConfig, ci.Reset)
}

// Discard スカルトで 3 枚を捨てる
func (ci *ScartoInteractor) Discard(cardIndices []int) string {
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.PlayerScarto(cardIndices); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// Play カードをプレイ
func (ci *ScartoInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.PlayerPlay(cardIndex); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決する。
	if ci.Game.GetPhase() == domain.ScartoPhaseTrickEnd {
		ci.Game.ResolveTrick()
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// NextTrick 次のトリックへ進む
func (ci *ScartoInteractor) NextTrick() string {
	ci.Game.NextTrick()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// NextRound ディールをスコアリングして次のディールへ進む
func (ci *ScartoInteractor) NextRound() string {
	ci.Game.ScoreRound()
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	ci.Game.NextRound()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を取得
func (ci *ScartoInteractor) GetConfig() domain.ScartoConfig {
	return ci.Game.GetConfig()
}

// Hint ヒント取得
func (ci *ScartoInteractor) Hint() string {
	return ci.tp.HintOutput(ci.Game)
}

// ActionLog 棋譜を出力する
func (ci *ScartoInteractor) ActionLog() string {
	return ci.tp.ActionLogOutput(ci.Game)
}

// advance CPU の親スカルト・プレイを、人間の手番もしくはトリック/ラウンド終了になるまで
// 自動実行する。
func (ci *ScartoInteractor) advance() {
	// CPU の親のスカルトを自動実行する。
	runCpuTurnsUntil(ci.Game, func() bool {
		return ci.Game.GetPhase() != domain.ScartoPhaseScarto || ci.Game.IsHumanScartoTurn()
	}, ci.Game.CpuScarto)
	runCpuTurnsLoop(ci.Game, scartoTrickPhases())
}

// RestoreScartoInteractor deserialises JSON into a ScartoInteractor.
func RestoreScartoInteractor(data []byte, tp presenter.ScartoPresenter) (*ScartoInteractor, error) {
	return restoreAndBuild[domain.Scarto](data, func(g *domain.Scarto) *ScartoInteractor {
		return &ScartoInteractor{GameBase: GameBase[interfaces.ScartoGame]{Game: g}, tp: tp}
	})
}
