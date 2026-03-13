package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BlackJackInteractorIF ブラックジャックインタラクターインタフェース
type BlackJackInteractorIF interface {
	Reset() string
	Hit() string
	Stand() string
	Bet(amount, ppBet, t3Bet, handCount int) string
	DoubleDown() string
	Split() string
	Insurance() string
	DeclineInsurance() string
	Surrender() string
	EarlySurrender() string
	DeclineEarlySurrender() string
	SetDeckCount(count int) string
	ToggleHint() string
	ToggleSoft17() string
	ToggleCounting() string
	ToggleDAS() string
	SetCountingSystem(system int) string
	SetDeckPenetration(penetration int) string
	SetCpuPlayerCount(count int) string
	SetSurrenderRule(rule int) string
	ResetWithConfig(dealerHitsSoft17 bool, cpuPlayerCount int, countingEnabled bool, doubleAfterSplit bool, countingSystem int, deckPenetration int, surrenderRule int) string
	ActionLog() string
}

// BlackJackInteractor ブラックジャックインタラクタークラス
type BlackJackInteractor struct {
	bj  interfaces.BlackJackGame
	bjp presenter.BlackJackPresenter
}

// NewBlackJackInteractor コンストラクタ
func NewBlackJackInteractor(bj interfaces.BlackJackGame, bjp presenter.BlackJackPresenter) *BlackJackInteractor {
	mustNotNil("BlackJackInteractor", map[string]any{"bj": bj, "bjp": bjp})
	return &BlackJackInteractor{
		bj:  bj,
		bjp: bjp,
	}
}

// Reset ゲーム初期化
func (bi *BlackJackInteractor) Reset() string {
	bi.bj.Reset()
	return bi.bjp.Output(bi.bj, nil)
}

// Hit ヒット
func (bi *BlackJackInteractor) Hit() string {
	err := bi.bj.PlayerHit()
	return bi.bjp.Output(bi.bj, err)
}

// Stand スタンド
func (bi *BlackJackInteractor) Stand() string {
	err := bi.bj.PlayerStand()
	return bi.bjp.Output(bi.bj, err)
}

// Bet ベット
func (bi *BlackJackInteractor) Bet(amount, ppBet, t3Bet, handCount int) string {
	err := bi.bj.PlayerBet(amount, ppBet, t3Bet, handCount)
	return bi.bjp.Output(bi.bj, err)
}

// DoubleDown ダブルダウン
func (bi *BlackJackInteractor) DoubleDown() string {
	err := bi.bj.PlayerDoubleDown()
	return bi.bjp.Output(bi.bj, err)
}

// Split スプリット
func (bi *BlackJackInteractor) Split() string {
	err := bi.bj.PlayerSplit()
	return bi.bjp.Output(bi.bj, err)
}

// Insurance インシュランス
func (bi *BlackJackInteractor) Insurance() string {
	err := bi.bj.PlayerInsurance()
	return bi.bjp.Output(bi.bj, err)
}

// DeclineInsurance インシュランス辞退
func (bi *BlackJackInteractor) DeclineInsurance() string {
	err := bi.bj.PlayerDeclineInsurance()
	return bi.bjp.Output(bi.bj, err)
}

// Surrender サレンダー
func (bi *BlackJackInteractor) Surrender() string {
	err := bi.bj.PlayerSurrender()
	return bi.bjp.Output(bi.bj, err)
}

// EarlySurrender アーリーサレンダー
func (bi *BlackJackInteractor) EarlySurrender() string {
	err := bi.bj.PlayerEarlySurrender()
	return bi.bjp.Output(bi.bj, err)
}

// DeclineEarlySurrender アーリーサレンダー辞退
func (bi *BlackJackInteractor) DeclineEarlySurrender() string {
	err := bi.bj.PlayerDeclineEarlySurrender()
	return bi.bjp.Output(bi.bj, err)
}

// SetDeckCount デッキ数設定
func (bi *BlackJackInteractor) SetDeckCount(count int) string {
	err := bi.bj.SetDeckCount(count)
	return bi.bjp.Output(bi.bj, err)
}

// ToggleHint ヒント表示切り替え
func (bi *BlackJackInteractor) ToggleHint() string {
	bi.bj.ToggleHint()
	return bi.bjp.Output(bi.bj, nil)
}

// applyConfig は設定の取得・変更・保存・表示を行う共通ヘルパー
func (bi *BlackJackInteractor) applyConfig(modify func(*domain.BlackJackConfig)) string {
	config := bi.bj.GetConfig()
	modify(&config)
	err := bi.bj.SetConfig(config)
	return bi.bjp.Output(bi.bj, err)
}

// ToggleSoft17 ソフト17ルール切り替え
func (bi *BlackJackInteractor) ToggleSoft17() string {
	return bi.applyConfig(func(c *domain.BlackJackConfig) { c.DealerHitsSoft17 = !c.DealerHitsSoft17 })
}

// ToggleCounting カウンティング表示切り替え
func (bi *BlackJackInteractor) ToggleCounting() string {
	return bi.applyConfig(func(c *domain.BlackJackConfig) { c.CountingEnabled = !c.CountingEnabled })
}

// ToggleDAS スプリット後ダブルダウン許可切り替え
func (bi *BlackJackInteractor) ToggleDAS() string {
	return bi.applyConfig(func(c *domain.BlackJackConfig) { c.DoubleAfterSplit = !c.DoubleAfterSplit })
}

// SetCountingSystem カウンティングシステム変更
func (bi *BlackJackInteractor) SetCountingSystem(system int) string {
	return bi.applyConfig(func(c *domain.BlackJackConfig) { c.CountingSystem = system })
}

// SetDeckPenetration デッキペネトレーション率設定
func (bi *BlackJackInteractor) SetDeckPenetration(penetration int) string {
	return bi.applyConfig(func(c *domain.BlackJackConfig) { c.DeckPenetration = penetration })
}

// SetCpuPlayerCount CPUプレイヤー数変更
func (bi *BlackJackInteractor) SetCpuPlayerCount(count int) string {
	return bi.applyConfig(func(c *domain.BlackJackConfig) { c.CpuPlayerCount = count })
}

// SetSurrenderRule サレンダールール変更
func (bi *BlackJackInteractor) SetSurrenderRule(rule int) string {
	return bi.applyConfig(func(c *domain.BlackJackConfig) { c.SurrenderRule = rule })
}

// ResetWithConfig 設定付きリセット
func (bi *BlackJackInteractor) ResetWithConfig(dealerHitsSoft17 bool, cpuPlayerCount int, countingEnabled bool, doubleAfterSplit bool, countingSystem int, deckPenetration int, surrenderRule int) string {
	bi.bj.Reset()
	err := bi.bj.SetConfig(domain.BlackJackConfig{
		DealerHitsSoft17: dealerHitsSoft17,
		CpuPlayerCount:   cpuPlayerCount,
		CountingEnabled:  countingEnabled,
		DoubleAfterSplit: doubleAfterSplit,
		CountingSystem:   countingSystem,
		DeckPenetration:  deckPenetration,
		SurrenderRule:    surrenderRule,
	})
	if err != nil {
		return bi.bjp.Output(bi.bj, err)
	}
	bi.bj.Reset()
	return bi.bjp.Output(bi.bj, nil)
}

// ActionLog 棋譜を出力する
func (bi *BlackJackInteractor) ActionLog() string {
	return bi.bjp.ActionLogOutput(bi.bj)
}
