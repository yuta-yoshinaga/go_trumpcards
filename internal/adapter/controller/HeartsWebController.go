package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"github.com/ant0ine/go-json-rest/rest"
)

// HeartsWebInput ハーツWebインプット
type HeartsWebInput struct {
	BaseWebInput
	CardIndices []int            `json:"cardIndices,omitempty"`
	CardIndex   *int             `json:"cardIndex,omitempty"`
	Config      *HeartsWebConfig `json:"config,omitempty"`
}

// HeartsWebConfig ハーツWeb設定
type HeartsWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	PointLimit    *int `json:"pointLimit,omitempty"`
}

// HeartsWebOutputPlayer ハーツWebアウトプットプレイヤー
type HeartsWebOutputPlayer struct {
	ID              int              `json:"id"`
	IsHuman         bool             `json:"isHuman"`
	CardCount       int              `json:"cardCount"`
	Cards           []*WebOutputCard `json:"cards"`
	RoundScore      int              `json:"roundScore"`
	CumulativeScore int              `json:"cumulativeScore"`
	TrickCount      int              `json:"trickCount"`
}

// HeartsWebOutputTrickCard トリック中の1枚
type HeartsWebOutputTrickCard struct {
	PlayerIdx int            `json:"playerIdx"`
	Card      *WebOutputCard `json:"card"`
}

// HeartsWebOutput ハーツWebアウトプット
type HeartsWebOutput struct {
	Players          []*HeartsWebOutputPlayer    `json:"players"`
	Phase            int                         `json:"phase"`
	RoundNumber      int                         `json:"roundNumber"`
	TrickNumber      int                         `json:"trickNumber"`
	CurrentPlayerIdx int                         `json:"currentPlayerIdx"`
	CurrentTrick     []*HeartsWebOutputTrickCard `json:"currentTrick"`
	HeartsBroken     bool                        `json:"heartsBroken"`
	PassDirection    int                         `json:"passDirection"`
	GameEndFlag      bool                        `json:"gameEndFlag"`
	WinnerIdx        int                         `json:"winnerIdx"`
	LeadPlayerIdx    int                         `json:"leadPlayerIdx"`
	Message          string                      `json:"message"`
	MessageCode      string                      `json:"messageCode,omitempty"`
	MessageParams    map[string]string           `json:"messageParams,omitempty"`
	Config           HeartsWebOutputConfig       `json:"config"`
}

// HeartsWebOutputConfig ハーツ設定アウトプット
type HeartsWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	PointLimit    int `json:"pointLimit"`
}

// HeartsWebController ハーツWebコントローラークラス
type HeartsWebController = GameWebController[usecase.HeartsInteractorIF, HeartsWebInput, *HeartsWebOutput]

// NewHeartsWebController コンストラクタ
func NewHeartsWebController(factory func() usecase.HeartsInteractorIF) *HeartsWebController {
	return NewGameWebController(factory, newHeartsDefaultOutput, heartsDispatch)
}

func newHeartsDefaultOutput(msg string) *HeartsWebOutput {
	return &HeartsWebOutput{
		Players:      make([]*HeartsWebOutputPlayer, 0),
		CurrentTrick: make([]*HeartsWebOutputTrickCard, 0),
		WinnerIdx:    -1,
		Message:      msg,
	}
}

func heartsDispatch(bc *baseController, w rest.ResponseWriter, hi usecase.HeartsInteractorIF, param HeartsWebInput, newDefault func(string) *HeartsWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		cfg := domain.DefaultHeartsConfig()
		if param.Config != nil {
			if param.Config.CpuDifficulty != nil {
				d := *param.Config.CpuDifficulty
				if d >= int(domain.HeartsCpuDifficultyEasy) && d <= int(domain.HeartsCpuDifficultyHard) {
					cfg.CpuDifficulty = domain.HeartsCpuDifficulty(d)
				}
			}
			if param.Config.PointLimit != nil && *param.Config.PointLimit >= 1 && *param.Config.PointLimit <= 1000 {
				cfg.PointLimit = *param.Config.PointLimit
			}
		}
		bc.writePresenterResponse(w, hi.ResetWithConfig(cfg))
	case "pass":
		if len(param.CardIndices) != 3 {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: pass requires exactly 3 card indices."))
			return true
		}
		bc.writePresenterResponse(w, hi.Pass(param.CardIndices))
	case "p", "play":
		if param.CardIndex == nil {
			bc.writeJsonResponse(w, http.StatusBadRequest, newDefault("param error: cardIndex is required."))
			return true
		}
		bc.writePresenterResponse(w, hi.Play(*param.CardIndex))
	case "n", "next":
		bc.writePresenterResponse(w, hi.NextTrick())
	case "nr", "nextround":
		bc.writePresenterResponse(w, hi.NextRound())
	case "log", "l":
		bc.writePresenterResponse(w, hi.ActionLog())
	default:
		return false
	}
	return true
}
