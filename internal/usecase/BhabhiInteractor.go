//go:build !js || !wasm || extra4

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BhabhiInteractorIF バービーインタラクターインタフェース
type BhabhiInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.BhabhiConfig) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// GiveUp 投了する
	GiveUp() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.BhabhiConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// BhabhiInteractor バービーインタラクタークラス
type BhabhiInteractor struct {
	GameBase[interfaces.BhabhiGame]
	bp presenter.BhabhiPresenter
}

// NewBhabhiInteractor コンストラクタ
func NewBhabhiInteractor(b interfaces.BhabhiGame, bp presenter.BhabhiPresenter) *BhabhiInteractor {
	mustNotNil("BhabhiInteractor", map[string]any{"b": b, "bp": bp})
	return &BhabhiInteractor{GameBase: GameBase[interfaces.BhabhiGame]{Game: b}, bp: bp}
}

// Reset ゲーム初期化。配り終えたら人間の番まで進める。
//
// **ハンドの区切りが無い。** 配り切りの 1 ゲームで最後の 1 人が決まるので、
// next に相当するコマンドは存在しない。
func (bi *BhabhiInteractor) Reset() string {
	bi.Game.Reset()
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (bi *BhabhiInteractor) ResetWithConfig(cfg domain.BhabhiConfig) string {
	return resetWithValidatedConfig(bi.Game, bi.bp, cfg, bi.Game.SetConfig, bi.Reset)
}

// Play カードをプレイ
func (bi *BhabhiInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(bi.Game, bi.bp); blocked {
		return out
	}
	if err := bi.Game.PlayerPlay(cardIndex); err != nil {
		return bi.bp.Output(bi.Game, err)
	}
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// GiveUp 投了する
func (bi *BhabhiInteractor) GiveUp() string {
	if out, blocked := guardGameEnd(bi.Game, bi.bp); blocked {
		return out
	}
	bi.Game.GiveUp()
	return bi.bp.Output(bi.Game, nil)
}

// GetConfig 現在の設定を取得
func (bi *BhabhiInteractor) GetConfig() domain.BhabhiConfig { return bi.Game.GetConfig() }

// Hint ヒント取得
func (bi *BhabhiInteractor) Hint() string { return bi.bp.HintOutput(bi.Game) }

// ActionLog 棋譜を出力する
func (bi *BhabhiInteractor) ActionLog() string { return bi.bp.ActionLogOutput(bi.Game) }

// bhabhiMaxCpuTurns は 1 リクエストで回す CPU の手数上限。
//
// **共通の maxCpuTurnsPerCall (1000) では足りない。** 人間が上がったあとは
// CPU 同士で最後の 1 人が決まるまで一気に進むので、最悪ケースは
// 「膠着上限のトリック数 × 最大人数」になる。1000 で切ると**まだ CPU の手番の
// 盤面を返してしまい、画面がそこで固まる**。上限はドメイン側の定数から導く。
const bhabhiMaxCpuTurns = domain.BhabhiStalemateTricks * domain.BhabhiMaxPlayers * 2

// runCpuTurns 人間の手番になるかゲームが終わるまで CPU を進める
func (bi *BhabhiInteractor) runCpuTurns() {
	for turns := 0; !bi.Game.GetGameEndFlag() && !bi.Game.IsHumanTurn(); turns++ {
		if turns >= bhabhiMaxCpuTurns {
			return
		}
		bi.Game.CpuPlay()
	}
}

// RestoreBhabhiInteractor deserialises JSON into a BhabhiInteractor.
func RestoreBhabhiInteractor(data []byte, bp presenter.BhabhiPresenter) (*BhabhiInteractor, error) {
	return restoreAndBuild[domain.Bhabhi](data, func(g *domain.Bhabhi) *BhabhiInteractor {
		return NewBhabhiInteractor(g, bp)
	})
}
