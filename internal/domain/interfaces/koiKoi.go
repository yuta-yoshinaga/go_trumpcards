//go:build !js || !wasm || extra3

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// KoiKoiGame はこいこい (Koi-Koi) のゲームインタフェース。
type KoiKoiGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerPlay 人間が手札を出す (fieldIdx で 2 枚一致時の捕獲対象を指定; 不要なら -1)
	PlayerPlay(handIdx, fieldIdx int) error
	// PlayerDecide 人間のこいこい決断 (true=こいこい, false=勝負)
	PlayerDecide(koikoi bool) error
	// CpuPlay CPU のプレイ手番を 1 ステップ実行する
	CpuPlay()
	// CpuDecide CPU のこいこい決断を 1 ステップ実行する
	CpuDecide()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.KoiKoiConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.KoiKoiConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.KoiKoiPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetCurrentTurn 現在の手番プレイヤーインデックスを取得する
	GetCurrentTurn() int
	// GetFieldCards 場札を取得する
	GetFieldCards() []*domain.Card
	// GetRemainingDeck 山札の残り枚数を取得する
	GetRemainingDeck() int
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetKoikoiCount このラウンドのこいこい宣言回数を取得する
	GetKoikoiCount() int
	// GetRoundWinner 直近ラウンドの勝者を取得する (-1=引き分け/未決)
	GetRoundWinner() int
	// GetLastRoundResult 直近ラウンド結果を取得する
	GetLastRoundResult() *domain.KoiKoiRoundResult
	// GetPendingYaku 決断フェーズで表示する成立役を取得する
	GetPendingYaku() []domain.KoiKoiYaku
	// GetPendingPoints 決断フェーズの役合計点を取得する
	GetPendingPoints() int
	// GetWinner 終局時の勝者を取得する (-1=引き分け/未決)
	GetWinner() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.KoiKoiPlayer
	// GetYaku 指定プレイヤーの現在の成立役と合計点を取得する
	GetYaku(playerIdx int) ([]domain.KoiKoiYaku, int)
	// GetPlayableIndices プレイ可能なカードのインデックスを取得する
	GetPlayableIndices(playerIdx int) []int
	// GetCaptureOptions 各手札が捕獲できる場札インデックスを取得する
	GetCaptureOptions(playerIdx int) map[int][]int
	// GetHint ヒントを取得する
	GetHint() *domain.KoiKoiHint
}
