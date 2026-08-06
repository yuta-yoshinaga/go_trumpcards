package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
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

// SpadesWebOutputHint ヒント出力
type SpadesWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Bid       *int   `json:"bid,omitempty"`
	Reason    string `json:"reason"`
}

// SpadesWebOutput スペードWebアウトプット
type SpadesWebOutput struct {
	Players          []*SpadesWebOutputPlayer `json:"players"`
	Phase            int                      `json:"phase"`
	RoundNumber      int                      `json:"roundNumber"`
	TrickNumber      int                      `json:"trickNumber"`
	CurrentPlayerIdx int                      `json:"currentPlayerIdx"`
	BidPlayerIdx     int                      `json:"bidPlayerIdx"`
	CurrentTrick     []*WebOutputTrickCard    `json:"currentTrick"`
	SpadesBroken     bool                     `json:"spadesBroken"`
	GameEndFlag      bool                     `json:"gameEndFlag"`
	WinnerIdx        int                      `json:"winnerIdx"`
	LeadPlayerIdx    int                      `json:"leadPlayerIdx"`
	Hint             *SpadesWebOutputHint     `json:"hint,omitempty"`
	// ValidPlayIndices は人間がいま出せる手札の位置。フォロースートと
	// スペードブレイク前のリード制限はドメインが判定済みだが Web に載って
	// おらず、違反札をクリックしてエラーで確かめるしかなかった。
	// プレイフェーズで人間の手番のときだけ埋まり、それ以外は空。
	ValidPlayIndices []int `json:"validPlayIndices"`
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
	webutil.ApplyBoundedInt(&cfg.PointLimit, c.PointLimit, 1, 1000)
	webutil.ApplyBoundedInt(&cfg.NilBonus, c.NilBonus, 0, 500)
	webutil.ApplyBoundedInt(&cfg.BagPenaltyThreshold, c.BagPenaltyThreshold, 1, 100)
	return cfg
}

// ToConfig builds a SpadesConfig from the web input.
func (p SpadesWebInput) ToConfig() domain.SpadesConfig {
	return configOrDefault(p.Config, (*SpadesWebConfig).ToConfig, domain.DefaultSpadesConfig())
}

// SpadesWebController スペードWebコントローラークラス
type SpadesWebController = GameWebController[usecase.SpadesInteractorIF, SpadesWebInput, *SpadesWebOutput]

// NewSpadesWebController and NewSpadesWebControllerWithProvider are
// the standard and provider-backed constructors for SpadesWebController.
var NewSpadesWebController, NewSpadesWebControllerWithProvider = webControllerPair[usecase.SpadesInteractorIF, SpadesWebInput, *SpadesWebOutput](
	newSpadesDefaultOutput, spadesDispatch,
)

func newSpadesDefaultOutput(msg string) *SpadesWebOutput {
	return &SpadesWebOutput{
		Players:       make([]*SpadesWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		WinnerIdx:     -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func spadesDispatch(bc *baseController, w http.ResponseWriter, si usecase.SpadesInteractorIF, param SpadesWebInput, newDefault func(string) *SpadesWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, si.ResetWithConfig(param.ToConfig()))
	case "b", "bid":
		if !requireParam(bc, w, newDefault, param.Bid == nil, "param error: bid is required.") {
			return true
		}
		bc.writePresenterResponse(w, si.Bid(*param.Bid))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, si.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, si.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, si.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, si.Hint, si.ActionLog)
	}
	return true
}
