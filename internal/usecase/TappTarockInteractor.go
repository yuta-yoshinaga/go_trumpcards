//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// TappTarockInteractorIF タップ・タロックのインタラクターインタフェース。
type TappTarockInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.TappTarockConfig) string
	// Bid 入札する
	Bid(bid domain.TappTarockBid) string
	// Pass パスする
	Pass() string
	// Discard 場札交換で 6 枚を伏せる
	Discard(cardIndices []int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound 次のディールへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.TappTarockConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// TappTarockInteractor タップ・タロックのインタラクター。
type TappTarockInteractor struct {
	GameBase[interfaces.TappTarockGame]
	tp presenter.TappTarockPresenter
}

// NewTappTarockInteractor コンストラクタ。
func NewTappTarockInteractor(g interfaces.TappTarockGame, tp presenter.TappTarockPresenter) *TappTarockInteractor {
	mustNotNil("TappTarockInteractor", map[string]any{"g": g, "tp": tp})
	return &TappTarockInteractor{GameBase: GameBase[interfaces.TappTarockGame]{Game: g}, tp: tp}
}

// Reset ゲーム初期化。
func (zi *TappTarockInteractor) Reset() string {
	zi.Game.Reset()
	zi.advance()
	return zi.tp.Output(zi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化。
func (zi *TappTarockInteractor) ResetWithConfig(cfg domain.TappTarockConfig) string {
	return resetWithValidatedConfig(zi.Game, zi.tp, cfg, zi.Game.SetConfig, zi.Reset)
}

// Bid 入札する。
func (zi *TappTarockInteractor) Bid(bid domain.TappTarockBid) string {
	if out, blocked := guardGameEnd(zi.Game, zi.tp); blocked {
		return out
	}
	if err := zi.Game.PlayerBid(bid); err != nil {
		return zi.tp.Output(zi.Game, err)
	}
	zi.advance()
	return zi.tp.Output(zi.Game, nil)
}

// Pass パスする。
func (zi *TappTarockInteractor) Pass() string {
	if out, blocked := guardGameEnd(zi.Game, zi.tp); blocked {
		return out
	}
	if err := zi.Game.PlayerPass(); err != nil {
		return zi.tp.Output(zi.Game, err)
	}
	zi.advance()
	return zi.tp.Output(zi.Game, nil)
}

// Discard 場札交換で 6 枚を伏せる。
func (zi *TappTarockInteractor) Discard(cardIndices []int) string {
	if out, blocked := guardGameEnd(zi.Game, zi.tp); blocked {
		return out
	}
	if err := zi.Game.PlayerDiscard(cardIndices); err != nil {
		return zi.tp.Output(zi.Game, err)
	}
	zi.advance()
	return zi.tp.Output(zi.Game, nil)
}

// Play カードをプレイ。
func (zi *TappTarockInteractor) Play(cardIndex int) string {
	if out, blocked := guardGameEnd(zi.Game, zi.tp); blocked {
		return out
	}
	if err := zi.Game.PlayerPlayCard(cardIndex); err != nil {
		return zi.tp.Output(zi.Game, err)
	}
	zi.advance()
	return zi.tp.Output(zi.Game, nil)
}

// NextTrick 次のトリックへ進む。
func (zi *TappTarockInteractor) NextTrick() string {
	zi.Game.NextTrick()
	zi.advance()
	return zi.tp.Output(zi.Game, nil)
}

// NextRound 次のディールへ進む。
func (zi *TappTarockInteractor) NextRound() string {
	zi.Game.NextRound()
	zi.advance()
	return zi.tp.Output(zi.Game, nil)
}

// GetConfig 現在の設定を取得。
func (zi *TappTarockInteractor) GetConfig() domain.TappTarockConfig {
	return zi.Game.GetConfig()
}

// Hint ヒント取得。
func (zi *TappTarockInteractor) Hint() string { return zi.tp.HintOutput(zi.Game) }

// ActionLog 棋譜を出力する。
func (zi *TappTarockInteractor) ActionLog() string { return zi.tp.ActionLogOutput(zi.Game) }

// tapptarockMaxCpuSteps advance の防御的な反復上限。
const tapptarockMaxCpuSteps = 500

// advance 人間の入力が必要になるまで CPU を進める。
//
// **トリック終了では止める。** 出揃った 4 枚を見せずに次を配ると、何が起きたのか
// 分からないまま盤面が変わる。次へ進めるのは人間の操作 (NextTrick)。
func (zi *TappTarockInteractor) advance() {
	for range tapptarockMaxCpuSteps {
		if zi.Game.GetGameEndFlag() || zi.Game.IsHumanTurn() {
			return
		}
		switch zi.Game.GetPhase() {
		case domain.TappTarockPhaseBid:
			zi.Game.CpuBid()
		case domain.TappTarockPhaseTalon:
			zi.Game.CpuDiscard()
		case domain.TappTarockPhasePlay:
			zi.Game.CpuPlayCard()
		default:
			// TrickEnd / RoundEnd / GameEnd は人間の操作待ち。
			return
		}
	}
}

// RestoreTappTarockInteractor deserialises JSON into a TappTarockInteractor.
func RestoreTappTarockInteractor(data []byte, tp presenter.TappTarockPresenter) (*TappTarockInteractor, error) {
	return restoreAndBuild[domain.TappTarock](data, func(g *domain.TappTarock) *TappTarockInteractor {
		return &TappTarockInteractor{GameBase: GameBase[interfaces.TappTarockGame]{Game: g}, tp: tp}
	})
}
