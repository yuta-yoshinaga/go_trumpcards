package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// NinetyNineInteractorIF ナインティナインインタラクターインタフェース
type NinetyNineInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.NinetyNineConfig) string
	// Bid 3枚を伏せてビッドを宣言
	Bid(buryIndices []int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ディールをスコアリングして次のディールへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.NinetyNineConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// NinetyNineInteractor ナインティナインインタラクタークラス
type NinetyNineInteractor struct {
	GameBase[interfaces.NinetyNineGame]
	op presenter.NinetyNinePresenter
}

// NewNinetyNineInteractor コンストラクタ
func NewNinetyNineInteractor(o interfaces.NinetyNineGame, op presenter.NinetyNinePresenter) *NinetyNineInteractor {
	mustNotNil("NinetyNineInteractor", map[string]any{"o": o, "op": op})
	return &NinetyNineInteractor{GameBase: GameBase[interfaces.NinetyNineGame]{Game: o}, op: op}
}

// Reset ゲーム初期化
func (oi *NinetyNineInteractor) Reset() string {
	oi.Game.Reset()
	oi.runCpuBids()
	if oi.Game.GetPhase() == domain.NinetyNinePhasePlay {
		oi.runCpuTurns()
	}
	return oi.op.Output(oi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (oi *NinetyNineInteractor) ResetWithConfig(cfg domain.NinetyNineConfig) string {
	return resetWithValidatedConfig(oi.Game, oi.op, cfg, oi.Game.SetConfig, oi.Reset)
}

// Bid 3枚を伏せてビッドを宣言
func (oi *NinetyNineInteractor) Bid(buryIndices []int) string {
	if out, blocked := guardGameEnd(oi.Game, oi.op); blocked {
		return out
	}
	err := oi.Game.PlayerBid(buryIndices)
	if err != nil {
		return oi.op.Output(oi.Game, err)
	}
	oi.runCpuBids()
	if oi.Game.GetPhase() == domain.NinetyNinePhasePlay {
		oi.runCpuTurns()
	}
	return oi.op.Output(oi.Game, nil)
}

// Play カードをプレイ
func (oi *NinetyNineInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(oi.Game, oi.op); blocked {
		return out
	}
	err := oi.Game.PlayerPlay(cardIndex)
	if err != nil {
		return oi.op.Output(oi.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決する。
	// (CPUが最後に出した場合は runCpuTurns のループ内で解決される。)
	if oi.Game.GetPhase() == domain.NinetyNinePhaseTrickEnd {
		oi.Game.ResolveTrick()
	}
	oi.runCpuTurns()
	return oi.op.Output(oi.Game, nil)
}

// NextTrick 次のトリックへ進む
func (oi *NinetyNineInteractor) NextTrick() string {
	oi.Game.NextTrick()
	oi.runCpuTurns()
	return oi.op.Output(oi.Game, nil)
}

// NextRound ディールをスコアリングして次のディールへ進む
func (oi *NinetyNineInteractor) NextRound() string {
	oi.Game.ScoreRound()
	if out, blocked := guardGameEnd(oi.Game, oi.op); blocked {
		return out
	}
	oi.Game.NextRound()
	oi.runCpuBids()
	if oi.Game.GetPhase() == domain.NinetyNinePhasePlay {
		oi.runCpuTurns()
	}
	return oi.op.Output(oi.Game, nil)
}

// GetConfig 現在の設定を取得
func (oi *NinetyNineInteractor) GetConfig() domain.NinetyNineConfig {
	return oi.Game.GetConfig()
}

// Hint ヒント取得
func (oi *NinetyNineInteractor) Hint() string {
	return oi.op.HintOutput(oi.Game)
}

// ActionLog 棋譜を出力する
func (oi *NinetyNineInteractor) ActionLog() string {
	return oi.op.ActionLogOutput(oi.Game)
}

// runCpuBids ビッドフェーズでCPUのビッドを自動実行
func (oi *NinetyNineInteractor) runCpuBids() {
	runCpuBidsLoop(oi.Game, domain.NinetyNinePhaseBid)
}

// runCpuTurns ゲームが終わるか人間の手番またはトリック/ディール終了になるまでCPUターンを実行
func (oi *NinetyNineInteractor) runCpuTurns() {
	runCpuTurnsLoop(oi.Game, trickPhases[domain.NinetyNinePhase]{
		play:     domain.NinetyNinePhasePlay,
		trickEnd: domain.NinetyNinePhaseTrickEnd,
		roundEnd: domain.NinetyNinePhaseRoundEnd,
		gameEnd:  domain.NinetyNinePhaseGameEnd,
	})
}

// RestoreNinetyNineInteractor deserialises JSON into a NinetyNineInteractor.
func RestoreNinetyNineInteractor(data []byte, op presenter.NinetyNinePresenter) (*NinetyNineInteractor, error) {
	o, err := restoreGame[domain.NinetyNine](data)
	if err != nil {
		return nil, err
	}
	return &NinetyNineInteractor{GameBase: GameBase[interfaces.NinetyNineGame]{Game: o}, op: op}, nil
}
