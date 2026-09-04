package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// PageOneGame ページワンゲームインタフェース
type PageOneGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// PlayerDraw プレイヤーが山札からカードを引く
	PlayerDraw() error
	// PlayerDeclare プレイヤーが「ページワン！」と宣言する
	PlayerDeclare() error
	// PlayerSkipDeclare プレイヤーが宣言をスキップする（ペナルティ）
	PlayerSkipDeclare() error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()
	// CpuDeclare CPUプレイヤーが宣言する
	CpuDeclare()
	// ScoreRound ラウンドの得点を計算する
	ScoreRound()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.PageOneConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.PageOneConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.PageOnePhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetDiscardTop 捨て札の一番上のカードを取得する
	GetDiscardTop() *domain.Card
	// GetDrawPileCount 山札の残り枚数を取得する
	GetDrawPileCount() int
	// GetWinnerIdx 勝者インデックスを取得する
	GetWinnerIdx() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.PageOnePlayer
	// IsValidPlay カードが現在の場状態でプレイ可能かを返す
	IsValidPlay(card *domain.Card) bool
	// GetRecentPenalties 直前ターンで発生したペナルティ一覧を取得する
	GetRecentPenalties() []domain.PageOnePenalty
}
