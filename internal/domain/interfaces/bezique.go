//go:build !js || !wasm || classic

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// BeziqueGame ベジーク (Bezique) ゲームインタフェース
type BeziqueGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のディールを開始する
	NextRound()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPUプレイヤーが1ターン実行する (プレイフェーズ)
	CpuPlay()
	// PlayerDeclareMeld プレイヤー (トリック勝者) が役を宣言する
	PlayerDeclareMeld(meldIndex int) error
	// PlayerSkipMeld プレイヤーが役宣言をパスする
	PlayerSkipMeld() error
	// CpuMeld 現在の役宣言手番が CPU の場合に最善メルド (なければパス) を実行する
	CpuMeld()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.BeziqueConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.BeziqueConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.BeziquePhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// IsEndgame 第2フェーズ (マストフォロー) かを返す
	IsEndgame() bool
	// GetRoundNumber 現在のディール番号を取得する
	GetRoundNumber() int
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetCurrentTrick 現在のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetTrumpSuit トランプスートを取得する
	GetTrumpSuit() int
	// GetTrumpCard 場に表向きで置かれている切り札表示カードを取得する (山札に残っていなければ nil)
	GetTrumpCard() *domain.Card
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetDealPoints プレイヤーの当ディール得点を取得する
	GetDealPoints(i int) int
	// GetDealMeldPoints プレイヤーの当ディール得点のうちメルド由来分を取得する
	GetDealMeldPoints(i int) int
	// GetMatchScore プレイヤーの試合累積得点を取得する
	GetMatchScore(i int) int
	// GetWinnerIdx 勝者プレイヤーインデックスを取得する (-1: 未確定)
	GetWinnerIdx() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.BeziquePlayer
	// GetStockRemaining 山札の残り枚数を取得する
	GetStockRemaining() int
	// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す
	GetValidPlayIndices(playerIdx int) []int
	// GetAvailableMelds トリック勝者が宣言できる役の一覧を返す
	GetAvailableMelds(playerIdx int) []domain.BeziqueMeld
	// GetHint ヒントを取得する
	GetHint() *domain.BeziqueHint
}
