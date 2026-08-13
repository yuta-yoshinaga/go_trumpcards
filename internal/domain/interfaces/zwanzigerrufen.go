//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// ZwanzigerrufenGame ツヴァンツィガールーフェン (Zwanzigerrufen) のゲームインタフェース
type ZwanzigerrufenGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のディールを開始する
	NextRound()
	// PlayerBid 人間プレイヤーが入札する
	PlayerBid(bid domain.ZwanzigerrufenBid) error
	// PlayerPass 人間プレイヤーがパスする
	PlayerPass() error
	// CpuBid CPU プレイヤーが 1 回入札する
	CpuBid()
	// PlayerDiscard 人間デクレアラーが場札交換で 6 枚伏せる
	PlayerDiscard(indices []int) error
	// CpuDiscard CPU デクレアラーの伏せ札を進める
	CpuDiscard()
	// PlayerPlayCard 人間が手札を 1 枚出す
	PlayerPlayCard(handIdx int) error
	// CpuPlayCard CPU の 1 手を進める
	CpuPlayCard()
	// NextTrick トリック終了から次のトリックへ進む
	NextTrick()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.ZwanzigerrufenConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.ZwanzigerrufenConfig)

	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.ZwanzigerrufenPhase
	// IsHumanTurn 人間の入力を待っているか
	IsHumanTurn() bool
	// HumanSeat 人間の席を取得する
	HumanSeat() int
	// GetRoundNumber 現在のディール番号を取得する
	GetRoundNumber() int
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetCurrentPlayerIdx 手番の席を取得する
	GetCurrentPlayerIdx() int
	// GetCurrentTrick 場に出ている札を取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetDealerIdx 親の席を取得する
	GetDealerIdx() int
	// GetBidPlayerIdx 入札の手番を取得する
	GetBidPlayerIdx() int
	// GetHighestBid 現在の最高入札を取得する
	GetHighestBid() domain.ZwanzigerrufenBid
	// GetDeclarerIdx デクレアラーの席を取得する (-1 = 未確定 / Trischaken)
	GetDeclarerIdx() int
	// GetContract 成立した契約を取得する
	GetContract() domain.ZwanzigerrufenBid
	// GetCalledTrump 呼んだ切り札の番号を取得する (-1 = 呼んでいない)
	GetCalledTrump() int
	// GetPartnerRevealed パートナーが判明したかを取得する
	GetPartnerRevealed() bool
	// GetPartnerIdx パートナーの席を取得する (判明前は -1)
	GetPartnerIdx() int
	// GetTalonSize 場札の枚数を取得する
	GetTalonSize() int
	// GetLastTrickWinner 直前のトリックの勝者を取得する
	GetLastTrickWinner() int
	// GetLastTrickCards 直前のトリックの札を取得する
	GetLastTrickCards() []*domain.Card
	// GetOutcome ディール結果を取得する
	GetOutcome() domain.ZwanzigerrufenOutcome
	// GetBreakdown 直近ディールの精算内訳を取得する
	GetBreakdown() *domain.ZwanzigerrufenBreakdown
	// GetPlayerCnt 席数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定した席を取得する
	GetPlayer(i int) *domain.ZwanzigerrufenPlayer
	// GetPlayerScore 指定席の通算得点を取得する
	GetPlayerScore(i int) int
	// GetCardPoints 指定席が獲得したカードポイントを取得する
	GetCardPoints(i int) int
	// GetValidPlayIndices 出せる手札のインデックスを取得する
	GetValidPlayIndices(playerIdx int) []int
	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetWinnerPlayer 終局時の勝者を取得する (-1 = 引き分け/未決)
	GetWinnerPlayer() int
	// GetHint ヒントを取得する
	GetHint() *domain.ZwanzigerrufenHint
}
