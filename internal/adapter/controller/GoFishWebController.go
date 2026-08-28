package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// GoFishWebInput Go FishWebインプット
type GoFishWebInput struct {
	BaseWebInput
	TargetIdx *int                  `json:"targetIdx,omitempty"`
	Rank      *int                  `json:"rank,omitempty"`
	Config    *GoFishWebInputConfig `json:"config,omitempty"`
}

// GoFishWebInputConfig Go Fish設定インプット
type GoFishWebInputConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
}

// GoFishWebOutputPlayer Go FishWebアウトプットプレイヤー
type GoFishWebOutputPlayer struct {
	ID        int                    `json:"id"`
	IsHuman   bool                   `json:"isHuman"`
	CardCount int                    `json:"cardCount"`
	Cards     []*WebOutputCard       `json:"cards"`
	BookCount int                    `json:"bookCount"`
	Books     []*GoFishWebOutputBook `json:"books"`
	// KnownRanks はこの席が保有を公開済みのランク。**サーバが持っている権威値**
	// で、CUI は以前からこれを出している。クライアントが lastAsk の履歴から
	// 組み直すと、リロードで記憶が消える。
	KnownRanks []int `json:"knownRanks"`
}

// GoFishWebOutputBook Go FishWebアウトプットブック
type GoFishWebOutputBook struct {
	Rank  int              `json:"rank"`
	Cards []*WebOutputCard `json:"cards"`
}

// GoFishWebOutputCpuAction CPUの行動記録
type GoFishWebOutputCpuAction struct {
	AskPlayerIdx  int            `json:"askPlayerIdx"`
	AskTargetIdx  int            `json:"askTargetIdx"`
	AskRank       int            `json:"askRank"`
	Success       bool           `json:"success"`
	CardsReceived int            `json:"cardsReceived"`
	DrawnCard     *WebOutputCard `json:"drawnCard,omitempty"`
	BookFormed    bool           `json:"bookFormed"`
	BookRank      int            `json:"bookRank,omitempty"`
}

// GoFishWebOutputLastAsk 最後の要求情報
type GoFishWebOutputLastAsk struct {
	PlayerIdx     int              `json:"playerIdx"`
	TargetIdx     int              `json:"targetIdx"`
	Rank          int              `json:"rank"`
	Success       bool             `json:"success"`
	CardsReceived []*WebOutputCard `json:"cardsReceived,omitempty"`
	DrawnCard     *WebOutputCard   `json:"drawnCard,omitempty"`
	BookFormed    bool             `json:"bookFormed"`
	BookRank      int              `json:"bookRank,omitempty"`
}

// GoFishWebOutput Go FishWebアウトプット
type GoFishWebOutput struct {
	Players       []*GoFishWebOutputPlayer    `json:"players"`
	Phase         int                         `json:"phase"`
	CurrentTurn   int                         `json:"currentTurn"`
	GameEndFlag   bool                        `json:"gameEndFlag"`
	WinnerIdx     int                         `json:"winnerIdx"`
	TurnNumber    int                         `json:"turnNumber"`
	DeckRemaining int                         `json:"deckRemaining"`
	LastAsk       *GoFishWebOutputLastAsk     `json:"lastAsk,omitempty"`
	CpuActions    []*GoFishWebOutputCpuAction `json:"cpuActions,omitempty"`
	HumanAction   *GoFishWebOutputCpuAction   `json:"humanAction,omitempty"`
	WebOutputBase
	Config GoFishWebOutputConfig `json:"config"`
}

// GoFishWebOutputConfig Go Fish設定アウトプット
type GoFishWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
}

// GoFishWebController Go FishWebコントローラークラス
type GoFishWebController = GameWebController[usecase.GoFishInteractorIF, GoFishWebInput, *GoFishWebOutput]

// NewGoFishWebController and NewGoFishWebControllerWithProvider are
// the standard and provider-backed constructors for GoFishWebController.
var NewGoFishWebController, NewGoFishWebControllerWithProvider = webControllerPair[usecase.GoFishInteractorIF, GoFishWebInput, *GoFishWebOutput](
	newGoFishDefaultOutput, goFishDispatch,
)

func newGoFishDefaultOutput(msg string) *GoFishWebOutput {
	return &GoFishWebOutput{
		Players:       make([]*GoFishWebOutputPlayer, 0),
		CpuActions:    make([]*GoFishWebOutputCpuAction, 0),
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func goFishDispatch(bc *baseController, w http.ResponseWriter, gi usecase.GoFishInteractorIF, param GoFishWebInput, newDefault func(string) *GoFishWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		cfg := buildGoFishConfig(gi.GetConfig(), param.Config)
		bc.writePresenterResponse(w, gi.Reset(cfg))
	case "ask":
		if !requireParam(bc, w, newDefault, param.TargetIdx == nil || param.Rank == nil, "param error: targetIdx and rank are required.") {
			return true
		}
		bc.writePresenterResponse(w, gi.Ask(*param.TargetIdx, *param.Rank))
	default:
		return dispatchLog(param.Command, bc, w, gi.ActionLog)
	}
	return true
}

func buildGoFishConfig(current domain.GoFishConfig, input *GoFishWebInputConfig) domain.GoFishConfig {
	if input == nil {
		return current
	}
	cfg := current
	if input.CpuDifficulty != nil {
		cfg.CpuDifficulty = domain.GoFishCpuDifficulty(*input.CpuDifficulty)
	}
	return cfg
}
