package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// SixCardGolfGame シックスカードゴルフゲームインタフェース
type SixCardGolfGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// FlipInitial セットアップ時に伏せ札をめくる
	FlipInitial(pos int) error
	// DrawStock 山札からカードを引く
	DrawStock() error
	// DrawDiscard 捨て札からカードを引く
	DrawDiscard() error
	// SwapCard 引いたカードとグリッド位置を交換する
	SwapCard(pos int) error
	// DiscardDrawn 引いたカードを捨てる
	DiscardDrawn() error
	// FlipCard 捨て後に伏せ札をめくる
	FlipCard(pos int) error
	// SkipFlip めくりをスキップする
	SkipFlip() error
	// CpuPlay CPUが1ターン実行する
	CpuPlay()
	// ScorePlayer プレイヤーのスコアを計算する
	ScorePlayer(playerIdx int) int

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.SixCardGolfConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.SixCardGolfConfig)
	// GetGameEndFlag ゲーム終了フラグ
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズ
	GetPhase() domain.SixCardGolfPhase
	// IsHumanTurn 現在の手番が人間か
	IsHumanTurn() bool
	// GetRoundNumber 現在のラウンド番号
	GetRoundNumber() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックス
	GetCurrentPlayerIdx() int
	// GetDiscardTop 捨て札の一番上
	GetDiscardTop() *domain.Card
	// GetDrawPileCount 山札の残り枚数
	GetDrawPileCount() int
	// GetWinnerIdx 勝者インデックス
	GetWinnerIdx() int
	// GetPlayerCnt プレイヤー数
	GetPlayerCnt() int
	// GetPlayer プレイヤー取得
	GetPlayer(i int) *domain.SixCardGolfPlayer
	// GetDrawnCard 引いたカード
	GetDrawnCard() *domain.Card
	// GetDrawnFromDiscard 捨て札から引いたか
	GetDrawnFromDiscard() bool
	// GetCanFlip めくり可能か
	GetCanFlip() bool
	// GetFinalTurnTrigger 最終ターントリガーのプレイヤーインデックス
	GetFinalTurnTrigger() int
	// ShouldDrawFromDiscard 捨て札トップを引くべきか
	ShouldDrawFromDiscard() bool
	// RecommendedSwap 引いたカードの推奨交換位置 (-1=捨て) と列ペア成立可否
	RecommendedSwap() (pos int, formsPair bool)
}
