package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// BriscolaGame ブリスコラゲームインタフェース
type BriscolaGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()
	// ResolveTrick トリックを解決する
	ResolveTrick()
	// NextTrick 次のトリックを開始する (補充ドロー + ゲーム終了検出を含む)
	NextTrick()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.BriscolaConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.BriscolaConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.BriscolaPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetCurrentTrick 現在のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetTrumpSuit トランプスートを取得する
	GetTrumpSuit() int
	// GetTrumpCard 場に表向きで置かれているトランプカードを取得する (山札に残っていなければ nil)
	GetTrumpCard() *domain.Card
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetPlayerPoints プレイヤーの累積得点を取得する
	GetPlayerPoints(i int) int
	// GetWinnerIdx 勝者プレイヤーインデックスを取得する (-1: 未確定または引き分け)
	GetWinnerIdx() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.BriscolaPlayer
	// GetStockRemaining 山札の残り枚数 (場に出ている表向きトランプは含まない;
	// それは GetTrumpCard() != nil の間 別カウントとして残る最後の 1 枚)。
	GetStockRemaining() int
	// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す
	GetValidPlayIndices(playerIdx int) []int
	// GetHint ヒントを取得する
	GetHint() *domain.BriscolaHint
}
