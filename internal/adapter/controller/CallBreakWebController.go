package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CallBreakWebInput Call Break Web インプット
type CallBreakWebInput struct {
	BaseWebInput
	Bid       *int                `json:"bid,omitempty"`
	CardIndex *int                `json:"cardIndex,omitempty"`
	Config    *CallBreakWebConfig `json:"config,omitempty"`
}

// CallBreakWebConfig Call Break Web 設定
type CallBreakWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	MaxRounds     *int `json:"maxRounds,omitempty"`
}

// CallBreakWebOutputPlayer Call Break Web アウトプットプレイヤー
//
// RoundScore / CumulativeScore は内部値 (×10) をそのまま返す。
// 表示側で /10 して "X.Y" 形式で描画する。
type CallBreakWebOutputPlayer struct {
	ID              int              `json:"id"`
	IsHuman         bool             `json:"isHuman"`
	CardCount       int              `json:"cardCount"`
	Cards           []*WebOutputCard `json:"cards"`
	Bid             int              `json:"bid"`
	RoundScore      int              `json:"roundScore"`
	CumulativeScore int              `json:"cumulativeScore"`
	TrickCount      int              `json:"trickCount"`
	// Bags はビッドを超えて取った余剰トリック数。式をページ側に置くと CUI と
	// 食い違うので、ドメインの GetBags() をそのまま載せる (#4752)。
	Bags int `json:"bags"`
}

// CallBreakWebOutputHint ヒント出力
type CallBreakWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Bid       *int   `json:"bid,omitempty"`
	Reason    string `json:"reason"`
}

// CallBreakWebOutput Call Break Web アウトプット
type CallBreakWebOutput struct {
	Players          []*CallBreakWebOutputPlayer `json:"players"`
	Phase            int                         `json:"phase"`
	RoundNumber      int                         `json:"roundNumber"`
	TrickNumber      int                         `json:"trickNumber"`
	CurrentPlayerIdx int                         `json:"currentPlayerIdx"`
	BidPlayerIdx     int                         `json:"bidPlayerIdx"`
	CurrentTrick     []*WebOutputTrickCard       `json:"currentTrick"`
	SpadesBroken     bool                        `json:"spadesBroken"`
	GameEndFlag      bool                        `json:"gameEndFlag"`
	WinnerIdx        int                         `json:"winnerIdx"`
	LeadPlayerIdx    int                         `json:"leadPlayerIdx"`
	Hint             *CallBreakWebOutputHint     `json:"hint,omitempty"`
	ValidPlayIndices []int                       `json:"validPlayIndices"`
	WebOutputBase
	Config CallBreakWebOutputConfig `json:"config"`
}

// CallBreakWebOutputConfig Call Break 設定アウトプット
type CallBreakWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	MaxRounds     int `json:"maxRounds"`
}

// ToConfig builds a CallBreakConfig from the nested web config, applying bounds checking.
func (c *CallBreakWebConfig) ToConfig() domain.CallBreakConfig {
	cfg := domain.DefaultCallBreakConfig()
	cfg.CpuDifficulty = domain.CallBreakCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.CallBreakCpuDifficultyEasy), int(domain.CallBreakCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.MaxRounds, c.MaxRounds, 1, 50)
	return cfg
}

// ToConfig builds a CallBreakConfig from the web input.
func (p CallBreakWebInput) ToConfig() domain.CallBreakConfig {
	return configOrDefault(p.Config, (*CallBreakWebConfig).ToConfig, domain.DefaultCallBreakConfig())
}

// CallBreakWebController Call Break Web コントローラークラス
type CallBreakWebController = GameWebController[usecase.CallBreakInteractorIF, CallBreakWebInput, *CallBreakWebOutput]

// NewCallBreakWebController and NewCallBreakWebControllerWithProvider are
// the standard and provider-backed constructors for CallBreakWebController.
var NewCallBreakWebController, NewCallBreakWebControllerWithProvider = webControllerPair[usecase.CallBreakInteractorIF, CallBreakWebInput, *CallBreakWebOutput](
	newCallBreakDefaultOutput, callBreakDispatch,
)

func newCallBreakDefaultOutput(msg string) *CallBreakWebOutput {
	return &CallBreakWebOutput{
		Players:          make([]*CallBreakWebOutputPlayer, 0),
		CurrentTrick:     make([]*WebOutputTrickCard, 0),
		ValidPlayIndices: make([]int, 0),
		WinnerIdx:        -1,
		WebOutputBase:    WebOutputBase{Message: msg},
	}
}

func callBreakDispatch(bc *baseController, w http.ResponseWriter, ci usecase.CallBreakInteractorIF, param CallBreakWebInput, newDefault func(string) *CallBreakWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ci.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		if !requireParam(bc, w, newDefault, param.Bid == nil, "param error: bid is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Bid(*param.Bid))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, ci.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, ci.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ci.Hint, ci.ActionLog)
	}
	return true
}
