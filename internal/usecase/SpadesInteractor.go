package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SpadesInteractorIF スペードインタラクターインタフェース
type SpadesInteractorIF interface {
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.SpadesConfig) string
	// Bid ビッドを宣言
	Bid(bid int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.SpadesConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// SpadesInteractor スペードインタラクタークラス
type SpadesInteractor struct {
	s  interfaces.SpadesGame
	sp presenter.SpadesPresenter
}

// NewSpadesInteractor コンストラクタ
func NewSpadesInteractor(s interfaces.SpadesGame, sp presenter.SpadesPresenter) *SpadesInteractor {
	mustNotNil("SpadesInteractor", map[string]any{"s": s, "sp": sp})
	return &SpadesInteractor{s: s, sp: sp}
}

// Reset ゲーム初期化
func (si *SpadesInteractor) Reset() string {
	si.s.Reset()
	si.runCpuBids()
	if si.s.GetPhase() == domain.SpadesPhasePlay {
		si.runCpuTurns()
	}
	return si.sp.Output(si.s, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (si *SpadesInteractor) ResetWithConfig(cfg domain.SpadesConfig) string {
	if err := cfg.Validate(); err != nil {
		return si.sp.Output(si.s, err)
	}
	si.s.SetConfig(cfg)
	return si.Reset()
}

// Bid ビッドを宣言
func (si *SpadesInteractor) Bid(bid int) string {
	if out, blocked := guardGameEnd(si.s, si.sp); blocked {
		return out
	}
	err := si.s.PlayerBid(bid)
	if err != nil {
		return si.sp.Output(si.s, err)
	}
	si.runCpuBids()
	if si.s.GetPhase() == domain.SpadesPhasePlay {
		si.runCpuTurns()
	}
	return si.sp.Output(si.s, nil)
}

// Play カードをプレイ
func (si *SpadesInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(si.s, si.sp); blocked {
		return out
	}
	err := si.s.PlayerPlay(cardIndex)
	if err != nil {
		return si.sp.Output(si.s, err)
	}
	si.runCpuTurns()
	return si.sp.Output(si.s, nil)
}

// NextTrick 次のトリックへ進む
func (si *SpadesInteractor) NextTrick() string {
	si.s.NextTrick()
	si.runCpuTurns()
	return si.sp.Output(si.s, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (si *SpadesInteractor) NextRound() string {
	si.s.ScoreRound()
	if out, blocked := guardGameEnd(si.s, si.sp); blocked {
		return out
	}
	si.s.NextRound()
	si.runCpuBids()
	return si.sp.Output(si.s, nil)
}

// GetConfig 現在の設定を取得
func (si *SpadesInteractor) GetConfig() domain.SpadesConfig {
	return si.s.GetConfig()
}

// Hint ヒント取得
func (si *SpadesInteractor) Hint() string {
	return si.sp.HintOutput(si.s)
}

// ActionLog 棋譜を出力する
func (si *SpadesInteractor) ActionLog() string {
	return si.sp.ActionLogOutput(si.s)
}

// runCpuBids ゲームが終わるかヒューマンのビッド番またはビッドフェーズが終了するまでCPUビッドを実行
func (si *SpadesInteractor) runCpuBids() {
	for !si.s.GetGameEndFlag() {
		if si.s.GetPhase() != domain.SpadesPhaseBid {
			break
		}
		if si.s.IsHumanBidTurn() {
			break
		}
		si.s.CpuBid()
	}
}

// runCpuTurns ゲームが終わるか人間の手番またはトリック/ラウンド終了になるまでCPUターンを実行
func (si *SpadesInteractor) runCpuTurns() {
	for !si.s.GetGameEndFlag() {
		phase := si.s.GetPhase()
		if phase == domain.SpadesPhaseTrickEnd || phase == domain.SpadesPhaseRoundEnd || phase == domain.SpadesPhaseGameEnd {
			break
		}
		if phase != domain.SpadesPhasePlay {
			break
		}
		if si.s.IsHumanTurn() {
			break
		}
		si.s.CpuPlay()
		if si.s.GetPhase() == domain.SpadesPhaseTrickEnd {
			si.s.ResolveTrick()
			if si.s.GetPhase() == domain.SpadesPhaseRoundEnd {
				break
			}
			si.s.NextTrick()
		}
	}
}
