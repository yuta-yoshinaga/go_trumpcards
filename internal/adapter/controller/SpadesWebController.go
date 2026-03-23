package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"github.com/ant0ine/go-json-rest/rest"
)

// SpadesWebInput スペードWebインプット
type SpadesWebInput struct {
	BaseWebInput
	Bid       *int             `json:"bid,omitempty"`
	CardIndex *int             `json:"cardIndex,omitempty"`
	Config    *SpadesWebConfig `json:"config,omitempty"`
}

// SpadesWebConfig スペードWeb設定
type SpadesWebConfig struct {
	CpuDifficulty       *int `json:"cpuDifficulty,omitempty"`
	PointLimit          *int `json:"pointLimit,omitempty"`
	NilBonus            *int `json:"nilBonus,omitempty"`
	BagPenaltyThreshold *int `json:"bagPenaltyThreshold,omitempty"`
}

// SpadesWebOutputPlayer スペードWebアウトプットプレイヤー
type SpadesWebOutputPlayer struct {
	ID              int              `json:"id"`
	IsHuman         bool             `json:"isHuman"`
	CardCount       int              `json:"cardCount"`
	Cards           []*WebOutputCard `json:"cards"`
	Bid             int              `json:"bid"`
	RoundScore      int              `json:"roundScore"`
	CumulativeScore int              `json:"cumulativeScore"`
	TrickCount      int              `json:"trickCount"`
	Bags            int              `json:"bags"`
}

// SpadesWebOutputTrickCard トリック中の1枚
type SpadesWebOutputTrickCard struct {
	PlayerIdx int            `json:"playerIdx"`
	Card      *WebOutputCard `json:"card"`
}

// SpadesWebOutputHint ヒント出力
type SpadesWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Bid       *int   `json:"bid,omitempty"`
	Reason    string `json:"reason"`
}

// SpadesWebOutput スペードWebアウトプット
type SpadesWebOutput struct {
	Players          []*SpadesWebOutputPlayer    `json:"players"`
	Phase            int                         `json:"phase"`
	RoundNumber      int                         `json:"roundNumber"`
	TrickNumber      int                         `json:"trickNumber"`
	CurrentPlayerIdx int                         `json:"currentPlayerIdx"`
	BidPlayerIdx     int                         `json:"bidPlayerIdx"`
	CurrentTrick     []*SpadesWebOutputTrickCard `json:"currentTrick"`
	SpadesBroken     bool                        `json:"spadesBroken"`
	GameEndFlag      bool                        `json:"gameEndFlag"`
	WinnerIdx        int                         `json:"winnerIdx"`
	LeadPlayerIdx    int                         `json:"leadPlayerIdx"`
	Hint             *SpadesWebOutputHint        `json:"hint,omitempty"`
	WebOutputBase
	Config SpadesWebOutputConfig `json:"config"`
}

// SpadesWebOutputConfig スペード設定アウトプット
type SpadesWebOutputConfig struct {
	CpuDifficulty       int `json:"cpuDifficulty"`
	PointLimit          int `json:"pointLimit"`
	NilBonus            int `json:"nilBonus"`
	BagPenaltyThreshold int `json:"bagPenaltyThreshold"`
}

// ToConfig builds a SpadesConfig from the nested web config, applying bounds checking.
func (c *SpadesWebConfig) ToConfig() domain.SpadesConfig {
	cfg := domain.DefaultSpadesConfig()
	cfg.CpuDifficulty = domain.SpadesCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.SpadesCpuDifficultyEasy), int(domain.SpadesCpuDifficultyHard), int(cfg.CpuDifficulty)))
	cfg.PointLimit = webutil.BoundedIntPtr(c.PointLimit, 1, 1000, cfg.PointLimit)
	cfg.NilBonus = webutil.BoundedIntPtr(c.NilBonus, 0, 500, cfg.NilBonus)
	cfg.BagPenaltyThreshold = webutil.BoundedIntPtr(c.BagPenaltyThreshold, 1, 100, cfg.BagPenaltyThreshold)
	return cfg
}

// ToConfig builds a SpadesConfig from the web input.
func (p SpadesWebInput) ToConfig() domain.SpadesConfig {
	return configOrDefault(p.Config, (*SpadesWebConfig).ToConfig, domain.DefaultSpadesConfig())
}

// SpadesWebController スペードWebコントローラークラス
type SpadesWebController = GameWebController[usecase.SpadesInteractorIF, SpadesWebInput, *SpadesWebOutput]

// NewSpadesWebController コンストラクタ
func NewSpadesWebController(factory func() usecase.SpadesInteractorIF) *SpadesWebController {
	return NewGameWebController(factory, newSpadesDefaultOutput, spadesDispatch)
}

func newSpadesDefaultOutput(msg string) *SpadesWebOutput {
	return &SpadesWebOutput{
		Players:       make([]*SpadesWebOutputPlayer, 0),
		CurrentTrick:  make([]*SpadesWebOutputTrickCard, 0),
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func spadesDispatch(bc *baseController, w rest.ResponseWriter, si usecase.SpadesInteractorIF, param SpadesWebInput, newDefault func(string) *SpadesWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, si.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		if param.Bid == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: bid is required."))
			return true
		}
		bc.writePresenterResponse(w, si.Bid(*param.Bid))
	case "p", "play":
		if param.CardIndex == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: cardIndex is required."))
			return true
		}
		bc.writePresenterResponse(w, si.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, si.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, si.NextRound())
	case "h", "hint":
		bc.writePresenterResponse(w, si.Hint())
	case "log", "l":
		bc.writePresenterResponse(w, si.ActionLog())
	default:
		return false
	}
	return true
}
