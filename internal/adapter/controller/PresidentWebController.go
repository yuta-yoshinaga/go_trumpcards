package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PresidentWebConfig ローカルルール設定 (入力・出力共用)
type PresidentWebConfig struct {
	RevolutionEnabled     bool `json:"revolutionEnabled"`
	CardExchangeEnabled   bool `json:"cardExchangeEnabled"`
	PassFieldFlushEnabled bool `json:"passFieldFlushEnabled"`
	CpuDifficulty         int  `json:"cpuDifficulty"`
}

// ToConfig converts PresidentWebConfig to domain.PresidentConfig.
func (c PresidentWebConfig) ToConfig() domain.PresidentConfig {
	return domain.PresidentConfig{
		RevolutionEnabled:     c.RevolutionEnabled,
		CardExchangeEnabled:   c.CardExchangeEnabled,
		PassFieldFlushEnabled: c.PassFieldFlushEnabled,
		CpuDifficulty:         domain.PresidentCpuDifficulty(c.CpuDifficulty),
	}
}

// PresidentWebInput プレジデントWebインプット
type PresidentWebInput struct {
	BaseWebInput
	Indices []int               `json:"indices"` // 出すカードのインデックス。play コマンド用。空の場合はパス。
	Config  *PresidentWebConfig `json:"config"`  // リセット時のローカルルール設定 (省略可)
}

// PresidentWebOutputPlayer プレジデントWebアウトプットプレイヤー
type PresidentWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	IsFinished bool             `json:"isFinished"`
	Rank       int              `json:"rank"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
}

// PresidentWebOutputAction プレイヤー行動記録
type PresidentWebOutputAction struct {
	PlayerIdx   int              `json:"playerIdx"`
	PlayedCards []*WebOutputCard `json:"playedCards"` // nil = パス
}

// PresidentWebOutputExchangeAction カード交換記録
type PresidentWebOutputExchangeAction struct {
	FromPlayerIdx int              `json:"fromPlayerIdx"`
	ToPlayerIdx   int              `json:"toPlayerIdx"`
	Cards         []*WebOutputCard `json:"cards"`
}

// PresidentWebOutput プレジデントWebアウトプット
type PresidentWebOutput struct {
	Players           []*PresidentWebOutputPlayer         `json:"players"`
	CurrentTurn       int                                 `json:"currentTurn"`
	TableCards        []*WebOutputCard                    `json:"tableCards"`
	LastPlayPlayerIdx int                                 `json:"lastPlayPlayerIdx"`
	GameEndFlag       bool                                `json:"gameEndFlag"`
	RevolutionActive  bool                                `json:"revolutionActive"`
	Config            PresidentWebConfig                  `json:"config"`
	ExchangeActions   []*PresidentWebOutputExchangeAction `json:"exchangeActions"`
	CpuActions        []*PresidentWebOutputAction         `json:"cpuActions"`
	HumanAction       *PresidentWebOutputAction           `json:"humanAction"`
	WebOutputBase
}

// PresidentWebController プレジデントWebコントローラークラス
type PresidentWebController = GameWebController[usecase.PresidentInteractorIF, PresidentWebInput, *PresidentWebOutput]

// NewPresidentWebController, NewPresidentWebControllerWithProvider are the
// standard and provider-backed constructors for PresidentWebController.
var NewPresidentWebController, NewPresidentWebControllerWithProvider = webControllerPair[usecase.PresidentInteractorIF, PresidentWebInput, *PresidentWebOutput](
	newPresidentDefaultOutput, presidentDispatch,
)

func newPresidentDefaultOutput(msg string) *PresidentWebOutput {
	return &PresidentWebOutput{
		Players:         make([]*PresidentWebOutputPlayer, 0),
		TableCards:      make([]*WebOutputCard, 0),
		CpuActions:      make([]*PresidentWebOutputAction, 0),
		ExchangeActions: make([]*PresidentWebOutputExchangeAction, 0),
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func presidentDispatch(bc *baseController, w http.ResponseWriter, pi usecase.PresidentInteractorIF, param PresidentWebInput, _ func(string) *PresidentWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		if param.Config != nil {
			bc.writePresenterResponse(w, pi.ResetWithConfig(param.Config.ToConfig()))
		} else {
			bc.writePresenterResponse(w, pi.Reset())
		}
	case "p", "play":
		indices := param.Indices
		if indices == nil {
			indices = []int{}
		}
		bc.writePresenterResponse(w, pi.Play(indices))
	default:
		return dispatchHintAndLog(param.Command, bc, w, pi.Hint, pi.ActionLog)
	}
	return true
}
