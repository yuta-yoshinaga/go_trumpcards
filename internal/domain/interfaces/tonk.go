package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// TonkGame Tonkゲームインタフェース
type TonkGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerDrawFromStock プレイヤーが山札からカードを引く
	PlayerDrawFromStock() error
	// PlayerDrawFromDiscard プレイヤーが捨て札からカードを引く
	PlayerDrawFromDiscard() error
	// PlayerDiscard プレイヤーがカードを捨てる
	PlayerDiscard(cardIndex int) error
	// PlayerKnock プレイヤーがノックする
	PlayerKnock(cardIndex int) error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()
	// ScoreRound ラウンドの得点を計算する
	ScoreRound()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.TonkConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.TonkConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.TonkPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetBestDeadwood 1枚捨てて到達できる最小デッドウッドとその捨て札位置
	GetBestDeadwood(playerIdx int) (int, int)
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
	GetPlayer(i int) *domain.TonkPlayer
	// GetKnockerIdx ノックしたプレイヤーインデックスを取得する
	GetKnockerIdx() int
	// GetKnockerMelds ノッカーのメルド一覧を取得する
	GetKnockerMelds() [][]*domain.Card
	// GetKnockerDeadwood ノッカーのデッドウッドを取得する
	GetKnockerDeadwood() []*domain.Card
	// GetOpponentMelds 相手側のメルド一覧を取得する
	GetOpponentMelds() [][]*domain.Card
	// GetOpponentDeadwood 相手側のデッドウッドを取得する
	GetOpponentDeadwood() []*domain.Card
	// GetIsTonk 配牌Tonkかを返す
	GetIsTonk() bool
	// GetIsUndercut アンダーカットかを返す
	GetIsUndercut() bool
}
