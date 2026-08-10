//go:build !js || !wasm || extra3

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// KlaberjassGame クラバーヤス (Klaberjass) ゲームインタフェース
type KlaberjassGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextDeal 次のディールを配る
	NextDeal() error
	// AcceptTrump 表向きカードのスートを切札にする
	AcceptTrump(player int) error
	// CallTrump 好きなスートを切札に指名する
	CallTrump(player, suit int) error
	// Pass ビッドを見送る
	Pass(player int) error
	// Schmeiss この配りを流すことを提案する
	Schmeiss(player int) error
	// AnswerSchmeiss 投げの提案に答える
	AnswerSchmeiss(player int, accept bool) error
	// PlayCard 手札を1枚出す
	PlayCard(player, idx int) error
	// CpuPlay CPUプレイヤーが1アクション実行する
	CpuPlay()
	// KlaberjassValidPlays 出せる手札インデックスを返す
	KlaberjassValidPlays(player int) []int

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.KlaberjassConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.KlaberjassConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.KlaberjassPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetCurrentPlayerIdx 現在の手番を取得する
	GetCurrentPlayerIdx() int
	// GetBidPlayerIdx ビッド中の手番を取得する
	GetBidPlayerIdx() int
	// GetDealerIdx ディーラーを取得する
	GetDealerIdx() int
	// GetTrumpSuit 切札スートを取得する (0 なら未確定)
	GetTrumpSuit() int
	// GetTurnUpCard 表向きカードを取得する
	GetTurnUpCard() *domain.Card
	// GetMakerIdx 切札を決めた席を取得する
	GetMakerIdx() int
	// GetTrick 場に出ている札を取得する
	GetTrick() []*domain.Card
	// GetTrickLeaderIdx このトリックのリード席を取得する
	GetTrickLeaderIdx() int
	// GetTrickNumber 済んだトリック数を取得する
	GetTrickNumber() int
	// GetHandPoints このディールで取った点を取得する
	GetHandPoints(idx int) int
	// GetSequences シーケンス役を取得する
	GetSequences(idx int) []*domain.KlaberjassSequence
	// GetSequenceWinner シーケンス勝負に勝った席を取得する
	GetSequenceWinner() int
	// GetLastTrickWinner 最終トリックを取った席を取得する (-1 ならまだ)
	GetLastTrickWinner() int
	// GetBelaHolder 切札 K+Q を持っていた席を取得する
	GetBelaHolder() int
	// IsBelaScored ベラが成立したかを取得する
	IsBelaScored() bool
	// IsDixUsed 切札の7の交換が行われたかを取得する
	IsDixUsed() bool
	// IsBete 直前のディールでメイカーがベートしたかを取得する
	IsBete() bool
	// GetSchmeissBy 投げを提案した席を取得する
	GetSchmeissBy() int
	// GetScore 通算点を取得する
	GetScore(idx int) int
	// GetDealNumber 現在のディール番号を取得する
	GetDealNumber() int
	// GetWinnerIdx 勝者を取得する
	GetWinnerIdx() int
	// GetPlayers 全プレイヤーを取得する
	GetPlayers() []*domain.KlaberjassPlayer
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(idx int) *domain.KlaberjassPlayer
}
