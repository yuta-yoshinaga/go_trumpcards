//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BalootInteractorIF バルートインタラクターインタフェース
type BalootInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.BalootConfig) string
	// DeclareSun 切り札なし (Sun) を宣言する
	DeclareSun() string
	// DeclareHokom 指定スートを切り札として Hokom を宣言する
	DeclareHokom(suit int) string
	// PassDeclaration 宣言を見送る
	PassDeclaration() string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GiveUp 投了する
	GiveUp() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.BalootConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// BalootInteractor バルートインタラクタークラス
type BalootInteractor struct {
	GameBase[interfaces.BalootGame]
	bp presenter.BalootPresenter
}

// NewBalootInteractor コンストラクタ
func NewBalootInteractor(b interfaces.BalootGame, bp presenter.BalootPresenter) *BalootInteractor {
	mustNotNil("BalootInteractor", map[string]any{"b": b, "bp": bp})
	return &BalootInteractor{GameBase: GameBase[interfaces.BalootGame]{Game: b}, bp: bp}
}

// Reset ゲーム初期化。配り終えたら人間の番まで進める。
//
// **宣言だけでなくプレイも進める。** 宣言は親の左隣（既定では CPU）から始まり、
// そこで CPU がモードを決めるとその場でプレイに移る。ここで runCpuTurns を
// 呼ばないと誰も打たないまま返り、人間の手番が永久に来ない。
func (bi *BalootInteractor) Reset() string {
	bi.Game.Reset()
	bi.runCpuDeclares()
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (bi *BalootInteractor) ResetWithConfig(cfg domain.BalootConfig) string {
	return resetWithValidatedConfig(bi.Game, bi.bp, cfg, bi.Game.SetConfig, bi.Reset)
}

// DeclareSun 切り札なし (Sun) を宣言する
func (bi *BalootInteractor) DeclareSun() string {
	return bi.declare(bi.Game.DeclareSun)
}

// DeclareHokom 指定スートを切り札として Hokom を宣言する
func (bi *BalootInteractor) DeclareHokom(suit int) string {
	return bi.declare(func() error { return bi.Game.DeclareHokom(suit) })
}

// PassDeclaration 宣言を見送る
func (bi *BalootInteractor) PassDeclaration() string {
	return bi.declare(bi.Game.PassDeclaration)
}

// declare 宣言系コマンドの共通処理
func (bi *BalootInteractor) declare(act func() error) string {
	if out, blocked := guardGameEnd(bi.Game, bi.bp); blocked {
		return out
	}
	if err := act(); err != nil {
		return bi.bp.Output(bi.Game, err)
	}
	bi.runCpuDeclares()
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// Play カードをプレイ
func (bi *BalootInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(bi.Game, bi.bp); blocked {
		return out
	}
	if err := bi.Game.PlayerPlay(cardIndex); err != nil {
		return bi.bp.Output(bi.Game, err)
	}
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// NextRound 次のラウンドへ進む
func (bi *BalootInteractor) NextRound() string {
	if out, blocked := guardGameEnd(bi.Game, bi.bp); blocked {
		return out
	}
	bi.Game.NextRound()
	// 次のラウンドも宣言から始まるので、人間の番まで CPU を進める。
	bi.runCpuDeclares()
	bi.runCpuTurns()
	return bi.bp.Output(bi.Game, nil)
}

// GiveUp 投了する
func (bi *BalootInteractor) GiveUp() string {
	if out, blocked := guardGameEnd(bi.Game, bi.bp); blocked {
		return out
	}
	bi.Game.GiveUp()
	return bi.bp.Output(bi.Game, nil)
}

// GetConfig 現在の設定を取得
func (bi *BalootInteractor) GetConfig() domain.BalootConfig { return bi.Game.GetConfig() }

// Hint ヒント取得
func (bi *BalootInteractor) Hint() string { return bi.bp.HintOutput(bi.Game) }

// ActionLog 棋譜を出力する
func (bi *BalootInteractor) ActionLog() string { return bi.bp.ActionLogOutput(bi.Game) }

// runCpuDeclares 宣言フェーズのあいだ、人間の番になるまで CPU に決めさせる。
//
// **親は見送れないので、宣言は必ず 4 手以内に終わる。** それでも上限を置くのは
// 進まない CpuDeclare でハングしないため (#4607 と同じ理由)。
func (bi *BalootInteractor) runCpuDeclares() {
	for turns := 0; !bi.Game.GetGameEndFlag(); turns++ {
		if turns >= maxCpuTurnsPerCall {
			return
		}
		if bi.Game.GetPhase() != domain.BalootPhaseDeclare || bi.Game.IsHumanDeclareTurn() {
			return
		}
		bi.Game.CpuDeclare()
	}
}

// runCpuTurns 人間の手番になるかラウンド／ゲームが終わるまで CPU を進める
func (bi *BalootInteractor) runCpuTurns() {
	for turns := 0; !bi.Game.GetGameEndFlag() && !bi.Game.IsHumanTurn(); turns++ {
		if turns >= maxCpuTurnsPerCall {
			return
		}
		if bi.Game.GetPhase() != domain.BalootPhasePlay {
			return
		}
		bi.Game.CpuPlay()
	}
}

// RestoreBalootInteractor deserialises JSON into a BalootInteractor.
func RestoreBalootInteractor(data []byte, bp presenter.BalootPresenter) (*BalootInteractor, error) {
	return restoreAndBuild[domain.Baloot](data, func(g *domain.Baloot) *BalootInteractor {
		return &BalootInteractor{GameBase: GameBase[interfaces.BalootGame]{Game: g}, bp: bp}
	})
}
