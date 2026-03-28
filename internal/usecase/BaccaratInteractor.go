package usecase

import (
	"encoding/json"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BaccaratInteractorIF バカラインタラクターインタフェース
type BaccaratInteractorIF interface {
	// Reset ゲーム初期化
	Reset() string
	// Bet ベット
	Bet(amount, betType, ppBet, bpBet int) string
	// ClearHistory 罫線履歴クリア
	ClearHistory() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// BaccaratInteractor バカラインタラクタークラス
type BaccaratInteractor struct {
	b  interfaces.BaccaratGame
	bp presenter.BaccaratPresenter
}

// NewBaccaratInteractor コンストラクタ
func NewBaccaratInteractor(b interfaces.BaccaratGame, bp presenter.BaccaratPresenter) *BaccaratInteractor {
	mustNotNil("BaccaratInteractor", map[string]any{"b": b, "bp": bp})
	return &BaccaratInteractor{
		b:  b,
		bp: bp,
	}
}

// Reset ゲーム初期化
func (bi *BaccaratInteractor) Reset() string {
	return runAndPresent(bi.b, bi.bp, bi.b.Reset)
}

// Bet ベット
func (bi *BaccaratInteractor) Bet(amount, betType, ppBet, bpBet int) string {
	return execAndPresent(bi.b, bi.bp, func() error { return bi.b.Bet(amount, betType, ppBet, bpBet) })
}

// ClearHistory 罫線履歴クリア
func (bi *BaccaratInteractor) ClearHistory() string {
	return runAndPresent(bi.b, bi.bp, bi.b.ClearHistory)
}

// ActionLog 棋譜を出力する
func (bi *BaccaratInteractor) ActionLog() string {
	return bi.bp.ActionLogOutput(bi.b)
}

// Snapshot serialises the game state to JSON for KV persistence.
func (bi *BaccaratInteractor) Snapshot() ([]byte, error) {
	return json.Marshal(bi.b)
}

// RestoreBaccaratInteractor deserialises JSON into a BaccaratInteractor.
func RestoreBaccaratInteractor(data []byte, bp presenter.BaccaratPresenter) (*BaccaratInteractor, error) {
	var bac domain.Baccarat
	if err := json.Unmarshal(data, &bac); err != nil {
		return nil, err
	}
	return &BaccaratInteractor{b: &bac, bp: bp}, nil
}
