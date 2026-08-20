//go:build !js || !wasm || extra4

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// CegoInteractorIF チェゴ (Cego) のインタラクターインタフェース
type CegoInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.CegoConfig) string
	// Bid 入札する
	Bid(bid domain.CegoBid) string
	// Pass パスする
	Pass() string
	// ChooseContract コントラクト (Cego / Handspiel) を選ぶ
	ChooseContract(ct domain.CegoContract) string
	// Discard Cego 交換で残す 1 枚を選ぶ
	Discard(keepIndices []int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ディールをスコアリングして次のディールへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.CegoConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// CegoInteractor チェゴのインタラクタークラス
type CegoInteractor struct {
	GameBase[interfaces.CegoGame]
	tp presenter.CegoPresenter
}

// NewCegoInteractor コンストラクタ
func NewCegoInteractor(g interfaces.CegoGame, tp presenter.CegoPresenter) *CegoInteractor {
	mustNotNil("CegoInteractor", map[string]any{"g": g, "tp": tp})
	return &CegoInteractor{GameBase: GameBase[interfaces.CegoGame]{Game: g}, tp: tp}
}

// cegoTrickPhases Cego のトリックフェーズ定数
func cegoTrickPhases() trickPhases[domain.CegoPhase] {
	return trickPhases[domain.CegoPhase]{
		play:     domain.CegoPhasePlay,
		trickEnd: domain.CegoPhaseTrickEnd,
		roundEnd: domain.CegoPhaseRoundEnd,
		gameEnd:  domain.CegoPhaseGameEnd,
	}
}

// Reset ゲーム初期化
func (ci *CegoInteractor) Reset() string {
	ci.Game.Reset()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *CegoInteractor) ResetWithConfig(cfg domain.CegoConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.tp, cfg, ci.Game.SetConfig, ci.Reset)
}

// Bid 入札する
func (ci *CegoInteractor) Bid(bid domain.CegoBid) string {
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.PlayerBid(bid); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// Pass パスする
func (ci *CegoInteractor) Pass() string {
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.PlayerPass(); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// ChooseContract コントラクト (Cego / Handspiel) を選ぶ
func (ci *CegoInteractor) ChooseContract(ct domain.CegoContract) string {
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.PlayerChooseContract(ct); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// Discard Cego 交換で残す 1 枚を選ぶ
func (ci *CegoInteractor) Discard(keepIndices []int) string {
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.PlayerDiscard(keepIndices); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// Play カードをプレイ
func (ci *CegoInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.PlayerPlay(cardIndex); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決する。
	if ci.Game.GetPhase() == domain.CegoPhaseTrickEnd {
		ci.Game.ResolveTrick()
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// NextTrick 次のトリックへ進む
func (ci *CegoInteractor) NextTrick() string {
	ci.Game.NextTrick()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// NextRound ディールをスコアリングして次のディールへ進む
func (ci *CegoInteractor) NextRound() string {
	ci.Game.ScoreRound()
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	ci.Game.NextRound()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を取得
func (ci *CegoInteractor) GetConfig() domain.CegoConfig {
	return ci.Game.GetConfig()
}

// Hint ヒント取得
func (ci *CegoInteractor) Hint() string {
	return ci.tp.HintOutput(ci.Game)
}

// ActionLog 棋譜を出力する
func (ci *CegoInteractor) ActionLog() string {
	return ci.tp.ActionLogOutput(ci.Game)
}

// advance CPU の入札・コントラクト選択・場札交換・プレイを、人間の手番もしくはトリック/ラウンド
// 終了になるまで自動実行する。
func (ci *CegoInteractor) advance() {
	runCpuBidsLoop[domain.CegoPhase](ci.Game, domain.CegoPhaseBid)
	// CPU デクレアラーのコントラクト選択を自動実行する。
	runCpuTurnsUntil(ci.Game, func() bool {
		return ci.Game.GetPhase() != domain.CegoPhaseContract || ci.Game.IsHumanContractTurn()
	}, ci.Game.CpuChooseContract)
	// CPU デクレアラーの場札交換を自動実行する。
	runCpuTurnsUntil(ci.Game, func() bool {
		return ci.Game.GetPhase() != domain.CegoPhaseExchange || ci.Game.IsHumanExchangeTurn()
	}, ci.Game.CpuDiscard)
	runCpuTurnsLoop(ci.Game, cegoTrickPhases())
}

// RestoreCegoInteractor deserialises JSON into a CegoInteractor.
func RestoreCegoInteractor(data []byte, tp presenter.CegoPresenter) (*CegoInteractor, error) {
	return restoreAndBuild[domain.Cego](data, func(g *domain.Cego) *CegoInteractor {
		return &CegoInteractor{GameBase: GameBase[interfaces.CegoGame]{Game: g}, tp: tp}
	})
}
