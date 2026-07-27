package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// WhistWebInput ホイストWebインプット
type WhistWebInput struct {
	BaseWebInput
	CardIndex *int            `json:"cardIndex,omitempty"`
	Config    *WhistWebConfig `json:"config,omitempty"`
}

// WhistWebConfig ホイストWeb設定
type WhistWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	PointLimit    *int `json:"pointLimit,omitempty"`
}

// WhistWebOutputPlayer ホイストWebアウトプットプレイヤー
type WhistWebOutputPlayer struct {
	ID              int              `json:"id"`
	IsHuman         bool             `json:"isHuman"`
	CardCount       int              `json:"cardCount"`
	Cards           []*WebOutputCard `json:"cards"`
	RoundScore      int              `json:"roundScore"`
	CumulativeScore int              `json:"cumulativeScore"`
	TrickCount      int              `json:"trickCount"`
	Team            int              `json:"team"`
}

// WhistWebOutputHint ヒント出力
type WhistWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	Reason    string `json:"reason"`
}

// WhistWebOutput ホイストWebアウトプット
type WhistWebOutput struct {
	Players          []*WhistWebOutputPlayer `json:"players"`
	Phase            int                     `json:"phase"`
	RoundNumber      int                     `json:"roundNumber"`
	TrickNumber      int                     `json:"trickNumber"`
	CurrentPlayerIdx int                     `json:"currentPlayerIdx"`
	CurrentTrick     []*WebOutputTrickCard   `json:"currentTrick"`
	TrumpSuit        int                     `json:"trumpSuit"`
	DealerIdx        int                     `json:"dealerIdx"`
	TeamScores       [2]int                  `json:"teamScores"`
	GameEndFlag      bool                    `json:"gameEndFlag"`
	WinnerTeam       int                     `json:"winnerTeam"`
	LeadPlayerIdx    int                     `json:"leadPlayerIdx"`
	Hint             *WhistWebOutputHint     `json:"hint,omitempty"`
	WebOutputBase
	Config WhistWebOutputConfig `json:"config"`
}

// WhistWebOutputConfig ホイスト設定アウトプット
type WhistWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	PointLimit    int `json:"pointLimit"`
}

// ToConfig builds a WhistConfig from the nested web config, applying bounds checking.
func (c *WhistWebConfig) ToConfig() domain.WhistConfig {
	cfg := domain.DefaultWhistConfig()
	cfg.CpuDifficulty = domain.WhistCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.WhistCpuDifficultyEasy), int(domain.WhistCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.PointLimit, c.PointLimit, 1, 1000)
	return cfg
}

// ToConfig builds a WhistConfig from the web input.
func (p WhistWebInput) ToConfig() domain.WhistConfig {
	return configOrDefault(p.Config, (*WhistWebConfig).ToConfig, domain.DefaultWhistConfig())
}

// WhistWebController ホイストWebコントローラークラス
type WhistWebController = GameWebController[usecase.WhistInteractorIF, WhistWebInput, *WhistWebOutput]

// NewWhistWebController and NewWhistWebControllerWithProvider are
// the standard and provider-backed constructors for WhistWebController.
var NewWhistWebController, NewWhistWebControllerWithProvider = webControllerPair[usecase.WhistInteractorIF, WhistWebInput, *WhistWebOutput](
	newWhistDefaultOutput, whistDispatch,
)

func newWhistDefaultOutput(msg string) *WhistWebOutput {
	return &WhistWebOutput{
		Players:       make([]*WhistWebOutputPlayer, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		WinnerTeam:    -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func whistDispatch(bc *baseController, w http.ResponseWriter, wi usecase.WhistInteractorIF, param WhistWebInput, newDefault func(string) *WhistWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, wi.ResetWithConfig(param.ToConfig()))
	case "p", "play":
		if !requireParam(bc, w, newDefault, param.CardIndex == nil, "param error: cardIndex is required.") {
			return true
		}
		bc.writePresenterResponse(w, wi.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, wi.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, wi.NextRound())
	default:
		return dispatchHintAndLog(param.Command, bc, w, wi.Hint, wi.ActionLog)
	}
	return true
}
