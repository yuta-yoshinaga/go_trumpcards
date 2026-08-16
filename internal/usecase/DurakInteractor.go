package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// DurakInteractorIF ドゥラークインタラクターインタフェース
type DurakInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Attack 人間プレイヤーが攻撃カードを出す
	Attack(cardIdx int) string
	// Defend 人間プレイヤーが防御カードを出す
	Defend(attackIdx, handIdx int) string
	// Pass 人間プレイヤーがパス (攻撃停止)
	Pass() string
	// TakeCards 人間プレイヤーがカードを引き取る (防御放棄)
	TakeCards() string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(config domain.DurakConfig) string
	// GetConfig 現在の設定を返す
	GetConfig() domain.DurakConfig
	// Sort 手札ソートモードを変更
	Sort(mode domain.DurakSortMode) string
	// ActionLog 棋譜を出力する
	ActionLog() string
	// Hint サーバー計算の推奨手を出力する
	Hint() string
}

// DurakInteractor ドゥラークインタラクタークラス
type DurakInteractor struct {
	GameBase[interfaces.DurakGame]
	dp presenter.DurakPresenter
}

// NewDurakInteractor コンストラクタ
func NewDurakInteractor(d interfaces.DurakGame, dp presenter.DurakPresenter) *DurakInteractor {
	mustNotNil("DurakInteractor", map[string]any{"d": d, "dp": dp})
	return &DurakInteractor{
		GameBase: GameBase[interfaces.DurakGame]{Game: d},
		dp:       dp,
	}
}

// Reset ゲーム初期化
func (di *DurakInteractor) Reset() string {
	di.Game.Reset()
	di.runCpuTurns()
	return di.dp.Output(di.Game, nil)
}

// Attack 人間プレイヤーが攻撃カードを出す
func (di *DurakInteractor) Attack(cardIdx int) string {
	if out, blocked := guardNotPlayable(di.Game, di.dp); blocked {
		return out
	}
	err := di.Game.PlayerAttack(cardIdx)
	if err == nil && !di.Game.GetGameEndFlag() {
		di.runCpuTurns()
	}
	return di.dp.Output(di.Game, err)
}

// Defend 人間プレイヤーが防御カードを出す
func (di *DurakInteractor) Defend(attackIdx, handIdx int) string {
	if out, blocked := guardNotPlayable(di.Game, di.dp); blocked {
		return out
	}
	err := di.Game.PlayerDefend(attackIdx, handIdx)
	if err == nil && !di.Game.GetGameEndFlag() {
		di.runCpuTurns()
	}
	return di.dp.Output(di.Game, err)
}

// Pass 人間プレイヤーがパス (攻撃停止)
func (di *DurakInteractor) Pass() string {
	if out, blocked := guardNotPlayable(di.Game, di.dp); blocked {
		return out
	}
	err := di.Game.PlayerPass()
	if err == nil && !di.Game.GetGameEndFlag() {
		di.runCpuTurns()
	}
	return di.dp.Output(di.Game, err)
}

// TakeCards 人間プレイヤーがカードを引き取る (防御放棄)
func (di *DurakInteractor) TakeCards() string {
	if out, blocked := guardNotPlayable(di.Game, di.dp); blocked {
		return out
	}
	err := di.Game.PlayerTakeCards()
	if err == nil && !di.Game.GetGameEndFlag() {
		di.runCpuTurns()
	}
	return di.dp.Output(di.Game, err)
}

// GetConfig 現在の設定を返す
func (di *DurakInteractor) GetConfig() domain.DurakConfig {
	return di.Game.GetConfig()
}

// ResetWithConfig 設定を変更してゲームを初期化
func (di *DurakInteractor) ResetWithConfig(config domain.DurakConfig) string {
	return resetWithValidatedConfig(di.Game, di.dp, config, di.Game.SetConfig, di.Reset)
}

// Sort 手札ソートモードを変更
func (di *DurakInteractor) Sort(mode domain.DurakSortMode) string {
	return execAndPresent(di.Game, di.dp, func() error { return di.Game.SortHumanHand(mode) })
}

// ActionLog 棋譜を出力する
func (di *DurakInteractor) ActionLog() string {
	return di.dp.ActionLogOutput(di.Game)
}

// Hint サーバー計算の推奨手を出力する
func (di *DurakInteractor) Hint() string {
	return di.dp.HintOutput(di.Game)
}

// runCpuTurns ゲームが終わるか人間の手番になるまでCPUターンを実行
func (di *DurakInteractor) runCpuTurns() {
	runCpuTurnsCapped(di.Game, di.Game.CpuPlay)
}

// RestoreDurakInteractor deserialises JSON into a DurakInteractor.
func RestoreDurakInteractor(data []byte, dp presenter.DurakPresenter) (*DurakInteractor, error) {
	return restoreAndBuild[domain.Durak](data, func(g *domain.Durak) *DurakInteractor {
		return &DurakInteractor{GameBase: GameBase[interfaces.DurakGame]{Game: g}, dp: dp}
	})
}
