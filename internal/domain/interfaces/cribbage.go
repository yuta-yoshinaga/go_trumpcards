//go:build !js || !wasm || extra3

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// CribbageGame クリベッジゲームインタフェース
type CribbageGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerDiscard プレイヤーがクリブに2枚捨てる
	PlayerDiscard(indices []int) error
	// PlayerCut 人間の非ディーラーがデッキをカットしてスターターを公開する
	PlayerCut() error
	// PlayerPeg プレイヤーがペギングでカードを出す
	PlayerPeg(cardIndex int) error
	// PlayerGo プレイヤーがGoを宣言する
	PlayerGo() error
	// ShowNext ショーフェーズの次のスコア計算を実行する
	ShowNext() error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.CribbageConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.CribbageConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.CribbagePhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetHint 人間プレイヤー向けの推奨アクションを取得する（対象外フェーズや手番外は nil）
	GetHint() *domain.CribbageHint
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetWinnerIdx 勝者インデックスを取得する
	GetWinnerIdx() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.CribbagePlayer
	// GetCrib クリブを取得する
	GetCrib() []*domain.Card
	// GetStarter スターターカードを取得する
	GetStarter() *domain.Card
	// GetPegCount ペギングカウントを取得する
	GetPegCount() int
	// GetPegPlayedCards ペギングで出されたカードを取得する
	GetPegPlayedCards() []*domain.Card
	// GetShowPhaseStep ショーフェーズのステップを取得する
	GetShowPhaseStep() int
	// GetHandScoreDetails ハンドスコア詳細を取得する
	GetHandScoreDetails() [3]*domain.CribbageScoreDetail
	// GetOriginalHand ショーフェーズ用の元の手札を取得する
	GetOriginalHand(playerIdx int) []*domain.Card
	// GetPlayerPeggedCards プレイヤーがペギングで出したカードを取得する
	GetPlayerPeggedCards(playerIdx int) []*domain.Card
}
