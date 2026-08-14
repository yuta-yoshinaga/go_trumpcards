//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// TrogguInteractorIF トロッグのインタラクターインタフェース。
type TrogguInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.TrogguConfig) string
	// Bid 入札する
	Bid(bid domain.TrogguBid) string
	// Pass パスする
	Pass() string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound 次のディールへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.TrogguConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// TrogguInteractor トロッグのインタラクター。
type TrogguInteractor struct {
	GameBase[interfaces.TrogguGame]
	tp presenter.TrogguPresenter
}

// NewTrogguInteractor コンストラクタ。
func NewTrogguInteractor(g interfaces.TrogguGame, tp presenter.TrogguPresenter) *TrogguInteractor {
	mustNotNil("TrogguInteractor", map[string]any{"g": g, "tp": tp})
	return &TrogguInteractor{GameBase: GameBase[interfaces.TrogguGame]{Game: g}, tp: tp}
}

// Reset ゲーム初期化。
func (ti *TrogguInteractor) Reset() string {
	ti.Game.Reset()
	ti.advance()
	return ti.tp.Output(ti.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化。
func (ti *TrogguInteractor) ResetWithConfig(cfg domain.TrogguConfig) string {
	return resetWithValidatedConfig(ti.Game, ti.tp, cfg, ti.Game.SetConfig, ti.Reset)
}

// Bid 入札する。
func (ti *TrogguInteractor) Bid(bid domain.TrogguBid) string {
	if out, blocked := guardGameEnd(ti.Game, ti.tp); blocked {
		return out
	}
	if err := ti.Game.PlayerBid(bid); err != nil {
		return ti.tp.Output(ti.Game, err)
	}
	ti.advance()
	return ti.tp.Output(ti.Game, nil)
}

// Pass パスする。
func (ti *TrogguInteractor) Pass() string {
	if out, blocked := guardGameEnd(ti.Game, ti.tp); blocked {
		return out
	}
	if err := ti.Game.PlayerPass(); err != nil {
		return ti.tp.Output(ti.Game, err)
	}
	ti.advance()
	return ti.tp.Output(ti.Game, nil)
}

// Play カードをプレイ。
func (ti *TrogguInteractor) Play(cardIndex int) string {
	if out, blocked := guardGameEnd(ti.Game, ti.tp); blocked {
		return out
	}
	if err := ti.Game.PlayerPlayCard(cardIndex); err != nil {
		return ti.tp.Output(ti.Game, err)
	}
	ti.advance()
	return ti.tp.Output(ti.Game, nil)
}

// NextTrick 次のトリックへ進む。
func (ti *TrogguInteractor) NextTrick() string {
	ti.Game.NextTrick()
	ti.advance()
	return ti.tp.Output(ti.Game, nil)
}

// NextRound 次のディールへ進む。
func (ti *TrogguInteractor) NextRound() string {
	ti.Game.NextRound()
	ti.advance()
	return ti.tp.Output(ti.Game, nil)
}

// GetConfig 現在の設定を取得。
func (ti *TrogguInteractor) GetConfig() domain.TrogguConfig { return ti.Game.GetConfig() }

// Hint ヒント取得。
func (ti *TrogguInteractor) Hint() string { return ti.tp.HintOutput(ti.Game) }

// ActionLog 棋譜を出力する。
func (ti *TrogguInteractor) ActionLog() string { return ti.tp.ActionLogOutput(ti.Game) }

// trogguMaxCpuSteps advance の防御的な反復上限。
const trogguMaxCpuSteps = 500

// advance 人間の入力が必要になるまで CPU を進める。
//
// **トリック終了では止める。** 出揃った 4 枚を見せずに次を配ると、何が起きたのか
// 分からないまま盤面が変わる。次へ進めるのは人間の操作 (NextTrick)。
func (ti *TrogguInteractor) advance() {
	for range trogguMaxCpuSteps {
		if ti.Game.GetGameEndFlag() || ti.Game.IsHumanTurn() {
			return
		}
		switch ti.Game.GetPhase() {
		case domain.TrogguPhaseBid:
			ti.Game.CpuBid()
		case domain.TrogguPhasePlay:
			ti.Game.CpuPlayCard()
		default:
			// TrickEnd / RoundEnd / GameEnd は人間の操作待ち。
			return
		}
	}
}

// RestoreTrogguInteractor deserialises JSON into a TrogguInteractor.
func RestoreTrogguInteractor(data []byte, tp presenter.TrogguPresenter) (*TrogguInteractor, error) {
	return restoreAndBuild[domain.Troggu](data, func(g *domain.Troggu) *TrogguInteractor {
		return &TrogguInteractor{GameBase: GameBase[interfaces.TrogguGame]{Game: g}, tp: tp}
	})
}
