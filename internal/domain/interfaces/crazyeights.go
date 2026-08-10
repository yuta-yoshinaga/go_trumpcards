package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// CrazyEightsGame クレイジーエイトゲームインタフェース
type CrazyEightsGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// PlayerChooseSuit プレイヤーがスートを選択する
	PlayerChooseSuit(suit int) error
	// PlayerDraw プレイヤーが山札からカードを引く
	PlayerDraw() error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()
	// CpuChooseSuit CPUプレイヤーがスートを選択する
	CpuChooseSuit()
	// ScoreRound ラウンドの得点を計算する
	ScoreRound()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.CrazyEightsConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.CrazyEightsConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.CrazyEightsPhase
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
	// GetChosenSuit 選択されたスートを取得する
	GetChosenSuit() int
	// GetWinnerIdx 勝者インデックスを取得する
	GetWinnerIdx() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.CrazyEightsPlayer
	// GetHint サーバー計算の推奨手を返す (手番でなければ nil)
	GetHint() *domain.CrazyEightsHint
}
