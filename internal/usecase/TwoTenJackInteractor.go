package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// TwoTenJackInteractorIF ツーテンジャックインタラクターインタフェース
type TwoTenJackInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.TwoTenJackConfig) string
	// DeclareTrump トランプを宣言
	DeclareTrump(suit int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.TwoTenJackConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// TwoTenJackInteractor ツーテンジャックインタラクタークラス
type TwoTenJackInteractor struct {
	GameBase[interfaces.TwoTenJackGame]
	tp presenter.TwoTenJackPresenter
}

// NewTwoTenJackInteractor コンストラクタ
func NewTwoTenJackInteractor(t interfaces.TwoTenJackGame, tp presenter.TwoTenJackPresenter) *TwoTenJackInteractor {
	mustNotNil("TwoTenJackInteractor", map[string]any{"t": t, "tp": tp})
	return &TwoTenJackInteractor{GameBase: GameBase[interfaces.TwoTenJackGame]{Game: t}, tp: tp}
}

// Reset ゲーム初期化
func (ti *TwoTenJackInteractor) Reset() string {
	ti.Game.Reset()
	ti.runCpuDeclares()
	if ti.Game.GetPhase() == domain.TwoTenJackPhasePlay {
		ti.runCpuTurns()
	}
	return ti.tp.Output(ti.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ti *TwoTenJackInteractor) ResetWithConfig(cfg domain.TwoTenJackConfig) string {
	return resetWithValidatedConfig(ti.Game, ti.tp, cfg, ti.Game.SetConfig, ti.Reset)
}

// DeclareTrump トランプを宣言
func (ti *TwoTenJackInteractor) DeclareTrump(suit int) string {
	if out, blocked := guardGameEnd(ti.Game, ti.tp); blocked {
		return out
	}
	err := ti.Game.PlayerDeclareTrump(suit)
	if err != nil {
		return ti.tp.Output(ti.Game, err)
	}
	if ti.Game.GetPhase() == domain.TwoTenJackPhasePlay {
		ti.runCpuTurns()
	}
	return ti.tp.Output(ti.Game, nil)
}

// Play カードをプレイ
func (ti *TwoTenJackInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ti.Game, ti.tp); blocked {
		return out
	}
	err := ti.Game.PlayerPlay(cardIndex)
	if err != nil {
		return ti.tp.Output(ti.Game, err)
	}
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// NextTrick 次のトリックへ進む
func (ti *TwoTenJackInteractor) NextTrick() string {
	ti.Game.NextTrick()
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (ti *TwoTenJackInteractor) NextRound() string {
	ti.Game.ScoreRound()
	if out, blocked := guardGameEnd(ti.Game, ti.tp); blocked {
		return out
	}
	ti.Game.NextRound()
	ti.runCpuDeclares()
	if ti.Game.GetPhase() == domain.TwoTenJackPhasePlay {
		ti.runCpuTurns()
	}
	return ti.tp.Output(ti.Game, nil)
}

// GetConfig 現在の設定を取得
func (ti *TwoTenJackInteractor) GetConfig() domain.TwoTenJackConfig {
	return ti.Game.GetConfig()
}

// Hint ヒント取得
func (ti *TwoTenJackInteractor) Hint() string {
	return ti.tp.HintOutput(ti.Game)
}

// ActionLog 棋譜を出力する
func (ti *TwoTenJackInteractor) ActionLog() string {
	return ti.tp.ActionLogOutput(ti.Game)
}

// runCpuDeclares ゲームが終わるか人間の宣言番または宣言フェーズが終了するまでCPU宣言を実行
func (ti *TwoTenJackInteractor) runCpuDeclares() {
	for i := 0; i < MaxCpuIterations; i++ {
		if ti.Game.GetGameEndFlag() {
			return
		}
		if ti.Game.GetPhase() != domain.TwoTenJackPhaseDeclare {
			break
		}
		if ti.Game.IsHumanDeclareTurn() {
			break
		}
		ti.Game.CpuDeclareTrump()
	}
}

// runCpuTurns ゲームが終わるか人間の手番またはトリック/ラウンド終了になるまでCPUターンを実行
func (ti *TwoTenJackInteractor) runCpuTurns() {
	runCpuTurnsLoop(ti.Game, trickPhases[domain.TwoTenJackPhase]{
		play:     domain.TwoTenJackPhasePlay,
		trickEnd: domain.TwoTenJackPhaseTrickEnd,
		roundEnd: domain.TwoTenJackPhaseRoundEnd,
		gameEnd:  domain.TwoTenJackPhaseGameEnd,
	})
}

// RestoreTwoTenJackInteractor deserialises JSON into a TwoTenJackInteractor.
func RestoreTwoTenJackInteractor(data []byte, tp presenter.TwoTenJackPresenter) (*TwoTenJackInteractor, error) {
	return restoreAndBuild[domain.TwoTenJack](data, func(g *domain.TwoTenJack) *TwoTenJackInteractor {
		return &TwoTenJackInteractor{GameBase: GameBase[interfaces.TwoTenJackGame]{Game: g}, tp: tp}
	})
}
