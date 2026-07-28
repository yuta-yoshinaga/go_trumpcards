//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TarneebWebInput Tarneeb Web インプット
type TarneebWebInput struct {
	BaseWebInput
	Bid       *int              `json:"bid,omitempty"`
	TrumpSuit *int              `json:"trumpSuit,omitempty"`
	CardIndex *int              `json:"cardIndex,omitempty"`
	Config    *TarneebWebConfig `json:"config,omitempty"`
}

// TarneebWebConfig Tarneeb Web 設定
type TarneebWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	PointLimit    *int `json:"pointLimit,omitempty"`
	MinBid        *int `json:"minBid,omitempty"`
}

// TarneebWebOutputPlayer Tarneeb Web アウトプットプレイヤー
type TarneebWebOutputPlayer struct {
	ID              int              `json:"id"`
	IsHuman         bool             `json:"isHuman"`
	Team            int              `json:"team"`
	CardCount       int              `json:"cardCount"`
	Cards           []*WebOutputCard `json:"cards"`
	Bid             int              `json:"bid"`
	RoundScore      int              `json:"roundScore"`
	CumulativeScore int              `json:"cumulativeScore"`
	TrickCount      int              `json:"trickCount"`
}

// TarneebWebOutputHint ヒント出力
type TarneebWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Bid       *int   `json:"bid,omitempty"`
	TrumpSuit *int   `json:"trumpSuit,omitempty"`
	Reason    string `json:"reason"`
}

// TarneebWebOutputConfig Tarneeb 設定アウトプット
type TarneebWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	PointLimit    int `json:"pointLimit"`
	MinBid        int `json:"minBid"`
}

// TarneebWebOutput Tarneeb Web アウトプット
type TarneebWebOutput struct {
	Players          []*TarneebWebOutputPlayer `json:"players"`
	TeamScores       []int                     `json:"teamScores"`
	Phase            int                       `json:"phase"`
	RoundNumber      int                       `json:"roundNumber"`
	TrickNumber      int                       `json:"trickNumber"`
	CurrentPlayerIdx int                       `json:"currentPlayerIdx"`
	BidPlayerIdx     int                       `json:"bidPlayerIdx"`
	BidWinnerIdx     int                       `json:"bidWinnerIdx"`
	HighestBid       int                       `json:"highestBid"`
	TrumpSuit        int                       `json:"trumpSuit"`
	RedealCount      int                       `json:"redealCount"`
	DealerIdx        int                       `json:"dealerIdx"`
	CurrentTrick     []*WebOutputTrickCard     `json:"currentTrick"`
	GameEndFlag      bool                      `json:"gameEndFlag"`
	WinnerTeam       int                       `json:"winnerTeam"`
	LeadPlayerIdx    int                       `json:"leadPlayerIdx"`
	Hint             *TarneebWebOutputHint     `json:"hint,omitempty"`
	WebOutputBase
	Config TarneebWebOutputConfig `json:"config"`
}

// ToConfig builds a TarneebConfig from the nested web config, applying bounds checking.
func (c *TarneebWebConfig) ToConfig() domain.TarneebConfig {
	cfg := domain.DefaultTarneebConfig()
	cfg.CpuDifficulty = domain.TarneebCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.TarneebCpuDifficultyEasy), int(domain.TarneebCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.PointLimit, c.PointLimit, 1, domain.TarneebMaxPointLimit)
	webutil.ApplyBoundedInt(&cfg.MinBid, c.MinBid, 1, 13)
	return cfg
}

// ToConfig builds a TarneebConfig from the web input.
func (p TarneebWebInput) ToConfig() domain.TarneebConfig {
	return configOrDefault(p.Config, (*TarneebWebConfig).ToConfig, domain.DefaultTarneebConfig())
}

// TarneebWebController Tarneeb Web コントローラークラス
type TarneebWebController = GameWebController[usecase.TarneebInteractorIF, TarneebWebInput, *TarneebWebOutput]

// NewTarneebWebController and NewTarneebWebControllerWithProvider are
// the standard and provider-backed constructors for TarneebWebController.
var NewTarneebWebController, NewTarneebWebControllerWithProvider = webControllerPair[usecase.TarneebInteractorIF, TarneebWebInput, *TarneebWebOutput](
	newTarneebDefaultOutput, tarneebDispatch,
)

func newTarneebDefaultOutput(msg string) *TarneebWebOutput {
	return &TarneebWebOutput{
		Players:       make([]*TarneebWebOutputPlayer, 0),
		TeamScores:    make([]int, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		WinnerTeam:    -1,
		BidWinnerIdx:  -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func tarneebDispatch(bc *baseController, w http.ResponseWriter, ti usecase.TarneebInteractorIF, param TarneebWebInput, newDefault func(string) *TarneebWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ti.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		if !requireParam(bc, w, newDefault, param.Bid == nil, "param error: bid is required.") {
			return true
		}
		bc.writePresenterResponse(w, ti.Bid(*param.Bid))
	case "t", "trump":
		if !requireParam(bc, w, newDefault, param.TrumpSuit == nil, "param error: trumpSuit is required.") {
			return true
		}
		bc.writePresenterResponse(w, ti.DeclareTrump(*param.TrumpSuit))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, ti.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, ti.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, ti.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, ti.Hint, ti.ActionLog)
	}
	return true
}
