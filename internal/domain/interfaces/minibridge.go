//go:build !js || !wasm || extra3

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// MinibridgeGame ミニブリッジゲームインタフェース
type MinibridgeGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlayerSelectContract 人間（落札者）が契約を選ぶ
	PlayerSelectContract(level, suit int) error
	// CpuSelectContract CPUの落札者が契約を選ぶ
	CpuSelectContract()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPUプレイヤーが1枚出す
	CpuPlay()
	// NextRound 次のディールを開始する
	NextRound()
	// GiveUp 投了する
	GiveUp()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.MinibridgeConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.MinibridgeConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.MinibridgePhase
	// IsHumanTurn 現在の手番が人間かを返す（ダミーの手番を含む）
	IsHumanTurn() bool
	// IsHumanContractTurn 人間が契約を選ぶ番かを返す
	IsHumanContractTurn() bool
	// GetRoundNumber 現在のディール番号を取得する
	GetRoundNumber() int
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetContractLevel 契約レベルを取得する (0: 未決定)
	GetContractLevel() int
	// GetContractSuit 契約の種別を取得する (0: ノートランプ)
	GetContractSuit() int
	// RequiredTricks 契約に必要なトリック数を取得する
	RequiredTricks() int
	// GetDeclarerIdx 落札者を取得する (-1: 未定)
	GetDeclarerIdx() int
	// GetDummyIdx ダミーを取得する (-1: 未定)
	GetDummyIdx() int
	// GetDummyHand ダミーの手札を取得する（契約決定後に公開）
	GetDummyHand() []*domain.Card
	// GetLastMade 直前のディールで契約が成立したかを取得する
	GetLastMade() bool
	// GetLastTricks 直前のディールで宣言側が取ったトリック数を取得する
	GetLastTricks() int
	// GetTeamScore チームの累計得点を取得する
	GetTeamScore(team int) int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetCurrentTrick 現在のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す
	GetValidPlayIndices(playerIdx int) []int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.MinibridgePlayer
	// GetWinnerTeam 勝ったチームを取得する (-1: 未確定/同点)
	GetWinnerTeam() int
	// GetHint ヒントを取得する
	GetHint() *domain.MinibridgeHint
}
