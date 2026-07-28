//go:build !js || !wasm || casino

package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CourtPieceWebInput Court Piece Web インプット
type CourtPieceWebInput struct {
	BaseWebInput
	TrumpSuit *int                 `json:"trumpSuit,omitempty"`
	CardIndex *int                 `json:"cardIndex,omitempty"`
	Config    *CourtPieceWebConfig `json:"config,omitempty"`
}

// CourtPieceWebConfig Court Piece Web 設定
type CourtPieceWebConfig struct {
	CpuDifficulty *int `json:"cpuDifficulty,omitempty"`
	PointLimit    *int `json:"pointLimit,omitempty"`
}

// CourtPieceWebOutputPlayer Court Piece Web アウトプットプレイヤー
type CourtPieceWebOutputPlayer struct {
	ID              int              `json:"id"`
	IsHuman         bool             `json:"isHuman"`
	Team            int              `json:"team"`
	CardCount       int              `json:"cardCount"`
	Cards           []*WebOutputCard `json:"cards"`
	RoundScore      int              `json:"roundScore"`
	CumulativeScore int              `json:"cumulativeScore"`
	TrickCount      int              `json:"trickCount"`
}

// CourtPieceWebOutputHint ヒント出力
type CourtPieceWebOutputHint struct {
	CardIndex *int   `json:"cardIndex,omitempty"`
	TrumpSuit *int   `json:"trumpSuit,omitempty"`
	Reason    string `json:"reason"`
}

// CourtPieceWebOutputConfig Court Piece 設定アウトプット
type CourtPieceWebOutputConfig struct {
	CpuDifficulty int `json:"cpuDifficulty"`
	PointLimit    int `json:"pointLimit"`
}

// CourtPieceWebOutput Court Piece Web アウトプット
type CourtPieceWebOutput struct {
	Players          []*CourtPieceWebOutputPlayer `json:"players"`
	TeamScores       []int                        `json:"teamScores"`
	Phase            int                          `json:"phase"`
	RoundNumber      int                          `json:"roundNumber"`
	TrickNumber      int                          `json:"trickNumber"`
	CurrentPlayerIdx int                          `json:"currentPlayerIdx"`
	CallerIdx        int                          `json:"callerIdx"`
	TrumpSuit        int                          `json:"trumpSuit"`
	CurrentTrick     []*WebOutputTrickCard        `json:"currentTrick"`
	ConsecutiveWins  int                          `json:"consecutiveWins"`
	LastWinnerTeam   int                          `json:"lastWinnerTeam"`
	LastRoundCourt   bool                         `json:"lastRoundCourt"`
	GameEndFlag      bool                         `json:"gameEndFlag"`
	WinnerTeam       int                          `json:"winnerTeam"`
	LeadPlayerIdx    int                          `json:"leadPlayerIdx"`
	Hint             *CourtPieceWebOutputHint     `json:"hint,omitempty"`
	WebOutputBase
	Config CourtPieceWebOutputConfig `json:"config"`
}

// ToConfig builds a CourtPieceConfig from the nested web config, applying bounds checking.
func (c *CourtPieceWebConfig) ToConfig() domain.CourtPieceConfig {
	cfg := domain.DefaultCourtPieceConfig()
	cfg.CpuDifficulty = domain.CourtPieceCpuDifficulty(webutil.BoundedIntPtr(c.CpuDifficulty, int(domain.CourtPieceCpuDifficultyEasy), int(domain.CourtPieceCpuDifficultyHard), int(cfg.CpuDifficulty)))
	webutil.ApplyBoundedInt(&cfg.PointLimit, c.PointLimit, 1, domain.CourtPieceMaxPointLimit)
	return cfg
}

// ToConfig builds a CourtPieceConfig from the web input.
func (p CourtPieceWebInput) ToConfig() domain.CourtPieceConfig {
	return configOrDefault(p.Config, (*CourtPieceWebConfig).ToConfig, domain.DefaultCourtPieceConfig())
}

// CourtPieceWebController Court Piece Web コントローラークラス
type CourtPieceWebController = GameWebController[usecase.CourtPieceInteractorIF, CourtPieceWebInput, *CourtPieceWebOutput]

// NewCourtPieceWebController and NewCourtPieceWebControllerWithProvider are
// the standard and provider-backed constructors for CourtPieceWebController.
var NewCourtPieceWebController, NewCourtPieceWebControllerWithProvider = webControllerPair[usecase.CourtPieceInteractorIF, CourtPieceWebInput, *CourtPieceWebOutput](
	newCourtPieceDefaultOutput, courtPieceDispatch,
)

func newCourtPieceDefaultOutput(msg string) *CourtPieceWebOutput {
	return &CourtPieceWebOutput{
		Players:       make([]*CourtPieceWebOutputPlayer, 0),
		TeamScores:    make([]int, 0),
		CurrentTrick:  make([]*WebOutputTrickCard, 0),
		WinnerTeam:    -1,
		WebOutputBase: WebOutputBase{Message: msg},
	}
}

func courtPieceDispatch(bc *baseController, w http.ResponseWriter, ti usecase.CourtPieceInteractorIF, param CourtPieceWebInput, newDefault func(string) *CourtPieceWebOutput) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, ti.ResetWithConfig(param.ToConfig()))
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
