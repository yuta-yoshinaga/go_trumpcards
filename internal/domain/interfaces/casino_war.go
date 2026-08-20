//go:build !js || !wasm || extra4

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// CasinoWarGame カジノウォーゲームインタフェース
type CasinoWarGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// Bet アンテをベットしカードを 1 枚ずつ配る
	Bet(amount int) error
	// ResolveInitial 初手 2 枚を比較しフェーズ遷移
	ResolveInitial()
	// Surrender タイ時の降参（アンテの半額返却）
	Surrender() error
	// GoToWar タイ時のウォー宣言（同額追加ベット＋焼き札 3 枚）
	GoToWar() error
	// ResolveWar ウォー後のカードを評価しペイアウト
	ResolveWar()

	// GetPlayerCard プレイヤーの初手カード
	GetPlayerCard() *domain.Card
	// GetDealerCard ディーラーの初手カード
	GetDealerCard() *domain.Card
	// GetPlayerWarCard プレイヤーのウォーカード
	GetPlayerWarCard() *domain.Card
	// GetDealerWarCard ディーラーのウォーカード
	GetDealerWarCard() *domain.Card
	// GetBurnCards 焼き札 3 枚
	GetBurnCards() []*domain.Card
	// GetPhase 現在のフェーズ
	GetPhase() int
	// GetGameEndFlag ゲーム終了フラグ
	GetGameEndFlag() bool
	// GetAnte アンテ額
	GetAnte() int
	// GetWarBet ウォーベット額
	GetWarBet() int
	// GetResult 勝敗結果
	GetResult() domain.GameResult
	// GetTotalPayout 合計配当
	GetTotalPayout() int
	// GetChips チップ
	GetChips() int
}
