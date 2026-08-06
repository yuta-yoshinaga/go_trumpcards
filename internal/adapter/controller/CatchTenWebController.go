package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CatchTenWebInput Catch the Ten Webインプット
type CatchTenWebInput struct {
	BaseWebInput
	CardIndex *int               `json:"cardIndex,omitempty"`
	Config    *CatchTenWebConfig `json:"config,omitempty"`
}

// CatchTenWebConfig Catch the Ten Web設定
type CatchTenWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	PointLimit    *int `json:"pointLimit,omitempty"`
}

// CatchTenWebOutputPlayer Catch the Ten Webアウトプットプレイヤー
type CatchTenWebOutputPlayer struct {
	ID              int              `json:"id"`
	IsHuman         bool             `json:"isHuman"`
	CardCount       int              `json:"cardCount"`
	Cards           []*WebOutputCard `json:"cards"`
	RoundScore      int              `json:"roundScore"`
	CumulativeScore int              `json:"cumulativeScore"`
	TrickCount      int              `json:"trickCount"`
	Team            int              `json:"team"`
}

// CatchTenWebOutputHint ヒント出力
type CatchTenWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Reason    string `json:"reason"`
}

// CatchTenWebOutput Catch the Ten Webアウトプット
type CatchTenWebOutput struct {
	Players          []*CatchTenWebOutputPlayer `json:"players"`
	Phase            int                        `json:"phase"`
	RoundNumber      int                        `json:"roundNumber"`
	TrickNumber      int                        `json:"trickNumber"`
	CurrentPlayerIdx int                        `json:"currentPlayerIdx"`
	CurrentTrick     []*WebOutputTrickCard      `json:"currentTrick"`
	TrumpSuit        int                        `json:"trumpSuit"`
	DealerIdx        int                        `json:"dealerIdx"`
	TeamScores       [2]int                     `json:"teamScores"`
	GameEndFlag      bool                       `json:"gameEndFlag"`
	WinnerTeam       int                        `json:"winnerTeam"`
	LeadPlayerIdx    int                        `json:"leadPlayerIdx"`
	Hint             *CatchTenWebOutputHint     `json:"hint,omitempty"`
	// ValidPlayIndices は人間がいま出せる手札の位置。ドメインの
	// GetValidPlayIndices はコメントに「Web用」と書かれているのに一度も
	// 呼ばれておらず、違反札をクリックしてエラーで確かめるしかなかった。
	// プレイフェーズで人間の手番のときだけ埋まり、それ以外は空。
	ValidPlayIndices []int `json:"validPlayIndices"`
	WebOutputBase
	Config CatchTenWebOutputConfig `json:"config"`
}

// CatchTenWebOutputConfig Catch the Ten 設定アウトプット
type CatchTenWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	PointLimit    int `json:"pointLimit"`
}

// ToConfig builds a CatchTenConfig from the nested web config, applying bounds checking.
func (c *CatchTenWebConfig) ToConfig() domain.CatchTenConfig {
	cfg := domain.DefaultCatchTenConfig()
	cfg.CpuDifficulty = domain.CatchTenCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.CatchTenCpuDifficultyEasy), int(domain.CatchTenCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.PointLimit, c.PointLimit, 1, 1000)
	return cfg
}

// ToConfig builds a CatchTenConfig from the web input.
func (p CatchTenWebInput) ToConfig() domain.CatchTenConfig {
	return configOrDefault(p.Config, (*CatchTenWebConfig).ToConfig, domain.DefaultCatchTenConfig())
}

// CatchTenWebController Catch the Ten Webコントローラークラス
type CatchTenWebController = GameWebController[usecase.CatchTenInteractorIF, CatchTenWebInput, *CatchTenWebOutput]

// NewCatchTenWebController and NewCatchTenWebControllerWithProvider are
// the standard and provider-backed constructors for CatchTenWebController.
var NewCatchTenWebController, NewCatchTenWebControllerWithProvider = webControllerPair[usecase.CatchTenInteractorIF, CatchTenWebInput, *CatchTenWebOutput](
	newCatchTenDefaultOutput, catchTenDispatch,
)

func newCatchTenDefaultOutput(msg string) *CatchTenWebOutput {
	return &CatchTenWebOutput{
		Players:       make([]*CatchTenWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		WinnerTeam:    -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func catchTenDispatch(bc *baseController, w http.ResponseWriter, ci usecase.CatchTenInteractorIF, param CatchTenWebInput, newDefault func(string) *CatchTenWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ci.ResetWithConfig(param.ToConfig()))
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
