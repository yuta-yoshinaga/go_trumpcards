//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// DramahaGame ドラマハホールデムゲームインタフェース
type DramahaGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset() error
	// PlayerAction プレイヤーのベッティングアクションを実行する
	// humanPlayMs: 迷い時間(ms, 0=計測なし)
	PlayerAction(action, amount, humanPlayMs int) error
	// Draw はドローラウンドで指定位置のホールカードを引き直す
	Draw(playerIdx int, indices []int) error
	// GetPhase 現在のフェーズを取得する
	GetPhase() int
	// GetPlayers プレイヤー一覧を取得する
	GetPlayers() []*domain.DramahaPlayer
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.DramahaPlayer
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetCommunityCards コミュニティカードを取得する
	GetCommunityCards() []*domain.Card
	// GetPot ポット額を取得する
	GetPot() int
	// GetSidePots サイドポット一覧を取得する
	GetSidePots() []domain.SidePot
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetCurrentTurn 現在のターンを取得する
	GetCurrentTurn() int
	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetLastBet 最後のベット額を取得する
	GetLastBet() int
	// GetMinRaise 最小レイズ額を取得する
	GetMinRaise() int
	// GetRaiseCount 現在のレイズ回数を取得する
	GetRaiseCount() int
	// GetRoundResults ラウンド結果を取得する
	GetRoundResults() []domain.HoldemResult
	// GetCpuActions CPU行動記録を取得する
	GetCpuActions() []domain.HoldemCpuAction
	// GetConfig ゲーム設定を取得する
	GetConfig() domain.DramahaConfig
	// SetConfig ゲーム設定を変更する
	SetConfig(cfg domain.DramahaConfig)
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetActedFlags 各プレイヤーのアクション済みフラグを取得する
	GetActedFlags() []bool
	// GetHandCount ハンド数を取得する
	GetHandCount() int
	// Resize プレイヤー数を変更する
	Resize(players []*domain.DramahaPlayer)
	TournamentActionGame
	// GetHumanProfile メタAIプロファイルを取得する
	GetHumanProfile() *domain.BettingHumanProfile
	// ResetProfile メタAIプロファイルをリセットする
	ResetProfile()
	// ExportProfile メタAIプロファイルをエクスポートする
	ExportProfile() interface{}
	// ImportProfile JSONバイトからメタAIプロファイルをインポートする
	ImportProfile(data []byte) error
	// GetEquity エクイティ計算結果を取得する
	GetEquity() *domain.HoldemEquityResult
	// GetPotOdds ポットオッズを取得する
	GetPotOdds() float64
	// GetHoleCardCount ホールカード配布枚数を取得する (ドラマハ=4, Big O=5)
	GetHoleCardCount() int
}
