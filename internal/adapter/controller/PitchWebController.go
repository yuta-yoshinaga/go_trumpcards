package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PitchWebInput ピッチWebインプット
type PitchWebInput struct {
	BaseWebInput
	Bid       *int            `json:"bid,omitempty"`
	CardIndex *int            `json:"cardIndex,omitempty"`
	Config    *PitchWebConfig `json:"config,omitempty"`
}

// PitchWebConfig ピッチWeb設定
type PitchWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	PointLimit    *int `json:"pointLimit,omitempty"`
}

// PitchWebOutputPlayer ピッチWebアウトプットプレイヤー
type PitchWebOutputPlayer struct {
	ID              int              `json:"id"`
	IsHuman         bool             `json:"isHuman"`
	CardCount       int              `json:"cardCount"`
	Cards           []*WebOutputCard `json:"cards"`
	Bid             int              `json:"bid"`
	RoundScore      int              `json:"roundScore"`
	CumulativeScore int              `json:"cumulativeScore"`
	TrickCount      int              `json:"trickCount"`
}

// PitchWebOutputTrickCard トリック中の1枚
type PitchWebOutputTrickCard struct {
	PlayerIdx int            `json:"playerIdx"`
	Card      *WebOutputCard `json:"card"`
}

// PitchWebOutputHint ヒント出力
type PitchWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Bid       *int   `json:"bid,omitempty"`
	Reason    string `json:"reason"`
}

// PitchWebOutput ピッチWebアウトプット
type PitchWebOutput struct {
	Players          []*PitchWebOutputPlayer    `json:"players"`
	Phase            int                        `json:"phase"`
	RoundNumber      int                        `json:"roundNumber"`
	TrickNumber      int                        `json:"trickNumber"`
	DealerIdx        int                        `json:"dealerIdx"`
	CurrentPlayerIdx int                        `json:"currentPlayerIdx"`
	BidPlayerIdx     int                        `json:"bidPlayerIdx"`
	CurrentBid       int                        `json:"currentBid"`
	BidWinnerIdx     int                        `json:"bidWinnerIdx"`
	TrumpSuit        int                        `json:"trumpSuit"`
	CurrentTrick     []*PitchWebOutputTrickCard `json:"currentTrick"`
	LastTrick        []*PitchWebOutputTrickCard `json:"lastTrick"`
	LastTrickWinner  int                        `json:"lastTrickWinner"`
	GameEndFlag      bool                       `json:"gameEndFlag"`
	WinnerIdx        int                        `json:"winnerIdx"`
	LeadPlayerIdx    int                        `json:"leadPlayerIdx"`
	ValidPlayIndices []int                      `json:"validPlayIndices,omitempty"`
	Hint             *PitchWebOutputHint        `json:"hint,omitempty"`
	WebOutputBase
	Config PitchWebOutputConfig `json:"config"`
}

// PitchWebOutputConfig ピッチ設定アウトプット
type PitchWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	PointLimit    int `json:"pointLimit"`
}

// ToConfig builds a PitchConfig from the nested web config, applying bounds checking.
func (c *PitchWebConfig) ToConfig() domain.PitchConfig {
	cfg := domain.DefaultPitchConfig()
	cfg.CpuDifficulty = domain.PitchCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.PitchCpuDifficultyEasy), int(domain.PitchCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.PointLimit, c.PointLimit, 1, 1000)
	return cfg
}

// ToConfig builds a PitchConfig from the web input.
func (p PitchWebInput) ToConfig() domain.PitchConfig {
	return configOrDefault(p.Config, (*PitchWebConfig).ToConfig, domain.DefaultPitchConfig())
}

// PitchWebController ピッチWebコントローラークラス
type PitchWebController = GameWebController[usecase.PitchInteractorIF, PitchWebInput, *PitchWebOutput]

// NewPitchWebController and NewPitchWebControllerWithProvider are
// the standard and provider-backed constructors for PitchWebController.
var NewPitchWebController, NewPitchWebControllerWithProvider = webControllerPair[usecase.PitchInteractorIF, PitchWebInput, *PitchWebOutput](
	newPitchDefaultOutput, pitchDispatch,
)

func newPitchDefaultOutput(msg string) *PitchWebOutput {
	return &PitchWebOutput{
		Players:         make([]*PitchWebOutputPlayer, 0),
		CurrentTrick:    make([]*PitchWebOutputTrickCard, 0),
		LastTrick:       make([]*PitchWebOutputTrickCard, 0),
		LastTrickWinner: -1,
		WinnerIdx:       -1,
		BidWinnerIdx:    -1,
		WebOutputBase:   WebOutputBase{Message: msg},
	}
}

func pitchDispatch(bc *baseController, w http.ResponseWriter, pi usecase.PitchInteractorIF, param PitchWebInput, newDefault func(string) *PitchWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, pi.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		if !requireParam(bc, w, newDefault, param.Bid == nil, "param error: bid is required.") {
			return true
		}
		bc.writePresenterResponse(w, pi.Bid(*param.Bid))
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
