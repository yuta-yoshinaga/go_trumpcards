//go:build !js || !wasm || extra4

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// PiedmonteseTarotInteractorIF はピエモンテ・タロッコのインタラクターインタフェース。
type PiedmonteseTarotInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.PiedmonteseTarotConfig) string
	// Discard スカルトでタロンぶんを捨てる
	Discard(cardIndices []int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ディールを精算して次のディールへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.PiedmonteseTarotConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// PiedmonteseTarotInteractor はピエモンテ・タロッコのインタラクター。
type PiedmonteseTarotInteractor struct {
	GameBase[interfaces.PiedmonteseTarotGame]
	tp presenter.PiedmonteseTarotPresenter
}

// NewPiedmonteseTarotInteractor コンストラクタ。
func NewPiedmonteseTarotInteractor(
	g interfaces.PiedmonteseTarotGame, tp presenter.PiedmonteseTarotPresenter,
) *PiedmonteseTarotInteractor {
	mustNotNil("PiedmonteseTarotInteractor", map[string]any{"g": g, "tp": tp})
	return &PiedmonteseTarotInteractor{
		GameBase: GameBase[interfaces.PiedmonteseTarotGame]{Game: g}, tp: tp,
	}
}

// piedmonteseTarotTrickPhases はトリック進行のフェーズ定数。
func piedmonteseTarotTrickPhases() trickPhases[domain.PiedmonteseTarotPhase] {
	return trickPhases[domain.PiedmonteseTarotPhase]{
		play:     domain.PiedmonteseTarotPhasePlay,
		trickEnd: domain.PiedmonteseTarotPhaseTrickEnd,
		roundEnd: domain.PiedmonteseTarotPhaseRoundEnd,
		gameEnd:  domain.PiedmonteseTarotPhaseGameEnd,
	}
}

// Reset ゲーム初期化。
func (ci *PiedmonteseTarotInteractor) Reset() string {
	ci.Game.Reset()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化。
func (ci *PiedmonteseTarotInteractor) ResetWithConfig(cfg domain.PiedmonteseTarotConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.tp, cfg, ci.Game.SetConfig, ci.Reset)
}

// Discard スカルトでタロンぶんを捨てる。
func (ci *PiedmonteseTarotInteractor) Discard(cardIndices []int) string {
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.PlayerScarto(cardIndices); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// Play カードをプレイ。
func (ci *PiedmonteseTarotInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.PlayerPlay(cardIndex); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	// 人間が最後の 1 枚を出してトリックが揃ったら、その場で解決する。
	if ci.Game.GetPhase() == domain.PiedmonteseTarotPhaseTrickEnd {
		ci.Game.ResolveTrick()
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// NextTrick 次のトリックへ進む。
func (ci *PiedmonteseTarotInteractor) NextTrick() string {
	ci.Game.NextTrick()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// NextRound ディールを精算して次のディールへ進む。
func (ci *PiedmonteseTarotInteractor) NextRound() string {
	ci.Game.ScoreRound()
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	ci.Game.NextRound()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を取得。
func (ci *PiedmonteseTarotInteractor) GetConfig() domain.PiedmonteseTarotConfig {
	return ci.Game.GetConfig()
}

// Hint ヒント取得。
func (ci *PiedmonteseTarotInteractor) Hint() string { return ci.tp.HintOutput(ci.Game) }

// ActionLog 棋譜を出力する。
func (ci *PiedmonteseTarotInteractor) ActionLog() string { return ci.tp.ActionLogOutput(ci.Game) }

// advance は CPU の親スカルトとプレイを、人間の手番かトリック/ディール終了まで進める。
func (ci *PiedmonteseTarotInteractor) advance() {
	runCpuTurnsUntil(ci.Game, func() bool {
		return ci.Game.GetPhase() != domain.PiedmonteseTarotPhaseScarto || ci.Game.IsHumanScartoTurn()
	}, ci.Game.CpuScarto)
	runCpuTurnsLoop(ci.Game, piedmonteseTarotTrickPhases())
}

// RestorePiedmonteseTarotInteractor deserialises JSON into an interactor.
func RestorePiedmonteseTarotInteractor(
	data []byte, tp presenter.PiedmonteseTarotPresenter,
) (*PiedmonteseTarotInteractor, error) {
	return restoreAndBuild[domain.PiedmonteseTarot](data, func(g *domain.PiedmonteseTarot) *PiedmonteseTarotInteractor {
		return &PiedmonteseTarotInteractor{
			GameBase: GameBase[interfaces.PiedmonteseTarotGame]{Game: g}, tp: tp,
		}
	})
}
