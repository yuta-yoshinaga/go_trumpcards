//go:build !js || !wasm || classic

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// EstimationInteractorIF エスティメーションインタラクターインタフェース
type EstimationInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.EstimationConfig) string
	// SelectTrump 切り札スートを決める
	SelectTrump(suit int) string
	// Bid 獲得予定トリック数を宣言する
	Bid(bid int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GiveUp 投了する
	GiveUp() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.EstimationConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// EstimationInteractor エスティメーションインタラクタークラス
type EstimationInteractor struct {
	GameBase[interfaces.EstimationGame]
	ep presenter.EstimationPresenter
}

// NewEstimationInteractor コンストラクタ
func NewEstimationInteractor(e interfaces.EstimationGame, ep presenter.EstimationPresenter) *EstimationInteractor {
	mustNotNil("EstimationInteractor", map[string]any{"e": e, "ep": ep})
	return &EstimationInteractor{GameBase: GameBase[interfaces.EstimationGame]{Game: e}, ep: ep}
}

// Reset ゲーム初期化。配り終えたら人間の番まで進める。
//
// **切り札選択・宣言・プレイの 3 つを順に進める。** どれか 1 つでも省くと、
// CPU の親が切り札を決めた直後や、CPU の宣言が終わった直後で止まった盤面を
// 返してしまい、人間の手番が来ない。
func (ei *EstimationInteractor) Reset() string {
	ei.Game.Reset()
	ei.advanceToHuman()
	return ei.ep.Output(ei.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ei *EstimationInteractor) ResetWithConfig(cfg domain.EstimationConfig) string {
	return resetWithValidatedConfig(ei.Game, ei.ep, cfg, ei.Game.SetConfig, ei.Reset)
}

// SelectTrump 切り札スートを決める
func (ei *EstimationInteractor) SelectTrump(suit int) string {
	return ei.act(func() error { return ei.Game.SelectTrump(suit) })
}

// Bid 獲得予定トリック数を宣言する
func (ei *EstimationInteractor) Bid(bid int) string {
	return ei.act(func() error { return ei.Game.PlayerBid(bid) })
}

// act 切り札選択 / 宣言の共通処理
func (ei *EstimationInteractor) act(fn func() error) string {
	if out, blocked := guardGameEnd(ei.Game, ei.ep); blocked {
		return out
	}
	if err := fn(); err != nil {
		return ei.ep.Output(ei.Game, err)
	}
	ei.advanceToHuman()
	return ei.ep.Output(ei.Game, nil)
}

// Play カードをプレイ
func (ei *EstimationInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ei.Game, ei.ep); blocked {
		return out
	}
	if err := ei.Game.PlayerPlay(cardIndex); err != nil {
		return ei.ep.Output(ei.Game, err)
	}
	ei.runCpuTurns()
	return ei.ep.Output(ei.Game, nil)
}

// NextRound 次のラウンドへ進む
func (ei *EstimationInteractor) NextRound() string {
	if out, blocked := guardGameEnd(ei.Game, ei.ep); blocked {
		return out
	}
	ei.Game.NextRound()
	// 次のラウンドも切り札選択から始まるので、人間の番まで進める。
	ei.advanceToHuman()
	return ei.ep.Output(ei.Game, nil)
}

// GiveUp 投了する
func (ei *EstimationInteractor) GiveUp() string {
	if out, blocked := guardGameEnd(ei.Game, ei.ep); blocked {
		return out
	}
	ei.Game.GiveUp()
	return ei.ep.Output(ei.Game, nil)
}

// GetConfig 現在の設定を取得
func (ei *EstimationInteractor) GetConfig() domain.EstimationConfig { return ei.Game.GetConfig() }

// Hint ヒント取得
func (ei *EstimationInteractor) Hint() string { return ei.ep.HintOutput(ei.Game) }

// ActionLog 棋譜を出力する
func (ei *EstimationInteractor) ActionLog() string { return ei.ep.ActionLogOutput(ei.Game) }

// advanceToHuman 切り札選択 → 宣言 → プレイ の順に、人間の番まで CPU を進める
func (ei *EstimationInteractor) advanceToHuman() {
	ei.runCpuTrump()
	ei.runCpuBids()
	ei.runCpuTurns()
}

// runCpuTrump CPU の親なら切り札を決めさせる
func (ei *EstimationInteractor) runCpuTrump() {
	if ei.Game.GetGameEndFlag() {
		return
	}
	if ei.Game.GetPhase() != domain.EstimationPhaseTrump || ei.Game.IsHumanTrumpTurn() {
		return
	}
	ei.Game.CpuSelectTrump()
}

// runCpuBids 宣言フェーズのあいだ、人間の番になるまで CPU に宣言させる。
//
// **宣言は 4 人で必ず終わる。** それでも上限を置くのは進まない CpuBid で
// ハングしないため (#4607 と同じ理由)。
func (ei *EstimationInteractor) runCpuBids() {
	for turns := 0; !ei.Game.GetGameEndFlag(); turns++ {
		if turns >= maxCpuTurnsPerCall {
			return
		}
		if ei.Game.GetPhase() != domain.EstimationPhaseBid || ei.Game.IsHumanBidTurn() {
			return
		}
		ei.Game.CpuBid()
	}
}

// runCpuTurns 人間の手番になるかラウンド／ゲームが終わるまで CPU を進める
func (ei *EstimationInteractor) runCpuTurns() {
	for turns := 0; !ei.Game.GetGameEndFlag() && !ei.Game.IsHumanTurn(); turns++ {
		if turns >= maxCpuTurnsPerCall {
			return
		}
		if ei.Game.GetPhase() != domain.EstimationPhasePlay {
			return
		}
		ei.Game.CpuPlay()
	}
}

// RestoreEstimationInteractor deserialises JSON into an EstimationInteractor.
func RestoreEstimationInteractor(data []byte, ep presenter.EstimationPresenter) (*EstimationInteractor, error) {
	return restoreAndBuild[domain.Estimation](data, func(g *domain.Estimation) *EstimationInteractor {
		return &EstimationInteractor{GameBase: GameBase[interfaces.EstimationGame]{Game: g}, ep: ep}
	})
}
