package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// GongZhuGame 拱猪（Gong Zhu）ゲームインタフェース
type GongZhuGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerExpose プレイヤーがポイントカードを公開する
	PlayerExpose(cardIndices []int) error
	// CpuExpose CPUプレイヤーの公開選択を行う
	CpuExpose()
	// ExecuteExpose 公開フェーズを終了しプレイへ移行する
	ExecuteExpose()
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
	GetConfig() domain.GongZhuConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.GongZhuConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.GongZhuPhase
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
	// GetPlayableIndices いま出せる手札の位置
	GetPlayableIndices(playerIdx int) []int
	// GetHeartsBroken ハーツブレイク済みかを返す
	GetHeartsBroken() bool
	// GetExposure 公開状態を取得する
	GetExposure() domain.GongZhuExposure
	// GetExposeReady 各プレイヤーの公開準備状態を取得する
	GetExposeReady() [domain.GongZhuPlayerCnt]bool
	// GetExposableIndices 公開できるカードのインデックスを取得する
	GetExposableIndices(playerIdx int) []int
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetWinnerIdx 勝者インデックスを取得する
	GetWinnerIdx() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.GongZhuPlayer
	// GetHint ヒントを取得する
	GetHint() *domain.GongZhuHint
}
