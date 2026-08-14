//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// TarabishInteractorIF タラビッシュインタラクターインタフェース
type TarabishInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.TarabishConfig) string
	// TakeTrump 切り札を引き受ける
	TakeTrump() string
	// PassTrump 切り札を見送る
	PassTrump() string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GiveUp 投了する
	GiveUp() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.TarabishConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// TarabishInteractor タラビッシュインタラクタークラス
type TarabishInteractor struct {
	GameBase[interfaces.TarabishGame]
	tp presenter.TarabishPresenter
}

// NewTarabishInteractor コンストラクタ
func NewTarabishInteractor(t interfaces.TarabishGame, tp presenter.TarabishPresenter) *TarabishInteractor {
	mustNotNil("TarabishInteractor", map[string]any{"t": t, "tp": tp})
	return &TarabishInteractor{GameBase: GameBase[interfaces.TarabishGame]{Game: t}, tp: tp}
}

// Reset ゲーム初期化。配り終えたら人間の番まで進める。
//
// **入札だけでなくプレイも進める。** 最初の入札ラウンドで CPU が切り札を
// 引き受けると、その時点でフェーズはプレイに移る。ここで runCpuTurns を
// 呼ばないと、リード（親の左隣＝CPU）のまま誰も打たず、人間の手番が永久に
// 来ない盤面が返る。
func (ti *TarabishInteractor) Reset() string {
	ti.Game.Reset()
	ti.runCpuBids()
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ti *TarabishInteractor) ResetWithConfig(cfg domain.TarabishConfig) string {
	return resetWithValidatedConfig(ti.Game, ti.tp, cfg, ti.Game.SetConfig, ti.Reset)
}

// TakeTrump 切り札を引き受ける
func (ti *TarabishInteractor) TakeTrump() string { return ti.bid(true) }

// PassTrump 切り札を見送る
func (ti *TarabishInteractor) PassTrump() string { return ti.bid(false) }

// bid 引き受け / 見送りの共通処理
func (ti *TarabishInteractor) bid(take bool) string {
	if out, blocked := guardGameEnd(ti.Game, ti.tp); blocked {
		return out
	}
	var err error
	if take {
		err = ti.Game.TakeTrump()
	} else {
		err = ti.Game.PassTrump()
	}
	if err != nil {
		return ti.tp.Output(ti.Game, err)
	}
	ti.runCpuBids()
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// Play カードをプレイ
func (ti *TarabishInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ti.Game, ti.tp); blocked {
		return out
	}
	if err := ti.Game.PlayerPlay(cardIndex); err != nil {
		return ti.tp.Output(ti.Game, err)
	}
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// NextRound 次のラウンドへ進む
func (ti *TarabishInteractor) NextRound() string {
	if out, blocked := guardGameEnd(ti.Game, ti.tp); blocked {
		return out
	}
	ti.Game.NextRound()
	// 次のラウンドも入札から始まるので、人間の番まで CPU を進める。
	ti.runCpuBids()
	ti.runCpuTurns()
	return ti.tp.Output(ti.Game, nil)
}

// GiveUp 投了する
func (ti *TarabishInteractor) GiveUp() string {
	if out, blocked := guardGameEnd(ti.Game, ti.tp); blocked {
		return out
	}
	ti.Game.GiveUp()
	return ti.tp.Output(ti.Game, nil)
}

// GetConfig 現在の設定を取得
func (ti *TarabishInteractor) GetConfig() domain.TarabishConfig { return ti.Game.GetConfig() }

// Hint ヒント取得
func (ti *TarabishInteractor) Hint() string { return ti.tp.HintOutput(ti.Game) }

// ActionLog 棋譜を出力する
func (ti *TarabishInteractor) ActionLog() string { return ti.tp.ActionLogOutput(ti.Game) }

// runCpuBids 入札フェーズのあいだ、人間の番になるまで CPU に決めさせる。
//
// **親は見送れないので、入札は必ず 4 手以内に終わる。** それでも上限を置くのは
// 進まない CpuBid でハングしないため (#4607 と同じ理由)。
func (ti *TarabishInteractor) runCpuBids() {
	for turns := 0; !ti.Game.GetGameEndFlag(); turns++ {
		if turns >= maxCpuTurnsPerCall {
			return
		}
		if ti.Game.GetPhase() != domain.TarabishPhaseBid || ti.Game.IsHumanBidTurn() {
			return
		}
		ti.Game.CpuBid()
	}
}

// runCpuTurns 人間の手番になるかラウンド／ゲームが終わるまで CPU を進める
func (ti *TarabishInteractor) runCpuTurns() {
	for turns := 0; !ti.Game.GetGameEndFlag() && !ti.Game.IsHumanTurn(); turns++ {
		if turns >= maxCpuTurnsPerCall {
			return
		}
		if ti.Game.GetPhase() != domain.TarabishPhasePlay {
			return
		}
		ti.Game.CpuPlay()
	}
}

// RestoreTarabishInteractor deserialises JSON into a TarabishInteractor.
func RestoreTarabishInteractor(data []byte, tp presenter.TarabishPresenter) (*TarabishInteractor, error) {
	return restoreAndBuild[domain.Tarabish](data, func(g *domain.Tarabish) *TarabishInteractor {
		return &TarabishInteractor{GameBase: GameBase[interfaces.TarabishGame]{Game: g}, tp: tp}
	})
}
