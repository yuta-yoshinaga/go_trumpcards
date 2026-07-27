//go:build !js || !wasm || extra

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CinchWebConfig はチンチ (Cinch) の Web 設定。
type CinchWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	PointLimit    *int `json:"pointLimit,omitempty"`
}

// ToConfig は CinchWebConfig を domain.CinchConfig に変換する (境界チェック付き)。
func (c *CinchWebConfig) ToConfig() domain.CinchConfig {
	cfg := domain.DefaultCinchConfig()
	cfg.CpuDifficulty = domain.CinchCpuDifficulty(webutil.BoundedIntPtr(
		c.CpuDifficulty, int(domain.CinchDifficultyEasy), int(domain.CinchDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.PointLimit, c.PointLimit, 1, 1000000)
	return cfg
}

// CinchWebInput はチンチ Web インプット。
type CinchWebInput struct {
	BaseWebInput
	CardIndex *int            `json:"cardIndex,omitempty"`
	Bid       *int            `json:"bid,omitempty"`
	TrumpSuit *int            `json:"trumpSuit,omitempty"`
	Config    *CinchWebConfig `json:"config,omitempty"`
}

// ToConfig は Web インプットから domain.CinchConfig を構築する。
func (p CinchWebInput) ToConfig() domain.CinchConfig {
	return configOrDefault(p.Config, (*CinchWebConfig).ToConfig, domain.DefaultCinchConfig())
}

// CinchWebOutputPlayer はチンチ Web アウトプットプレイヤー。
type CinchWebOutputPlayer struct {
	ID         int              `json:"id"`
	IsHuman    bool             `json:"isHuman"`
	CardCount  int              `json:"cardCount"`
	Cards      []*WebOutputCard `json:"cards"`
	TrickCount int              `json:"trickCount"`
	Bid        int              `json:"bid"`
	TotalScore int              `json:"totalScore"`
}

// CinchWebOutputDealDetail は 1 ディールの得点内訳。
type CinchWebOutputDealDetail struct {
	TrumpSuit int         `json:"trumpSuit"`
	BidderIdx int         `json:"bidderIdx"`
	Bid       int         `json:"bid"`
	SetBack   bool        `json:"setBack"`
	Points    map[int]int `json:"points"`
	Gained    map[int]int `json:"gained"`
}

// CinchWebOutputHint はヒント出力。
type CinchWebOutputHint struct {
	CardIndices []int  `json:"cardIndices"`
	Bid         *int   `json:"bid,omitempty"`
	TrumpSuit   *int   `json:"trumpSuit,omitempty"`
	Reason      string `json:"reason"`
}

// CinchWebOutput はチンチ Web アウトプット。
type CinchWebOutput struct {
	Players         []*CinchWebOutputPlayer   `json:"players"`
	Phase           int                       `json:"phase"`
	RoundNumber     int                       `json:"roundNumber"`
	TrickNumber     int                       `json:"trickNumber"`
	TotalTricks     int                       `json:"totalTricks"`
	DealerIdx       int                       `json:"dealerIdx"`
	CurrentTurn     int                       `json:"currentTurn"`
	BidPlayerIdx    int                       `json:"bidPlayerIdx"`
	CurrentBid      int                       `json:"currentBid"`
	BidWinnerIdx    int                       `json:"bidWinnerIdx"`
	TrumpSuit       int                       `json:"trumpSuit"`
	CurrentTrick    []*WebOutputTrickCard     `json:"currentTrick"`
	LastTrick       []*WebOutputTrickCard     `json:"lastTrick"`
	LastTrickWinner int                       `json:"lastTrickWinner"`
	PlayableIndices []int                     `json:"playableIndices"`
	GameEndFlag     bool                      `json:"gameEndFlag"`
	WinnerIdx       int                       `json:"winnerIdx"`
	RoundWinners    []int                     `json:"roundWinners"`
	LastDealDetail  *CinchWebOutputDealDetail `json:"lastDealDetail"`
	IsHumanTurn     bool                      `json:"isHumanTurn"`
	Hint            *CinchWebOutputHint       `json:"hint,omitempty"`
	WebOutputBase
	Config CinchWebConfigOutput `json:"config"`
}

// CinchWebConfigOutput は設定アウトプット。
type CinchWebConfigOutput struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	PointLimit    int `json:"pointLimit"`
}

// CinchWebController はチンチ Web コントローラークラス。
type CinchWebController = GameWebController[usecase.CinchInteractorIF, CinchWebInput, *CinchWebOutput]

// NewCinchWebController, NewCinchWebControllerWithProvider are the standard and
// provider-backed constructors for CinchWebController.
var NewCinchWebController, NewCinchWebControllerWithProvider = webControllerPair[usecase.CinchInteractorIF, CinchWebInput, *CinchWebOutput](
	newCinchDefaultOutput, cinchDispatch,
)

func newCinchDefaultOutput(msg string) *CinchWebOutput {
	return &CinchWebOutput{
		Players:         make([]*CinchWebOutputPlayer, 0),
		CurrentTrick:    make([]*WebOutputTrickCard, 0),
		LastTrick:       make([]*WebOutputTrickCard, 0),
		PlayableIndices: make([]int, 0),
		RoundWinners:    make([]int, 0),
		TotalTricks:     domain.CinchTotalTricks,
		LastTrickWinner: -1,
		BidWinnerIdx:    -1,
		WinnerIdx:       -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func cinchDispatch(bc *baseController, w http.ResponseWriter, ci usecase.CinchInteractorIF, param CinchWebInput, newDefault func(string) *CinchWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ci.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		if !requireParam(bc, w, newDefault, param.Bid == nil, "param error: bid is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Bid(*param.Bid))
	case "t", "trump":
		if !requireParam(bc, w, newDefault, param.TrumpSuit == nil, "param error: trumpSuit is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.NameTrump(*param.TrumpSuit))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ci.Play(*param.CardIndex))
	case "nr", "nextround", "n", "next":
		bc.writePresenterResponse(w, ci.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ci.Hint, ci.ActionLog)
	}
	return true
}
