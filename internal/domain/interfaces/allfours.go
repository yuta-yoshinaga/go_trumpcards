package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// AllFoursGame All Fours (Seven Up) ゲームインタフェース
type AllFoursGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のディールを開始する
	NextRound()
	// PlayerBeg 非親が stand / beg を選ぶ (beg=true で beg)
	PlayerBeg(beg bool) error
	// CpuBeg 非親CPUが stand / beg を決める
	CpuBeg()
	// PlayerRespondBeg 親が beg に応答する (run=true で run the cards)
	PlayerRespondBeg(run bool) error
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
	GetConfig() domain.AllFoursConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.AllFoursConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.AllFoursPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetDealerIdx ディーラー (親) のインデックスを取得する
	GetDealerIdx() int
	// GetNonDealerIdx 非親 (elder hand) のインデックスを取得する
	GetNonDealerIdx() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetCurrentTrick 現在のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetTrumpSuit 切り札スートを取得する (AllFoursTrumpUnset=未確定)
	GetTrumpSuit() int
	// GetTurnUp めくり札を取得する
	GetTurnUp() *domain.Card
	// GetRunCount このディールの run 回数を取得する
	GetRunCount() int
	// GetWinnerIdx 勝者インデックスを取得する
	GetWinnerIdx() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.AllFoursPlayer
	// GetHint ヒントを取得する
	GetHint() *domain.AllFoursHint
	// GetValidPlayIndices プレイ可能なカードのインデックスを返す
	GetValidPlayIndices(playerIdx int) []int
}
