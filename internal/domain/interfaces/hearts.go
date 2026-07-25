package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// HeartsGame ハーツゲームインタフェース
type HeartsGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerPass プレイヤーがカードをパスする
	PlayerPass(cardIndices []int) error
	// CpuPass CPUプレイヤーがカードをパスする
	CpuPass()
	// ExecutePass パス交換を実行する
	ExecutePass()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()
	// ResolveTrick トリックを解決する
	ResolveTrick()
	// NextTrick 次のトリックを開始する
	NextTrick()
	// ScoreRound ラウンドの得点を計算する
	ScoreRound()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.HeartsConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.HeartsConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.HeartsPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetCurrentTrick 現在のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetHeartsBroken ハーツブレイク済みかを返す
	GetHeartsBroken() bool
	// GetPassDirection パス方向を取得する
	GetPassDirection() domain.HeartsPassDirection
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetWinnerIdx 勝者インデックスを取得する
	GetWinnerIdx() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.HeartsPlayer
	// GetPassReady 各プレイヤーのパス準備状態を取得する
	GetPassReady() [domain.HeartsPlayerCnt]bool
	// GetPassedCards 各プレイヤーのパスしたカードを取得する
	GetPassedCards() [domain.HeartsPlayerCnt][]*domain.Card
	// GetHint ヒントを取得する
	GetHint() *domain.HeartsHint
}
