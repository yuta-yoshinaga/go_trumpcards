package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BigTwoInteractorIF Big Twoインタラクターインタフェース
type BigTwoInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Play 人間プレイヤーがカードを出す (または パスする)
	Play(indices []int) string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(config domain.BigTwoConfig) string
	// GetConfig 現在の設定を返す
	GetConfig() domain.BigTwoConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// BigTwoInteractor Big Twoインタラクタークラス
type BigTwoInteractor struct {
	GameBase[interfaces.BigTwoGame]
	btp presenter.BigTwoPresenter
}

// NewBigTwoInteractor コンストラクタ
func NewBigTwoInteractor(bg interfaces.BigTwoGame, btp presenter.BigTwoPresenter) *BigTwoInteractor {
	mustNotNil("BigTwoInteractor", map[string]any{"bg": bg, "btp": btp})
	return &BigTwoInteractor{
		GameBase: GameBase[interfaces.BigTwoGame]{Game: bg},
		btp:      btp,
	}
}

// Reset ゲーム初期化
func (bi *BigTwoInteractor) Reset() string {
	bi.Game.Reset()
	bi.runCpuTurns()
	return bi.btp.Output(bi.Game, nil)
}

// Play 人間プレイヤーがカードを出す (または パスする)
func (bi *BigTwoInteractor) Play(indices []int) string {
	if out, blocked := guardNotPlayable(bi.Game, bi.btp); blocked {
		return out
	}
	err := bi.Game.PlayerPlay(indices)
	if err == nil && !bi.Game.GetGameEndFlag() && !bi.Game.HasPendingAction() {
		bi.runCpuTurns()
	}
	return bi.btp.Output(bi.Game, err)
}

// GetConfig 現在の設定を返す
func (bi *BigTwoInteractor) GetConfig() domain.BigTwoConfig {
	return bi.Game.GetConfig()
}

// ResetWithConfig 設定を変更してゲームを初期化
func (bi *BigTwoInteractor) ResetWithConfig(config domain.BigTwoConfig) string {
	return resetWithValidatedConfig(bi.Game, bi.btp, config, bi.Game.SetConfig, bi.Reset)
}

// ActionLog 棋譜を出力する
func (bi *BigTwoInteractor) ActionLog() string {
	return bi.btp.ActionLogOutput(bi.Game)
}

// runCpuTurns ゲームが終わるか人間の手番になるまでCPUターンを実行
func (bi *BigTwoInteractor) runCpuTurns() {
	runCpuTurnsCapped(bi.Game, bi.Game.CpuPlay)
}

// RestoreBigTwoInteractor deserialises JSON into a BigTwoInteractor.
func RestoreBigTwoInteractor(data []byte, btp presenter.BigTwoPresenter) (*BigTwoInteractor, error) {
	return restoreAndBuild[domain.BigTwo](data, func(g *domain.BigTwo) *BigTwoInteractor {
		return &BigTwoInteractor{GameBase: GameBase[interfaces.BigTwoGame]{Game: g}, btp: btp}
	})
}
