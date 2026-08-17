package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PinochleWebInput ピノクルWebインプット
type PinochleWebInput struct {
	BaseWebInput
	BidAmount *int               `json:"bidAmount,omitempty"`
	Suit      *int               `json:"suit,omitempty"`
	CardIndex *int               `json:"cardIndex,omitempty"`
	Config    *PinochleWebConfig `json:"config,omitempty"`
}

// PinochleWebConfig ピノクルWeb設定
type PinochleWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	PointLimit    *int `json:"pointLimit,omitempty"`
}

// PinochleWebOutputPlayer ピノクルWebアウトプットプレイヤー
type PinochleWebOutputPlayer struct {
	ID          int              `json:"id"`
	IsHuman     bool             `json:"isHuman"`
	CardCount   int              `json:"cardCount"`
	Cards       []*WebOutputCard `json:"cards"`
	Team        int              `json:"team"`
	TrickCount  int              `json:"trickCount"`
	Bid         int              `json:"bid"`
	HasPassed   bool             `json:"hasPassed"`
	MeldScore   int              `json:"meldScore"`
	TrickPoints int              `json:"trickPoints"`
}

// PinochleWebOutputMeld メルド情報
type PinochleWebOutputMeld struct {
	Type   int              `json:"type"`
	Points int              `json:"points"`
	Cards  []*WebOutputCard `json:"cards"`
}

// PinochleWebOutputHint ヒント出力
type PinochleWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	BidAmount *int   `json:"bidAmount,omitempty"`
	Pass      *bool  `json:"pass,omitempty"`
	Suit      *int   `json:"suit,omitempty"`
	Reason    string `json:"reason"`
}

// PinochleWebOutput ピノクルWebアウトプット
type PinochleWebOutput struct {
	Players          []*PinochleWebOutputPlayer  `json:"players"`
	Phase            int                         `json:"phase"`
	RoundNumber      int                         `json:"roundNumber"`
	TrickNumber      int                         `json:"trickNumber"`
	CurrentPlayerIdx int                         `json:"currentPlayerIdx"`
	BidPlayerIdx     int                         `json:"bidPlayerIdx"`
	DealerIdx        int                         `json:"dealerIdx"`
	TrumpSuit        int                         `json:"trumpSuit"`
	HighestBid       int                         `json:"highestBid"`
	HighestBidder    int                         `json:"highestBidder"`
	CurrentTrick     []*WebOutputTrickCard       `json:"currentTrick"`
	TeamScores       [2]int                      `json:"teamScores"`
	GameEndFlag      bool                        `json:"gameEndFlag"`
	WinnerTeam       int                         `json:"winnerTeam"`
	LeadPlayerIdx    int                         `json:"leadPlayerIdx"`
	PlayerMelds      [4][]*PinochleWebOutputMeld `json:"playerMelds"`
	ValidPlayIndices []int                       `json:"validPlayIndices,omitempty"`
	Hint             *PinochleWebOutputHint      `json:"hint,omitempty"`
	// MeldTable はメルド15種類の点数一覧 (安い順)。ビッド額を決めるときの
	// 早見表として使う。**サーバが domain の値を送る。**フロントに書き写すと、
	// 加点を直したときに表だけが古いまま残る (#5519)。
	MeldTable []*PinochleWebOutputMeldTableEntry `json:"meldTable"`
	WebOutputBase
	Config PinochleWebOutputConfig `json:"config"`
}

// PinochleWebOutputMeldTableEntry メルド早見表の1行
type PinochleWebOutputMeldTableEntry struct {
	Type   int `json:"type"`
	Points int `json:"points"`
}

// PinochleWebOutputConfig ピノクル設定アウトプット
type PinochleWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	PointLimit    int `json:"pointLimit"`
}

// ToConfig builds a PinochleConfig from the nested web config, applying bounds checking.
func (c *PinochleWebConfig) ToConfig() domain.PinochleConfig {
	cfg := domain.DefaultPinochleConfig()
	cfg.CpuDifficulty = domain.PinochleCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.PinochleCpuDifficultyEasy), int(domain.PinochleCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.PointLimit, c.PointLimit, 1, 10000)
	return cfg
}

// ToConfig builds a PinochleConfig from the web input.
func (p PinochleWebInput) ToConfig() domain.PinochleConfig {
	return configOrDefault(p.Config, (*PinochleWebConfig).ToConfig, domain.DefaultPinochleConfig())
}

// PinochleWebController ピノクルWebコントローラークラス
type PinochleWebController = GameWebController[usecase.PinochleInteractorIF, PinochleWebInput, *PinochleWebOutput]

// NewPinochleWebController and NewPinochleWebControllerWithProvider are
// the standard and provider-backed constructors for PinochleWebController.
var NewPinochleWebController, NewPinochleWebControllerWithProvider = webControllerPair[usecase.PinochleInteractorIF, PinochleWebInput, *PinochleWebOutput](
	newPinochleDefaultOutput, pinochleDispatch,
)

func newPinochleDefaultOutput(msg string) *PinochleWebOutput {
	return &PinochleWebOutput{
		Players:       make([]*PinochleWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		WinnerTeam:    -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func pinochleDispatch(bc *baseController, w http.ResponseWriter, pi usecase.PinochleInteractorIF, param PinochleWebInput, newDefault func(string) *PinochleWebOutput) bool {
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
