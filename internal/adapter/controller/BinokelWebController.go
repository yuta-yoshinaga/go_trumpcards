package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BinokelWebInput ビノクルWebインプット
type BinokelWebInput struct {
	BaseWebInput
	BidAmount      *int              `json:"bidAmount,omitempty"`
	Suit           *int              `json:"suit,omitempty"`
	CardIndex      *int              `json:"cardIndex,omitempty"`
	DiscardIndices []int             `json:"discardIndices,omitempty"`
	CardIndices    []int             `json:"cardIndices,omitempty"`
	Config         *BinokelWebConfig `json:"config,omitempty"`
}

// BinokelWebConfig ビノクルWeb設定
type BinokelWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	PointLimit    *int `json:"pointLimit,omitempty"`
}

// BinokelWebOutputPlayer ビノクルWebアウトプットプレイヤー
type BinokelWebOutputPlayer struct {
	ID          int              `json:"id"`
	IsHuman     bool             `json:"isHuman"`
	CardCount   int              `json:"cardCount"`
	Cards       []*WebOutputCard `json:"cards"`
	Score       int              `json:"score"`
	TrickCount  int              `json:"trickCount"`
	Bid         int              `json:"bid"`
	HasPassed   bool             `json:"hasPassed"`
	MeldScore   int              `json:"meldScore"`
	TrickPoints int              `json:"trickPoints"`
}

// BinokelWebOutputMeld メルド情報
type BinokelWebOutputMeld struct {
	Type   int              `json:"type"`
	Points int              `json:"points"`
	Cards  []*WebOutputCard `json:"cards"`
}

// BinokelWebOutputHint ヒント出力
type BinokelWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	BidAmount *int   `json:"bidAmount,omitempty"`
	Pass      *bool  `json:"pass,omitempty"`
	Suit      *int   `json:"suit,omitempty"`
	Reason    string `json:"reason"`
}

// BinokelWebOutput ビノクルWebアウトプット
type BinokelWebOutput struct {
	Players          []*BinokelWebOutputPlayer                        `json:"players"`
	Phase            int                                              `json:"phase"`
	RoundNumber      int                                              `json:"roundNumber"`
	TrickNumber      int                                              `json:"trickNumber"`
	CurrentPlayerIdx int                                              `json:"currentPlayerIdx"`
	BidPlayerIdx     int                                              `json:"bidPlayerIdx"`
	DealerIdx        int                                              `json:"dealerIdx"`
	TrumpSuit        int                                              `json:"trumpSuit"`
	HighestBid       int                                              `json:"highestBid"`
	HighestBidder    int                                              `json:"highestBidder"`
	CurrentTrick     []*WebOutputTrickCard                            `json:"currentTrick"`
	Scores           [domain.BinokelPlayerCnt]int                     `json:"scores"`
	GameEndFlag      bool                                             `json:"gameEndFlag"`
	WinnerPlayer     int                                              `json:"winnerPlayer"`
	LeadPlayerIdx    int                                              `json:"leadPlayerIdx"`
	PlayerMelds      [domain.BinokelPlayerCnt][]*BinokelWebOutputMeld `json:"playerMelds"`
	Dabb             []*WebOutputCard                                 `json:"dabb"`
	DabbDiscarded    []*WebOutputCard                                 `json:"dabbDiscarded"`
	ValidPlayIndices []int                                            `json:"validPlayIndices,omitempty"`
	Hint             *BinokelWebOutputHint                            `json:"hint,omitempty"`
	MeldTable        []*BinokelWebOutputMeldTableEntry                `json:"meldTable"`
	WebOutputBase
	Config BinokelWebOutputConfig `json:"config"`
}

// BinokelWebOutputMeldTableEntry メルド早見表の1行
type BinokelWebOutputMeldTableEntry struct {
	Type   int `json:"type"`
	Points int `json:"points"`
}

// BinokelWebOutputConfig ビノクル設定アウトプット
type BinokelWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	PointLimit    int `json:"pointLimit"`
}

// ToConfig builds a BinokelConfig from the nested web config, applying bounds checking.
func (c *BinokelWebConfig) ToConfig() domain.BinokelConfig {
	cfg := domain.DefaultBinokelConfig()
	cfg.CpuDifficulty = domain.BinokelCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.BinokelCpuDifficultyEasy), int(domain.BinokelCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.PointLimit, c.PointLimit, 1, 10000)
	return cfg
}

// ToConfig builds a BinokelConfig from the web input.
func (p BinokelWebInput) ToConfig() domain.BinokelConfig {
	return configOrDefault(p.Config, (*BinokelWebConfig).ToConfig, domain.DefaultBinokelConfig())
}

// BinokelWebController ビノクルWebコントローラークラス
type BinokelWebController = GameWebController[usecase.BinokelInteractorIF, BinokelWebInput, *BinokelWebOutput]

// NewBinokelWebController and NewBinokelWebControllerWithProvider are
// the standard and provider-backed constructors for BinokelWebController.
var NewBinokelWebController, NewBinokelWebControllerWithProvider = webControllerPair[usecase.BinokelInteractorIF, BinokelWebInput, *BinokelWebOutput](
	newBinokelDefaultOutput, binokelDispatch,
)

func newBinokelDefaultOutput(msg string) *BinokelWebOutput {
	return &BinokelWebOutput{
		Players:       make([]*BinokelWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		Dabb:          make([]*WebOutputCard, 0),
		DabbDiscarded: make([]*WebOutputCard, 0),
		WinnerPlayer:  -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func binokelDispatch(bc *baseController, w http.ResponseWriter, pi usecase.BinokelInteractorIF, param BinokelWebInput, newDefault func(string) *BinokelWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, pi.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		if !requireParam(bc, w, newDefault, param.BidAmount == nil, "param error: bidAmount is required.") {
			return true
		}
		bc.writePresenterResponse(w, pi.Bid(*param.BidAmount))
	case "pa", "pass":
		bc.writePresenterResponse(w, pi.Pass())
	case "d", "discard":
		indices := param.DiscardIndices
		if len(indices) == 0 {
			indices = param.CardIndices
		}
		if !requireParam(bc, w, newDefault, len(indices) != domain.BinokelDabbSize, "param error: discard requires 3 card indices.") {
			return true
		}
		bc.writePresenterResponse(w, pi.DiscardToDabb(indices))
	case "t", "trump":
		if !requireParam(bc, w, newDefault, param.Suit == nil, "param error: suit is required.") {
			return true
		}
		bc.writePresenterResponse(w, pi.CallTrump(*param.Suit))
	case "m", "meld":
		bc.writePresenterResponse(w, pi.ConfirmMelds())
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, pi.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, pi.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, pi.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, pi.Hint, pi.ActionLog)
	}
	return true
}
