//go:build !js || !wasm || extra4

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// PiquetGame Piquetゲームインタフェース (usecase → domain ファサード)
type PiquetGame interface {
	BaseGame
	// Reset ゲーム初期化
	Reset()
	// NextDeal 次のディールに進む
	NextDeal()
	// ExchangeElder Elderの交換 (1..5枚)
	ExchangeElder(discardIndices []int) error
	// ExchangeYounger Youngerの交換 (0..3枚)
	ExchangeYounger(discardIndices []int) error
	// ResolveDeclaration 現宣言ステージを比較する
	ResolveDeclaration() (*domain.PiquetDeclarationResult, error)
	// PlayCard カードをプレイする
	PlayCard(cardIdx int) error
	// CpuPlay CPUの自動行動 (現フェーズに応じる)
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.PiquetConfig
	// SetConfig ゲーム設定を上書きする
	SetConfig(cfg domain.PiquetConfig)

	// GetGameEndFlag パルティ終了フラグ
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズ
	GetPhase() domain.PiquetPhase
	// GetDealNumber 現在のディール番号 (1始まり)
	GetDealNumber() int
	// GetDealsPerPartie パルティのディール数
	GetDealsPerPartie() int
	// GetElderIdx 現ディールのElderインデックス
	GetElderIdx() int
	// GetYoungerIdx 現ディールのYoungerインデックス
	GetYoungerIdx() int
	// IsHumanTurn 現在の手番が人間か
	IsHumanTurn() bool

	// GetPlayer プレイヤー
	GetPlayer(idx int) *domain.PiquetPlayer
	// GetPlayers 全プレイヤー
	GetPlayers() []*domain.PiquetPlayer

	// GetElderTalon Elderが交換可能な5枚
	GetElderTalon() []*domain.Card
	// GetYoungerTalon Youngerが交換可能な3枚
	GetYoungerTalon() []*domain.Card
	// GetExchangeTurn 現交換手番
	GetExchangeTurn() domain.PiquetExchangeTurn
	// GetElderExchangedCnt Elderが交換した枚数
	GetElderExchangedCnt() int
	// GetYoungerExchangedCnt Youngerが交換した枚数
	GetYoungerExchangedCnt() int
	// GetElderRevealedTalon Elderが見られるタロン残り
	GetElderRevealedTalon() []*domain.Card
	// GetYoungerRevealedTalon Youngerが見られるタロン残り
	GetYoungerRevealedTalon() []*domain.Card

	// GetCarteBlanche カルトブランシュ達成
	GetCarteBlanche(idx int) bool
	// GetDeclStage 現在の宣言ステージ
	GetDeclStage() domain.PiquetDeclarationKind
	// GetDeclResults 現ディールの宣言結果
	GetDeclResults() []*domain.PiquetDeclarationResult

	// GetCurrentTrick 現在のトリック
	GetCurrentTrick() []*domain.TrickCard
	// GetCurrentPlayerIdx 現プレイヤー
	GetCurrentPlayerIdx() int
	// GetTrickNumber 0始まりトリック番号
	GetTrickNumber() int
	// GetLeadPlayerIdx 現リード
	GetLeadPlayerIdx() int
	// GetTricksWon プレイヤーごとの獲得トリック数
	GetTricksWon(idx int) int
	// GetWinnerIdx パルティ勝者
	GetWinnerIdx() int

	// GetLegalPlayIndices 合法プレイインデックス
	GetLegalPlayIndices(playerIdx int) []int

	// GetHint ヒント
	GetHint(playerIdx int) *domain.PiquetHint
}
