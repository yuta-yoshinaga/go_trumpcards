package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NinetyNineWebInput ナインティナインWebインプット
type NinetyNineWebInput struct {
	BaseWebInput
	BuryIndices []int                `json:"buryIndices,omitempty"`
	CardIndex   *int                 `json:"cardIndex,omitempty"`
	Config      *NinetyNineWebConfig `json:"config,omitempty"`
}

// NinetyNineWebConfig ナインティナインWeb設定
type NinetyNineWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	TargetScore   *int `json:"targetScore,omitempty"`
}

// NinetyNineWebOutputPlayer ナインティナインWebアウトプットプレイヤー
type NinetyNineWebOutputPlayer struct {
	ID              int              `json:"id"`
	IsHuman         bool             `json:"isHuman"`
	CardCount       int              `json:"cardCount"`
	Cards           []*WebOutputCard `json:"cards"`
	Bid             int              `json:"bid"`
	RoundScore      int              `json:"roundScore"`
	CumulativeScore int              `json:"cumulativeScore"`
	TrickCount      int              `json:"trickCount"`
	BuriedCount     int              `json:"buriedCount"`
}

// NinetyNineWebOutputHint ヒント出力
type NinetyNineWebOutputHint struct {
	CardIndex   *int   `json:"cardIndex,omitempty"`
	BuryIndices []int  `json:"buryIndices,omitempty"`
	Reason      string `json:"reason"`
}

// NinetyNineWebOutput ナインティナインWebアウトプット
type NinetyNineWebOutput struct {
	Players          []*NinetyNineWebOutputPlayer `json:"players"`
	Phase            int                          `json:"phase"`
	DealNumber       int                          `json:"dealNumber"`
	TargetScore      int                          `json:"targetScore"`
	HandSize         int                          `json:"handSize"`
	TrickNumber      int                          `json:"trickNumber"`
	CurrentPlayerIdx int                          `json:"currentPlayerIdx"`
	BidPlayerIdx     int                          `json:"bidPlayerIdx"`
	DealerIdx        int                          `json:"dealerIdx"`
	TrumpSuit        int                          `json:"trumpSuit"`
	CurrentTrick     []*WebOutputTrickCard        `json:"currentTrick"`
	GameEndFlag      bool                         `json:"gameEndFlag"`
	WinnerIdx        int                          `json:"winnerIdx"`
	LeadPlayerIdx    int                          `json:"leadPlayerIdx"`
	Hint             *NinetyNineWebOutputHint     `json:"hint,omitempty"`
	WebOutputBase
	Config NinetyNineWebOutputConfig `json:"config"`
}

// NinetyNineWebOutputConfig ナインティナイン設定アウトプット
type NinetyNineWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	TargetScore   int `json:"targetScore"`
}

// ToConfig builds a NinetyNineConfig from the nested web config, applying bounds checking.
func (c *NinetyNineWebConfig) ToConfig() domain.NinetyNineConfig {
	cfg := domain.DefaultNinetyNineConfig()
	cfg.CpuDifficulty = domain.NinetyNineCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.NinetyNineCpuDifficultyEasy), int(domain.NinetyNineCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.TargetScore, c.TargetScore, 10, 1000)
	return cfg
}

// ToConfig builds a NinetyNineConfig from the web input.
func (p NinetyNineWebInput) ToConfig() domain.NinetyNineConfig {
	return configOrDefault(p.Config, (*NinetyNineWebConfig).ToConfig, domain.DefaultNinetyNineConfig())
}

// NinetyNineWebController ナインティナインWebコントローラークラス
type NinetyNineWebController = GameWebController[usecase.NinetyNineInteractorIF, NinetyNineWebInput, *NinetyNineWebOutput]

// NewNinetyNineWebController and NewNinetyNineWebControllerWithProvider are
// the standard and provider-backed constructors for NinetyNineWebController.
var NewNinetyNineWebController, NewNinetyNineWebControllerWithProvider = webControllerPair[usecase.NinetyNineInteractorIF, NinetyNineWebInput, *NinetyNineWebOutput](
	newNinetyNineDefaultOutput, ninetyNineDispatch,
)

func newNinetyNineDefaultOutput(msg string) *NinetyNineWebOutput {
	return &NinetyNineWebOutput{
		Players:       make([]*NinetyNineWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func ninetyNineDispatch(bc *baseController, w http.ResponseWriter, oi usecase.NinetyNineInteractorIF, param NinetyNineWebInput, newDefault func(string) *NinetyNineWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, oi.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		if !requireParam(bc, w, newDefault, param.BuryIndices == nil, "param error: buryIndices is required.") {
			return true
		}
		bc.writePresenterResponse(w, oi.Bid(param.BuryIndices))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, oi.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, oi.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, oi.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, oi.Hint, oi.ActionLog)
	}
	return true
}
