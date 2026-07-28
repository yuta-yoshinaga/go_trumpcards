//go:build !js || !wasm || extra3

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// MaoGame マオゲームインタフェース
type MaoGame interface {
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
	// PlayerDeclare プレイヤーが「マオ！」と宣言する
	PlayerDeclare() error
	// PlayerSkipDeclare プレイヤーが宣言をスキップする
	PlayerSkipDeclare() error
	// PlayerDeclareWord プレイヤーが隠しルールに従って言葉を宣言する
	PlayerDeclareWord(word string) error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()
	// CpuChooseSuit CPUプレイヤーがスートを選択する
	CpuChooseSuit()
	// CpuDeclare CPUプレイヤーが宣言する
	CpuDeclare()
	// ScoreRound ラウンドの得点を計算する
	ScoreRound()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.MaoConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.MaoConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.MaoPhase
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
	// GetPenaltyDrawCount 累積ドローツー枚数を取得する
	GetPenaltyDrawCount() int
	// GetDirection プレイ方向を取得する (+1 / -1)
	GetDirection() int
	// GetWinnerIdx 勝者インデックスを取得する
	GetWinnerIdx() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.MaoPlayer

	// --- 隠しルール (秘密のルールはWebに公開しない) ---
	// GetAwaitingWord 人間が言葉を宣言すべき状態かを返す
	GetAwaitingWord() bool
	// GetPlayerCorrectCount 人間が隠しルールに正しく従った累計回数を返す
	GetPlayerCorrectCount() int
	// GetHintUnlocked ハーフヒントが解放されたかを返す
	GetHintUnlocked() bool
	// GetRuleHint 解放済みのハーフヒントを返す (未解放なら空文字)
	GetRuleHint() string
	// GetRulePenaltyFlag 直近のアクションで隠しルール違反が発生したかを返す
	GetRulePenaltyFlag() bool
}
