package controller

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/webutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BeggarMyNeighbourWebInput Beggar-My-Neighbour Webインプット
type BeggarMyNeighbourWebInput struct {
	BaseWebInput
	MaxRounds *int `json:"maxRounds,omitempty"`
}

// ToConfig インプットから BeggarMyNeighbourConfig を生成する
func (in BeggarMyNeighbourWebInput) ToConfig() domain.BeggarMyNeighbourConfig {
	cfg := domain.DefaultBeggarMyNeighbourConfig()
	if in.MaxRounds != nil {
		cfg.MaxRounds = webutil.BoundedIntPtr(in.MaxRounds, domain.BeggarMyNeighbourMinMaxRounds, domain.BeggarMyNeighbourMaxMaxRounds, cfg.MaxRounds)
	}
	return cfg
}

// BeggarMyNeighbourWebOutputPlayer Beggar-My-Neighbour Webアウトプットプレイヤー
type BeggarMyNeighbourWebOutputPlayer struct {
	ID              int  `json:"id"`
	IsHuman         bool `json:"isHuman"`
	DrawPileSize    int  `json:"drawPileSize"`
	DiscardPileSize int  `json:"discardPileSize"`
	TotalCards      int  `json:"totalCards"`
}

// BeggarMyNeighbourWebOutputConfig 設定情報
type BeggarMyNeighbourWebOutputConfig struct {
	MaxRounds int `json:"maxRounds"`
}

// BeggarMyNeighbourWebOutput Beggar-My-Neighbour Webアウトプット
type BeggarMyNeighbourWebOutput struct {
	Players          []*BeggarMyNeighbourWebOutputPlayer `json:"players"`
	Phase            int                                 `json:"phase"`
	GameEndFlag      bool                                `json:"gameEndFlag"`
	WinnerIdx        int                                 `json:"winnerIdx"`
	CurrentPlayerIdx int                                 `json:"currentPlayerIdx"`
	PenaltyOwnerIdx  int                                 `json:"penaltyOwnerIdx"`
	PenaltyRemaining int                                 `json:"penaltyRemaining"`
	CentralPileSize  int                                 `json:"centralPileSize"`
	LastCardPlayed   *WebOutputCard                      `json:"lastCardPlayed"`
	RoundsPlayed     int                                 `json:"roundsPlayed"`
	Config           BeggarMyNeighbourWebOutputConfig    `json:"config"`
	WebOutputBase
}

// BeggarMyNeighbourWebController Beggar-My-Neighbour Webコントローラー型
type BeggarMyNeighbourWebController = GameWebController[usecase.BeggarMyNeighbourInteractorIF, BeggarMyNeighbourWebInput, *BeggarMyNeighbourWebOutput]

// NewBeggarMyNeighbourWebController and NewBeggarMyNeighbourWebControllerWithProvider are
// the standard and provider-backed constructors for BeggarMyNeighbourWebController.
var NewBeggarMyNeighbourWebController, NewBeggarMyNeighbourWebControllerWithProvider = webControllerPair[usecase.BeggarMyNeighbourInteractorIF, BeggarMyNeighbourWebInput, *BeggarMyNeighbourWebOutput](
	newBeggarMyNeighbourDefaultOutput, beggarMyNeighbourDispatch,
)

func newBeggarMyNeighbourDefaultOutput(msg string) *BeggarMyNeighbourWebOutput {
	o := new(BeggarMyNeighbourWebOutput)
	o.Message = msg
	return o
}

func beggarMyNeighbourDispatch(
	bc *baseController,
	w http.ResponseWriter,
	bi usecase.BeggarMyNeighbourInteractorIF,
	param BeggarMyNeighbourWebInput,
	_ func(string) *BeggarMyNeighbourWebOutput,
) bool {
	switch param.Command {
	case "r", "reset":
		bc.writePresenterResponse(w, bi.ResetWithConfig(param.ToConfig()))
		return true
	case "a", "autoplay":
		bc.writePresenterResponse(w, bi.AutoPlay())
		return true
	}
	return dispatchResetStepLog(param.Command, bc, w, bi)
}
